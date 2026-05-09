package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Khudo-R/sanguis/internal/limiter"
	"github.com/redis/go-redis/v9"
)

func TestLimiters(t *testing.T) {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	redisAvailable := true
	if err := rdb.Ping(ctx).Err(); err != nil {
		redisAvailable = false
	}

	testCases := []struct {
		name    string
		limiter limiter.Limiter
		skip    bool
	}{
		{
			name:    "InMemory",
			limiter: limiter.NewInMemoryLimiter(),
		},
		{
			name:    "SlidingWindow",
			limiter: limiter.NewSlidingWindowLimiter(),
		},
		{
			name:    "TokenBucket",
			limiter: limiter.NewTokenBucketLimiter(),
		},
		{
			name:    "Redis",
			limiter: limiter.NewRedisLimiter(rdb),
			skip:    !redisAvailable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip("Skipping limiter test because dependency (Redis) is not available")
				return
			}

			key := fmt.Sprintf("test-key-%s-%d", tc.name, time.Now().UnixNano())
			limit := 5
			window := 1 * time.Second

			for i := 0; i < limit; i++ {
				res, err := tc.limiter.Check(ctx, key, limit, window)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !res.Allowed {
					t.Errorf("expected request %d to be allowed", i+1)
				}
			}

			res, err := tc.limiter.Check(ctx, key, limit, window)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.Allowed {
				t.Error("expected request to be blocked")
			}

			time.Sleep(window + 100*time.Millisecond)

			res, err = tc.limiter.Check(ctx, key, limit, window)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !res.Allowed {
				t.Error("expected request to be allowed after window reset/refill")
			}
		})
	}
}
