package limiter

import (
	"context"
	"sync"
	"time"
)

type InMemoryLimiter struct {
	mu     sync.Mutex
	counts map[string]int
}

func NewInMemoryLimiter() *InMemoryLimiter {
	return &InMemoryLimiter{
		counts: make(map[string]int),
	}
}

func (l *InMemoryLimiter) Check(ctx context.Context, key string, limit int, window time.Duration) (Result, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	current := l.counts[key]

	if current >= limit {
		return Result{
			Allowed:   false,
			Remaining: 0,
			ResetTime: time.Now().Add(window),
		}, nil
	}

	l.counts[key]++

	return Result{
		Allowed:   true,
		Remaining: limit - current - 1,
		ResetTime: time.Now().Add(window),
	}, nil
}
