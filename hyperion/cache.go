package hyperion

import (
	"context"
	"time"
)

// Cache is the caching abstraction.
// Implementations should provide thread-safe caching operations.
type Cache interface {
	// Get retrieves the value for the given key.
	// Returns an error if the key doesn't exist or the operation fails.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores the value for the given key with the specified TTL.
	// A TTL of 0 means no expiration.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes the value for the given key.
	// Returns nil if the key doesn't exist.
	Delete(ctx context.Context, key string) error

	// Exists checks if the key exists in the cache.
	Exists(ctx context.Context, key string) (bool, error)

	// MGet retrieves multiple values for the given keys.
	// Returns a map of key-value pairs for keys that exist.
	// Missing keys are omitted from the result.
	MGet(ctx context.Context, keys ...string) (map[string][]byte, error)

	// MSet stores multiple key-value pairs with the specified TTL.
	MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error

	// Clear removes all entries from the cache.
	// Use with caution in production.
	Clear(ctx context.Context) error
}
