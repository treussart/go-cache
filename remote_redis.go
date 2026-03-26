package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRemoteCache adapts redis.UniversalClient to the RemoteCache interface.
type RedisRemoteCache struct {
	client redis.UniversalClient
}

var _ RemoteCache = (*RedisRemoteCache)(nil)

// NewRedisRemoteCache creates a RemoteCache backed by Redis.
func NewRedisRemoteCache(client redis.UniversalClient) *RedisRemoteCache {
	return &RedisRemoteCache{client: client}
}

// Get retrieves a value by key. Returns ErrCacheMiss when the key does not exist.
func (r *RedisRemoteCache) Get(ctx context.Context, key string) ([]byte, error) {
	b, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		return nil, err
	}
	return b, nil
}

// Set stores a key-value pair with the given TTL.
func (r *RedisRemoteCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Del removes a key.
func (r *RedisRemoteCache) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Ping checks whether the Redis connection is healthy.
func (r *RedisRemoteCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// WithRedisConn sets the remote cache to a Redis connection.
func WithRedisConn(conn redis.UniversalClient, remoteCacheTTL time.Duration) CustomOption {
	return func(c *customConfig) {
		c.remoteCache = NewRedisRemoteCache(conn)
		c.remoteCacheTTL = remoteCacheTTL
	}
}
