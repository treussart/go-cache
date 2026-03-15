package cache

import (
	"sync"
	"time"

	"github.com/vmihailenco/go-tinylfu"
)

const tinyLFUSamples = 100000

// TinyLFU is a concurrency-safe local cache using the TinyLFU admission policy.
type TinyLFU struct {
	mu  sync.Mutex
	lfu *tinylfu.T
	ttl time.Duration
}

var _ LocalCache = (*TinyLFU)(nil)

// NewTinyLFU creates a TinyLFU cache. size is the maximum number of items.
func NewTinyLFU(size int, ttl time.Duration) *TinyLFU {
	return &TinyLFU{
		lfu: tinylfu.New(size, tinyLFUSamples),
		ttl: ttl,
	}
}

// Set stores a key-value pair in the cache using the default TTL.
func (c *TinyLFU) Set(key []byte, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lfu.Set(&tinylfu.Item{
		Key:      string(key),
		Value:    data,
		ExpireAt: expireAt(c.ttl),
	})
	return nil
}

// SetExp stores a key-value pair in the cache with a custom TTL.
func (c *TinyLFU) SetExp(key []byte, data []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lfu.Set(&tinylfu.Item{
		Key:      string(key),
		Value:    data,
		ExpireAt: expireAt(ttl),
	})
	return nil
}

// expireAt returns the expiration time for the given TTL.
// A zero TTL means no expiration (zero time.Time keeps tinylfu items alive indefinitely).
func expireAt(ttl time.Duration) time.Time {
	if ttl > 0 {
		return time.Now().Add(ttl)
	}
	return time.Time{}
}

// Get retrieves a value by key from the cache.
func (c *TinyLFU) Get(key []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	val, ok := c.lfu.Get(string(key))
	if !ok {
		return nil, ErrCacheMiss
	}

	b, ok := val.([]byte)
	if !ok {
		return nil, ErrConvertingToBytes
	}
	return b, nil
}

// Del removes a key from the cache.
func (c *TinyLFU) Del(key []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lfu.Del(string(key))
}
