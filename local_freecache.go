package cache

import (
	"fmt"
	"time"

	"github.com/coocood/freecache"
)

// FreeCache is a LocalCache implementation backed by freecache.
type FreeCache struct {
	cache *freecache.Cache
	ttl   time.Duration
}

var _ LocalCache = (*FreeCache)(nil)

// NewFreeCache size is in bytes
func NewFreeCache(size int, ttl time.Duration) *FreeCache {
	return &FreeCache{
		cache: freecache.NewCache(size),
		ttl:   ttl,
	}
}

// Set stores a key-value pair in the local cache using the default TTL.
func (c *FreeCache) Set(key []byte, b []byte) error {
	err := c.cache.Set(key, b, int(c.ttl.Seconds()))
	if err != nil {
		return fmt.Errorf("c.cache.Set: %w", err)
	}
	return nil
}

// SetExp stores a key-value pair in the local cache with a custom TTL.
func (c *FreeCache) SetExp(key []byte, b []byte, ttl time.Duration) error {
	err := c.cache.Set(key, b, int(ttl.Seconds()))
	if err != nil {
		return fmt.Errorf("c.cache.Set: %w", err)
	}
	return nil
}

// Get retrieves a value by key from the local cache.
func (c *FreeCache) Get(key []byte) ([]byte, error) {
	val, err := c.cache.Get(key)
	if err != nil {
		return nil, fmt.Errorf("c.cache.Get: %w", err)
	}
	return val, nil
}

// Del removes a key from the local cache.
func (c *FreeCache) Del(key []byte) {
	c.cache.Del(key)
}
