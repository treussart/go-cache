package cache

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type customConfig struct {
	localCache            LocalCache
	statsProm             *StatsProm
	statsOTEL             *StatsOTEL
	remoteCache           redis.UniversalClient
	remoteCacheTTL        time.Duration
	prefixKey             []byte
	cbEnabled             bool
	cbTimeout             time.Duration
	cbMaxRequests         uint32
	cbConsecutiveFailures uint32
	coder                 Coder
}

// CustomOption is a functional option for configuring a simple Cache.
type CustomOption func(*customConfig)

// WithStatsProm sets the Prometheus stats collector.
func WithStatsProm(stats *StatsProm) CustomOption {
	return func(c *customConfig) {
		c.statsProm = stats
	}
}

// WithStatsOTEL sets the OpenTelemetry stats collector.
func WithStatsOTEL(stats *StatsOTEL) CustomOption {
	return func(c *customConfig) {
		c.statsOTEL = stats
	}
}

// WithRedisConn sets the Redis connection and remote cache TTL.
func WithRedisConn(conn redis.UniversalClient, remoteCacheTTL time.Duration) CustomOption {
	return func(c *customConfig) {
		c.remoteCache = conn
		c.remoteCacheTTL = remoteCacheTTL
	}
}

// WithLocalCacheTinyLFU configures a TinyLFU-based local cache with the given size and TTL.
func WithLocalCacheTinyLFU(cacheSize int, localCacheTTL time.Duration) CustomOption {
	return func(c *customConfig) {
		if cacheSize == 0 {
			cacheSize = 10000 // 10 000 items
		}
		c.localCache = NewTinyLFU(cacheSize, localCacheTTL)
	}
}

// WithLocalCacheFreeCache configures a FreeCache-based local cache with the given size and TTL.
func WithLocalCacheFreeCache(cacheSize int, localCacheTTL time.Duration) CustomOption {
	return func(c *customConfig) {
		if cacheSize == 0 {
			cacheSize = 1000000 // 1 MB
		}
		c.localCache = NewFreeCache(cacheSize, localCacheTTL)
	}
}

// WithCBEnabled enables or disables the circuit breaker for Redis operations.
func WithCBEnabled(v bool) CustomOption {
	return func(c *customConfig) {
		c.cbEnabled = v
	}
}

// WithCBTimeout sets the circuit breaker timeout duration.
func WithCBTimeout(v time.Duration) CustomOption {
	return func(c *customConfig) {
		c.cbTimeout = v
	}
}

// WithCBMaxRequests sets the maximum number of requests allowed in the half-open state.
func WithCBMaxRequests(v uint32) CustomOption {
	return func(c *customConfig) {
		c.cbMaxRequests = v
	}
}

// WithCBConsecutiveFailures sets the number of consecutive failures before the circuit breaker trips.
func WithCBConsecutiveFailures(v uint32) CustomOption {
	return func(c *customConfig) {
		c.cbConsecutiveFailures = v
	}
}

// WithCoder sets the encoder/decoder used for cache value serialization.
func WithCoder(v Coder) CustomOption {
	return func(c *customConfig) {
		c.coder = v
	}
}

// WithPrefixKey overrides the key namespace prefix. Pass nil to disable prefixing.
func WithPrefixKey(v []byte) CustomOption {
	return func(c *customConfig) {
		c.prefixKey = v
	}
}
