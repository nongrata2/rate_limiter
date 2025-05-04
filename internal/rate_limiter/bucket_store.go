package rate_limiter

import "sync"

type BucketStore struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
}

func NewBucketStore() *BucketStore {
	return &BucketStore{
		buckets: make(map[string]*TokenBucket),
	}
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
