package ratelimit

import (
	"context"
	"sync"
)

func NewInMemoryContainer() *inMemoryContainer {
	container := &inMemoryContainer{
		items: make(map[IP]RateLimiterStrategy),
		mux:   sync.RWMutex{},
	}

	return container
}

type inMemoryContainer struct {
	items map[IP]RateLimiterStrategy
	mux   sync.RWMutex
}

func (c *inMemoryContainer) delete(ctx context.Context, ip IP) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	delete(c.items, ip)

	return nil
}

func (c *inMemoryContainer) Has(ctx context.Context, ip IP) bool {
	c.mux.RLock()
	defer c.mux.RUnlock()

	_, ok := c.items[ip]

	return ok
}

func (c *inMemoryContainer) Consume(ctx context.Context, ip IP) error {
	c.mux.RLock()
	item, ok := c.items[ip]
	c.mux.RUnlock()

	if ok {
		if err := item.Consume(); err != nil {
			return err
		}
	}

	return nil
}

func (c *inMemoryContainer) New(ctx context.Context, ip IP, strategy RateLimiterStrategy) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.items[ip] = strategy

	ch := strategy.GetDestroySignal()
	go func() {
		_ = c.delete(ctx, <-ch)
	}()

	return nil
}
