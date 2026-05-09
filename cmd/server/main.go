package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/Khudo-R/sanguis/api/gen/v1"
	"github.com/Khudo-R/sanguis/configs"
	"github.com/Khudo-R/sanguis/internal/limiter"
	"github.com/redis/go-redis/v9"
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

	log.Printf("Received check request for key: %s", req.Key)

	return &pb.CheckResponse{
		Allowed:   res.Allowed,
		Remaining: int32(res.Remaining),
		ResetTime: res.ResetTime.Unix(),
	}, nil
}

func main() {
	cfg := configs.MustLoad("configs/config.yaml")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var l limiter.Limiter
	switch cfg.Limiter.Type {
	case "redis":
		rdb := redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Address,
			Password: cfg.Redis.Password,
		})
		l = limiter.NewRedisLimiter(rdb)
	case "sliding_window":
		l = limiter.NewSlidingWindowLimiter()
	case "token_bucket":
		l = limiter.NewTokenBucketLimiter()
	default:
		log.Printf("Unknown limiter type %s, falling back to memory", cfg.Limiter.Type)
		l = limiter.NewInMemoryLimiter()
	}

	s := grpc.NewServer()
	pb.RegisterLimiterServer(s, &server{limiter: l})

	log.Printf("server listening at %v with %s limiter", lis.Addr(), cfg.Limiter.Type)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
