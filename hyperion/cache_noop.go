package hyperion

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrNoOpCache is returned by no-op cache operations.
	ErrNoOpCache = errors.New("no-op cache: no adapter provided")
)

// noopCache is a no-op implementation of Cache interface.
type noopCache struct{}

// NewNoOpCache creates a new no-op Cache implementation.
func NewNoOpCache() Cache {
	return &noopCache{}
}

func (c *noopCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, ErrNoOpCache
}

func (c *noopCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

func (c *noopCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (c *noopCache) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (c *noopCache) MGet(ctx context.Context, keys ...string) (map[string][]byte, error) {
	return nil, ErrNoOpCache
}

func (c *noopCache) MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	return nil
}

func (c *noopCache) Clear(ctx context.Context) error {
	return nil
}
