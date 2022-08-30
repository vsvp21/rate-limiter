package ratelimit

import (
	"context"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"time"
)

func NewRedisRateLimiterContainer(host, port, pass string) *redisRateLimiterContainer {
	rdb := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: pass,
	})

	c := cache.New(&cache.Options{
		Redis: rdb,
	})

	return &redisRateLimiterContainer{
		cache: c,
	}
}

type redisRateLimiterContainer struct {
	cache *cache.Cache
}

func (l *redisRateLimiterContainer) delete(ctx context.Context, ip IP) error {
	return l.cache.Delete(ctx, string(ip))
}

func (l *redisRateLimiterContainer) Has(ctx context.Context, ip IP) bool {
	return l.cache.Exists(ctx, string(ip))
}

func (l *redisRateLimiterContainer) Consume(ctx context.Context, ip IP) error {
	var data map[string]interface{}

	err := l.cache.Get(ctx, string(ip), &data)
	if err != nil {
		return err
	}

	item := Factory.Make(ip)
	item.Unmarshal(data)

	if err = item.Consume(); err != nil {
		return err
	}

	if err = l.New(ctx, ip, item); err != nil {
		return err
	}

	return nil
}

func (l *redisRateLimiterContainer) New(ctx context.Context, ip IP, strategy RateLimiterStrategy) error {
	err := l.cache.Set(&cache.Item{
		Key:   string(ip),
		Value: strategy.Marshal(),
		TTL:   time.Hour,
	})
	if err != nil {
		return err
	}

	ch := strategy.GetDestroySignal()
	go func() {
		err = l.delete(context.Background(), <-ch)
		if err != nil {
			panic(err)
		}
	}()

	return nil
}
