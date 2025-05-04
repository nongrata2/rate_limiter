package rate_limiter

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   int64
	tokens     int64
	refillRate time.Duration
	lastRefill time.Time
	unlimited  bool
	mu         sync.Mutex
}

func NewTokenBucket(capacity int64, refillRate time.Duration, unlimited bool) *TokenBucket {
	if refillRate <= 0 {
		refillRate = time.Second
	}

	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
		unlimited:  unlimited,
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.unlimited {
		return true
	}

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	newTokens := int64(elapsed / tb.refillRate)

	if newTokens > 0 {
		tb.tokens = min(tb.capacity, tb.tokens+newTokens)
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}
