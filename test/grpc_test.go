package test

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/Khudo-R/sanguis/api/gen/v1"
	"github.com/Khudo-R/sanguis/internal/limiter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type testServer struct {
	pb.UnimplementedLimiterServer
	limiter limiter.Limiter
}

func (s *testServer) Check(ctx context.Context, req *pb.CheckRequest) (*pb.CheckResponse, error) {
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

func TestGRPCServer(t *testing.T) {
	ctx := context.Background()
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	
	l := limiter.NewInMemoryLimiter()
	pb.RegisterLimiterServer(s, &testServer{limiter: l})
	
	go func() {
		if err := s.Serve(lis); err != nil {
		}
	}()
	defer s.Stop()

	conn, err := grpc.DialContext(ctx, "bufnet", 
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), 
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	
	client := pb.NewLimiterClient(conn)

	req := &pb.CheckRequest{
		Key:      "grpc-test-key",
		Limit:    2,
		WindowWs: 1000,
	}

	res, err := client.Check(ctx, req)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if !res.Allowed {
		t.Error("Expected first request to be allowed")
	}

	res, err = client.Check(ctx, req)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if !res.Allowed {
		t.Error("Expected second request to be allowed")
	}

	res, err = client.Check(ctx, req)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if res.Allowed {
		t.Error("Expected third request to be blocked")
	}
}
