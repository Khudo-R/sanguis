package limiter

import (
	"context"
	"sync"
	"time"

	"github.com/Khudo-R/sanguis/internal/window"
)

type SlidingWindowLimiter struct {
	mu      sync.Mutex
	windows map[string]*window.SlidingWindowCounter
}

func NewSlidingWindowLimiter() *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		windows: make(map[string]*window.SlidingWindowCounter),
	}
}

func (l *SlidingWindowLimiter) Check(ctx context.Context, key string, limit int, windowDuration time.Duration) (Result, error) {
	l.mu.Lock()
	w, exists := l.windows[key]
	if !exists {
		w = window.NewSlidingWindowCounter()
		l.windows[key] = w
	}
	l.mu.Unlock()

	allowed, remaining := w.Allow(limit, windowDuration)

	return Result{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: time.Now().Add(windowDuration),
	}, nil
}
