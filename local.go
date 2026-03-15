package cache

import "time"

// LocalCache defines the interface for a local in-process cache (L1 layer).
type LocalCache interface {
	Set(key []byte, data []byte) error
	SetExp(key []byte, data []byte, ttl time.Duration) error
	Get(key []byte) ([]byte, error)
	Del(key []byte)
}
