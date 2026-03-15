package cache

import (
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFreecache_Get(t *testing.T) {
	localCache := NewFreeCache(2097152, time.Minute)
	val, err := localCache.Get([]byte("test"))
	require.ErrorIs(t, err, freecache.ErrNotFound)
	assert.Nil(t, val)

	err = localCache.Set([]byte("test"), []byte("value"))
	require.NoError(t, err)

	val, err = localCache.Get([]byte("test"))
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), val)
}

func TestFreecache_Get_expired(t *testing.T) {
	localCache := NewFreeCache(2097152, time.Second)

	err := localCache.Set([]byte("test"), []byte("value"))
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	val, err := localCache.Get([]byte("test"))
	require.ErrorIs(t, err, freecache.ErrNotFound)
	assert.Nil(t, val)
}

func TestFreecache_no_expiration(t *testing.T) {
	localCache := NewFreeCache(2097152, 0)

	err := localCache.Set([]byte("test"), []byte("value"))
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	val, err := localCache.Get([]byte("test"))
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), val)
}

func TestFreecache_SetExp_no_expiration(t *testing.T) {
	localCache := NewFreeCache(2097152, time.Minute)

	err := localCache.SetExp([]byte("test"), []byte("value"), 0)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	val, err := localCache.Get([]byte("test"))
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), val)
}
