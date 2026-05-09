package limiter

import (
	"context"
	"time"

	"github.com/Khudo-R/sanguis/internal/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func UnaryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		var allowed bool
		if res, ok := resp.(interface{ GetAllowed() bool }); ok {
			allowed = res.GetAllowed()
		}

		resultLabel := "allowed"
		if !allowed {
			resultLabel = "blocked"
		}
		metrics.RequestsTotal.WithLabelValues(resultLabel).Inc()

		logger.Info("gRPC Request",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.Bool("allowed", allowed),
			zap.Error(err),
		)

		return resp, err
	}
}
