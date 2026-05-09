package limiter

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type HybridLimiter struct {
	l1           *InMemoryLimiter
	l2           Limiter
	syncInterval time.Duration
	logger       *zap.Logger
	mu           sync.Mutex
	pending      map[string]int
}

func NewHybridLimiter(l2 Limiter, syncInterval time.Duration, logger *zap.Logger) *HybridLimiter {
	h := &HybridLimiter{
		l1:           NewInMemoryLimiter(),
		l2:           l2,
		syncInterval: syncInterval,
		logger:       logger,
		pending:      make(map[string]int),
	}
	go h.syncLoop()
	return h
}

func (h *HybridLimiter) Check(ctx context.Context, key string, limit int, window time.Duration) (Result, error) {
	res, err := h.l1.Check(ctx, key, limit, window)
	if err != nil {
		return Result{}, err
	}

	if res.Allowed {
		h.mu.Lock()
		h.pending[key]++
		h.mu.Unlock()
	}

	return res, nil
}

func (h *HybridLimiter) syncLoop() {
	ticker := time.NewTicker(h.syncInterval)
	defer ticker.Stop()

	for range ticker.C {
		h.sync()
	}
}

func (h *HybridLimiter) sync() {
	h.mu.Lock()
	if len(h.pending) == 0 {
		h.mu.Unlock()
		return
	}
	batch := h.pending
	h.pending = make(map[string]int)
	h.mu.Unlock()

	ctx := context.Background()
	for key, count := range batch {
		for i := 0; i < count; i++ {
			// In a real production scenario, we'd use a more efficient batch increment Lua script
			_, _ = h.l2.Check(ctx, key, 0, 0)
		}
	}
}
