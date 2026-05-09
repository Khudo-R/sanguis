package bucket

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   int
	tokens     float64
	rate       float64
	lastRefill time.Time
	mutex      sync.Mutex
}

func NewTokenBucket(capacity int, rate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		rate:       rate,
		tokens:     float64(capacity),
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	tokensToAdd := elapsed.Seconds() * tb.rate

	if tokensToAdd > 0 {
		tb.lastRefill = now
		tb.tokens = minFloat(tb.tokens+tokensToAdd, float64(tb.capacity))
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func (tb *TokenBucket) Take(tokens int) (bool, int) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	tb.refill()

	if tb.tokens >= float64(tokens) {
		tb.tokens -= float64(tokens)
		return true, int(tb.tokens)
	}
	return false, int(tb.tokens)
}

func (tb *TokenBucket) Tokens() int {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	tb.refill()
	return int(tb.tokens)
}
