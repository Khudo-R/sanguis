package limiter

import (
	"context"
	"sync"
	"time"
)

type entry struct {
	count     int
	expiresAt time.Time
}

type InMemoryLimiter struct {
	mu     sync.Mutex
	counts map[string]*entry
}

func NewInMemoryLimiter() *InMemoryLimiter {
	return &InMemoryLimiter{
		counts: make(map[string]*entry),
	}
}

func (l *InMemoryLimiter) Check(ctx context.Context, key string, limit int, window time.Duration) (Result, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	e, exists := l.counts[key]

	if !exists || now.After(e.expiresAt) {
		e = &entry{
			count:     0,
			expiresAt: now.Add(window),
		}
		l.counts[key] = e
	}

	if e.count >= limit {
		return Result{
			Allowed:   false,
			Remaining: 0,
			ResetTime: e.expiresAt,
		}, nil
	}

	e.count++

	return Result{
		Allowed:   true,
		Remaining: limit - e.count,
		ResetTime: e.expiresAt,
	}, nil
}
