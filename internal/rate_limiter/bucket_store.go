package rate_limiter

import (
	"ratelimiter/internal/models"
	"sync"
	"time"
)

type BucketStore struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
}

func NewBucketStore() *BucketStore {
	return &BucketStore{
		buckets: make(map[string]*TokenBucket),
	}
}

func (s *BucketStore) GetOrCreate(key string, limit models.Limit) *TokenBucket {
	s.mu.RLock()
	b, exists := s.buckets[key]
	s.mu.RUnlock()

	if exists {
		return b
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if b, exists = s.buckets[key]; exists {
		return b
	}

	tb := NewTokenBucket(limit.Capacity, time.Second*time.Duration(limit.RefillRate), false)
	s.buckets[key] = tb
	return tb
}

func (s *BucketStore) Get(key string) *TokenBucket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buckets[key]
}

func (s *BucketStore) Set(key string, bucket *TokenBucket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buckets[key] = bucket
}

func (s *BucketStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.buckets, key)
}

func (s *BucketStore) StartBackgroundRefill(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.RLock()
		for _, bucket := range s.buckets {
			if !bucket.unlimited && interval <= bucket.refillRate {
				bucket.Refill()
			}
		}
		s.mu.RUnlock()
	}
}

func (tb *TokenBucket) Refill() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	newTokens := int64(elapsed / tb.refillRate)

	if newTokens > 0 {
		tb.tokens = min(tb.capacity, tb.tokens+newTokens)
		tb.lastRefill = now
	}
}
