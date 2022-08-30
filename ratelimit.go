package ratelimit

import (
	"context"
	"errors"
	"time"
)

var Factory = NewTokenBucketStrategyFactory(60, time.Minute)
var Container RateLimiter = NewInMemoryContainer()

var ErrTooManyRequests = errors.New("Too Many Requests")

type IP string

type RateLimiterStrategyFactory interface {
	Make(ip IP) RateLimiterStrategy
}

type RateLimiterStrategy interface {
	Consume() error
	GetDestroySignal() <-chan IP
	Marshal() map[string]interface{}
	Unmarshal(state map[string]interface{})
}

type RateLimiter interface {
	delete(ctx context.Context, ip IP) error
	Has(ctx context.Context, ip IP) bool
	Consume(ctx context.Context, ip IP) error
	New(ctx context.Context, ip IP, strategy RateLimiterStrategy) error
}
