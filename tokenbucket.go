package ratelimit

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
)

var ErrTooManyRequests = errors.New("Too Many Requests")

var bucketContainer = newTokenBucketContainer()

func Wrap(h http.Handler, bucketSize int, delay time.Duration) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		handler:    h,
		bucketSize: bucketSize,
		delay:      delay,
	}
}

type RateLimiterMiddleware struct {
	handler    http.Handler
	bucketSize int
	delay      time.Duration
}

func (rl *RateLimiterMiddleware) fail(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")

	w.WriteHeader(http.StatusTooManyRequests)

	r, _ := json.Marshal(map[string]string{
		"message": ErrTooManyRequests.Error(),
	})

	_, err := w.Write(r)
	if err != nil {
		panic(err)
	}
}

func (rl *RateLimiterMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !bucketContainer.HasBucket(r.RemoteAddr) {
		bucketContainer.NewTokenBucket(r.RemoteAddr, rl.bucketSize, rl.delay)
	}

	if err := bucketContainer.Consume(r.RemoteAddr); err != nil {
		if err == ErrTooManyRequests {
			rl.fail(w)
			return
		}
		log.Fatal(err)
	}

	rl.handler.ServeHTTP(w, r)
}

func newTokenBucketContainer() *tokenBucketContainer {
	container := &tokenBucketContainer{
		buckets: make(map[string]*tokenBucket),
		mux:     sync.RWMutex{},
	}

	return container
}

type tokenBucketContainer struct {
	buckets map[string]*tokenBucket
	mux     sync.RWMutex
}

func (c *tokenBucketContainer) deleteBucket(ip string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	delete(c.buckets, ip)
}

func (c *tokenBucketContainer) HasBucket(ip string) bool {
	c.mux.RLock()
	defer c.mux.RUnlock()

	_, ok := c.buckets[ip]

	return ok
}

func (c *tokenBucketContainer) Consume(ip string) error {
	c.mux.RLock()
	bucket, ok := c.buckets[ip]
	c.mux.RUnlock()

	if ok {
		if err := bucket.consume(); err != nil {
			return err
		}
	}

	return nil
}

func (c *tokenBucketContainer) NewTokenBucket(ip string, bucketSize int, destroyDelay time.Duration) *tokenBucket {
	c.mux.Lock()
	defer c.mux.Unlock()

	t := &tokenBucket{
		name:         ip,
		bucketSize:   bucketSize,
		destroyDelay: destroyDelay,
		bucket:       make([]struct{}, bucketSize),
	}

	c.buckets[ip] = t

	ch := t.getDestroySignal()
	go func() {
		c.deleteBucket(<-ch)
	}()

	return t
}

type tokenBucket struct {
	name         string
	bucketSize   int
	bucket       []struct{}
	destroyDelay time.Duration
	mux          sync.Mutex
}

func (t *tokenBucket) consume() error {
	t.mux.Lock()
	defer t.mux.Unlock()

	if len(t.bucket) > 0 && len(t.bucket) <= t.bucketSize {
		t.bucket = t.bucket[:len(t.bucket)-1]
		return nil
	}

	return ErrTooManyRequests
}

func (t *tokenBucket) getDestroySignal() <-chan string {
	ticker := time.NewTicker(t.destroyDelay)

	ch := make(chan string)
	go func() {
		<-ticker.C

		ticker.Stop()

		ch <- t.name

		close(ch)
	}()

	return ch
}
