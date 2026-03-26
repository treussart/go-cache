package cache

import (
	"context"
	"time"
)

// RemoteCache defines the interface for a remote cache (L2 layer).
// Get must return ErrCacheMiss when the requested key does not exist.
type RemoteCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	Ping(ctx context.Context) error
}
