package cache

import "errors"

var (
	// ErrCacheMiss is returned when the requested key is not found in the cache.
	ErrCacheMiss = errors.New("cache key is missing")

	// ErrKeyEmpty is returned when an empty key is provided to a cache operation.
	ErrKeyEmpty = errors.New("key is empty")

	// ErrInitCache is returned when cache initialization fails due to missing RemoteCache and LocalCache.
	ErrInitCache = errors.New("can not init cache : RemoteCache or LocalCache must not be nil")

	// ErrConvertingToBytes is returned when a cache value cannot be converted to []byte.
	ErrConvertingToBytes = errors.New("type error when converting to []byte")
)
