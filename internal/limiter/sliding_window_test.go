package limiter

import (
	"context"
	"testing"
	"time"
)

func TestSlidingWindowLimiter_Check(t *testing.T) {
	l := NewSlidingWindowLimiter()
	ctx := context.Background()
	key := "test-key"
	limit := 5
	window := time.Second

	// 1. Consume tokens in the first half of the window
	for i := 0; i < limit; i++ {
		res, err := l.Check(ctx, key, limit, window)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.Allowed {
			t.Errorf("expected allowed=true at iteration %d", i)
		}
	}

	// 2. Next one should be blocked
	res, err := l.Check(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Allowed {
		t.Error("expected allowed=false when limit is reached")
	}

	// 3. Wait for the next window boundary.
	// We want to test the "sliding" aspect.
	// If we wait for a full window, it should be completely reset.
	time.Sleep(window + 100*time.Millisecond)

	res, err = l.Check(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Allowed {
		t.Error("expected allowed=true after window transition")
	}
}

func TestSlidingWindowLimiter_Isolation(t *testing.T) {
	l := NewSlidingWindowLimiter()
	ctx := context.Background()
	limit := 2
	window := time.Hour

	l.Check(ctx, "key1", limit, window)
	l.Check(ctx, "key1", limit, window)
	res1, _ := l.Check(ctx, "key1", limit, window)
	if res1.Allowed {
		t.Error("key1 should be exhausted")
	}

	res2, _ := l.Check(ctx, "key2", limit, window)
	if !res2.Allowed {
		t.Error("key2 should not be exhausted")
	}
}
