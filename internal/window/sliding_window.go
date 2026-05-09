package window

import (
	"sync"
	"time"
)

type SlidingWindowCounter struct {
	prevCount int
	currCount int
	currStart time.Time
	mu        sync.Mutex
}

func NewSlidingWindowCounter() *SlidingWindowCounter {
	return &SlidingWindowCounter{}
}

func (c *SlidingWindowCounter) Allow(limit int, window time.Duration) (bool, int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	currentWindowStart := now.Truncate(window)

	if c.currStart.IsZero() {
		c.currStart = currentWindowStart
		c.currCount = 0
		c.prevCount = 0
	} else if c.currStart != currentWindowStart {
		if c.currStart == currentWindowStart.Add(-window) {
			c.prevCount = c.currCount
		} else {
			c.prevCount = 0
		}
		c.currCount = 0
		c.currStart = currentWindowStart
	}

	elapsed := now.Sub(currentWindowStart)
	overlapPercentage := float64(window-elapsed) / float64(window)
	estimatedCount := float64(c.prevCount)*overlapPercentage + float64(c.currCount)

	if estimatedCount < float64(limit) {
		c.currCount++
		return true, limit - int(estimatedCount) - 1
	}

	return false, 0
}
