package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTinyLFU_Get(t *testing.T) {
	localCache := NewTinyLFU(2097152, time.Minute)
	val, err := localCache.Get([]byte("test"))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, val)

	err = localCache.Set([]byte("test"), []byte("value"))
	require.NoError(t, err)

	val, err = localCache.Get([]byte("test"))
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), val)
}

func TestTinyLFU_Get_expired(t *testing.T) {
	localCache := NewTinyLFU(2097152, time.Second)

	err := localCache.Set([]byte("test"), []byte("value"))
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	val, err := localCache.Get([]byte("test"))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, val)
}

func TestTinyLFU_no_expiration(t *testing.T) {
	localCache := NewTinyLFU(2097152, 0)

	err := localCache.Set([]byte("test"), []byte("value"))
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	val, err := localCache.Get([]byte("test"))
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), val)
}

func TestTinyLFU_SetExp_no_expiration(t *testing.T) {
	localCache := NewTinyLFU(2097152, time.Minute)

	err := localCache.SetExp([]byte("test"), []byte("value"), 0)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	val, err := localCache.Get([]byte("test"))
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), val)
}

func TestTinyLFU_SetExp_with_expiration(t *testing.T) {
	localCache := NewTinyLFU(2097152, 0)

	err := localCache.SetExp([]byte("test"), []byte("value"), time.Second)
	require.NoError(t, err)

	val, err := localCache.Get([]byte("test"))
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), val)

	time.Sleep(2 * time.Second)

	val, err = localCache.Get([]byte("test"))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, val)
}
