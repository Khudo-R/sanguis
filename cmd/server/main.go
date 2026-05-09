package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "github.com/Khudo-R/sanguis/api/gen/v1"
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
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	l := limiter.NewRedisLimiter(rdb)
	s := grpc.NewServer()

	pb.RegisterLimiterServer(s, &server{limiter: l})
	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
