package limiter

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucketLimiter_Check(t *testing.T) {
	l := NewTokenBucketLimiter()
	ctx := context.Background()
	key := "test-key"
	limit := 5
	window := time.Second

	// Consuming all tokens
	for i := 0; i < limit; i++ {
		res, err := l.Check(ctx, key, limit, window)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.Allowed {
			t.Errorf("expected allowed=true at iteration %d", i)
		}
		if res.Remaining != limit-i-1 {
			t.Errorf("expected remaining=%d, got %d", limit-i-1, res.Remaining)
		}
	}

	// Next one should be blocked
	res, err := l.Check(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Allowed {
		t.Error("expected allowed=false when bucket is empty")
	}

	// Wait for refill (at least 1/5 of a second)
	time.Sleep(250 * time.Millisecond)

	res, err = l.Check(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Allowed {
		t.Error("expected allowed=true after refill")
	}
}

func TestTokenBucketLimiter_Isolation(t *testing.T) {
	l := NewTokenBucketLimiter()
	ctx := context.Background()
	limit := 2
	window := time.Hour

	// Exhaust key1
	l.Check(ctx, "key1", limit, window)
	l.Check(ctx, "key1", limit, window)
	res1, _ := l.Check(ctx, "key1", limit, window)
	if res1.Allowed {
		t.Error("key1 should be exhausted")
	}

	// key2 should still be full
	res2, _ := l.Check(ctx, "key2", limit, window)
	if !res2.Allowed {
		t.Error("key2 should not be exhausted")
	}
	if res2.Remaining != 1 {
		t.Errorf("key2 remaining should be 1, got %d", res2.Remaining)
	}
}
