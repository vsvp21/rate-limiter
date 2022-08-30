package ratelimit

import (
	"sync"
	"time"
)

func NewTokenBucketStrategyFactory(bucketSize int, destroyDelay time.Duration) *TokenBucketStrategyFactory {
	return &TokenBucketStrategyFactory{
		bucketSize:   bucketSize,
		destroyDelay: destroyDelay,
	}
}

type TokenBucketStrategyFactory struct {
	bucketSize   int
	destroyDelay time.Duration
}

func (f TokenBucketStrategyFactory) Make(ip IP) RateLimiterStrategy {
	return newTokenBucket(ip, f.bucketSize, f.destroyDelay)
}

func newTokenBucket(ip IP, bucketSize int, destroyDelay time.Duration) *tokenBucket {
	return &tokenBucket{
		ip:           ip,
		bucketSize:   bucketSize,
		bucket:       make([]struct{}, bucketSize),
		destroyDelay: destroyDelay,
		mux:          sync.Mutex{},
	}
}

type tokenBucket struct {
	ip           IP
	bucketSize   int
	bucket       []struct{}
	destroyDelay time.Duration
	mux          sync.Mutex
}

func (t *tokenBucket) Consume() error {
	t.mux.Lock()
	defer t.mux.Unlock()

	if len(t.bucket) > 0 && len(t.bucket) <= t.bucketSize {
		t.bucket = t.bucket[:len(t.bucket)-1]
		return nil
	}

	return ErrTooManyRequests
}

func (t *tokenBucket) GetDestroySignal() <-chan IP {
	ticker := time.NewTicker(t.destroyDelay)

	ch := make(chan IP)
	go func() {
		<-ticker.C

		ticker.Stop()

		ch <- t.ip

		close(ch)
	}()

	return ch
}

func (t *tokenBucket) Marshal() map[string]interface{} {
	return map[string]interface{}{
		"ip":           t.ip,
		"bucketSize":   t.bucketSize,
		"destroyDelay": t.destroyDelay,
		"bucketLen":    len(t.bucket),
	}
}

func (t *tokenBucket) Unmarshal(state map[string]interface{}) {
	t.ip = IP(state["ip"].(string))
	t.bucketSize = int(state["bucketSize"].(int8))
	t.destroyDelay = time.Duration(state["destroyDelay"].(int64))
	t.bucket = make([]struct{}, state["bucketLen"].(int8), t.bucketSize)
}
