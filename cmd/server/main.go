package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/Khudo-R/sanguis/api/gen/v1"
	"github.com/Khudo-R/sanguis/configs"
	"github.com/Khudo-R/sanguis/internal/limiter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedLimiterServer
	limiter limiter.Limiter
}

func (s *server) Check(ctx context.Context, req *pb.CheckRequest) (*pb.CheckResponse, error) {
	window := time.Duration(req.WindowWs) * time.Millisecond
	res, err := s.limiter.Check(ctx, req.Key, int(req.Limit), window)
	if err != nil {
		return nil, err
	}

	return &pb.CheckResponse{
		Allowed:   res.Allowed,
		Remaining: int32(res.Remaining),
		ResetTime: res.ResetTime.Unix(),
	}, nil
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	cfg := configs.MustLoad("configs/config.yaml")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
	})

	var l limiter.Limiter
	switch cfg.Limiter.Type {
	case "redis":
		l = limiter.NewRedisLimiter(rdb)
	case "hybrid":
		rl := limiter.NewRedisLimiter(rdb)
		l = limiter.NewHybridLimiter(rl, cfg.Hybrid.SyncInterval, logger)
	case "sliding_window":
		l = limiter.NewSlidingWindowLimiter()
	case "token_bucket":
		l = limiter.NewTokenBucketLimiter()
	default:
		l = limiter.NewInMemoryLimiter()
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(limiter.UnaryInterceptor(logger)),
	)
	pb.RegisterLimiterServer(s, &server{limiter: l})

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		logger.Info("metrics server starting", zap.Int("port", cfg.Metrics.Port))
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Metrics.Port), mux); err != nil {
			logger.Error("metrics server failed", zap.Error(err))
		}
	}()

	go func() {
		logger.Info("gRPC server starting", zap.Int("port", cfg.Server.Port), zap.String("limiter", cfg.Limiter.Type))
		if err := s.Serve(lis); err != nil {
			logger.Fatal("gRPC server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down servers...")
	s.GracefulStop()
	rdb.Close()
	logger.Info("shutdown complete")
}
