package limiter

import (
	"context"
	"sync"
	"time"

	"github.com/Khudo-R/sanguis/internal/bucket"
)

type TokenBucketLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket.TokenBucket
}

func NewTokenBucketLimiter() *TokenBucketLimiter {
	return &TokenBucketLimiter{
		buckets: make(map[string]*bucket.TokenBucket),
	}
}

func (l *TokenBucketLimiter) Check(ctx context.Context, key string, limit int, window time.Duration) (Result, error) {
	l.mu.Lock()
	tb, exists := l.buckets[key]
	if !exists {
		rate := float64(limit) / window.Seconds()
		tb = bucket.NewTokenBucket(limit, rate)
		l.buckets[key] = tb
	}
	l.mu.Unlock()

	allowed, remaining := tb.Take(1)

	return Result{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: time.Now().Add(window),
	}, nil
}
