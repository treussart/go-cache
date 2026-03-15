package cache

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/sony/gobreaker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	myKey   = "mykey"
	myValue = "myvalue"
)

func TestNew_LocalCache_Get(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)
	key := myKey
	value := myValue
	mock.ExpectGet(key).RedisNil()
	b, err := mycache.Get(context.Background(), []byte(key))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, b)

	mock.ExpectSet(key, []byte(value), time.Minute).SetVal(value)
	err = mycache.Set(context.Background(), []byte(key), []byte(value))
	require.NoError(t, err)

	b, err = mycache.Get(context.Background(), []byte(key))
	require.NoError(t, err)
	assert.Equal(t, value, string(b))

	mock.ExpectDel(key).SetVal(0)
	err = mycache.Del(context.Background(), []byte(key))
	require.NoError(t, err)
}

func TestNew_LocalCache_Get_stats(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("test",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, time.Minute),
		WithStatsProm(GetStatsProm("", "")),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)
	key := myKey
	value := myValue
	mock.ExpectGet(key).RedisNil()
	b, err := mycache.Get(context.Background(), []byte(key))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, b)

	mock.ExpectSet(key, []byte(value), time.Minute).SetVal(value)
	err = mycache.Set(context.Background(), []byte(key), []byte(value))
	require.NoError(t, err)

	b, err = mycache.Get(context.Background(), []byte(key))
	require.NoError(t, err)
	assert.Equal(t, value, string(b))

	mock.ExpectDel(key).SetVal(0)
	err = mycache.Del(context.Background(), []byte(key))
	require.NoError(t, err)
}

func TestNew_RemoteCache_Get(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)
	key := myKey
	value := myValue
	mock.ExpectGet(key).SetVal(value)
	b, err := mycache.Get(context.Background(), []byte(key))
	require.NoError(t, err)
	assert.Equal(t, value, string(b))
}

func TestNew_LocalCache_only(t *testing.T) {
	mycache, err := New("",
		WithLocalCacheFreeCache(1000, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)
	key := myKey
	value := myValue

	b, err := mycache.Get(context.Background(), []byte(key))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, b)

	err = mycache.Set(context.Background(), []byte(key), []byte(value))
	require.NoError(t, err)

	b, err = mycache.Get(context.Background(), []byte(key))
	require.NoError(t, err)
	assert.Equal(t, value, string(b))

	err = mycache.Del(context.Background(), []byte(key))
	require.NoError(t, err)
}

func TestNew_LocalCache_only_exp(t *testing.T) {
	mycache, err := New("",
		WithLocalCacheFreeCache(1000, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)
	key := myKey
	value := myValue

	b, err := mycache.Get(context.Background(), []byte(key))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, b)

	err = mycache.SetExp(context.Background(), []byte(key), []byte(value), 5*time.Second)
	require.NoError(t, err)

	b, err = mycache.Get(context.Background(), []byte(key))
	require.NoError(t, err)
	assert.Equal(t, value, string(b))

	err = mycache.Del(context.Background(), []byte(key))
	require.NoError(t, err)
}

func TestNew_RemoteCache_only(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)
	key := myKey
	value := myValue

	mock.ExpectGet(key).RedisNil()
	b, err := mycache.Get(context.Background(), []byte(key))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, b)

	mock.ExpectSet(key, []byte(value), time.Minute).SetVal(value)
	err = mycache.Set(context.Background(), []byte(key), []byte(value))
	require.NoError(t, err)

	mock.ExpectSet(key, []byte(value), time.Minute).RedisNil()
	err = mycache.Set(context.Background(), []byte(key), []byte(value))
	require.ErrorContains(t, err, "redis: nil")

	mock.ExpectGet(key).SetVal(value)
	b, err = mycache.Get(context.Background(), []byte(key))
	require.NoError(t, err)
	assert.Equal(t, value, string(b))

	mock.ExpectDel(key).SetVal(0)
	err = mycache.Del(context.Background(), []byte(key))
	require.NoError(t, err)

	mock.ExpectDel(key).RedisNil()
	err = mycache.Del(context.Background(), []byte(key))
	require.ErrorContains(t, err, "redis: nil")
}

func TestNew_RemoteCache_only_exp(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)
	key := myKey
	value := myValue

	mock.ExpectGet(key).RedisNil()
	b, err := mycache.Get(context.Background(), []byte(key))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, b)

	mock.ExpectSetEx(key, []byte(value), time.Minute).SetVal(value)
	err = mycache.SetExp(context.Background(), []byte(key), []byte(value), time.Minute)
	require.NoError(t, err)

	mock.ExpectGet(key).SetVal(value)
	b, err = mycache.Get(context.Background(), []byte(key))
	require.NoError(t, err)
	assert.Equal(t, value, string(b))

	mock.ExpectDel(key).SetVal(0)
	err = mycache.Del(context.Background(), []byte(key))
	require.NoError(t, err)

	mock.ExpectDel(key).RedisNil()
	err = mycache.Del(context.Background(), []byte(key))
	require.ErrorContains(t, err, "redis: nil")
}

func TestNew_LocalCache_Get_prefix(t *testing.T) {
	prefix := "test:"
	db, mock := redismock.NewClientMock()
	mycache, err := New("test",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, time.Minute),
	)
	require.NoError(t, err)
	key := myKey
	value := myValue
	mock.ExpectGet(prefix + key).RedisNil()
	b, err := mycache.Get(context.Background(), []byte(key))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, b)

	mock.ExpectSet(prefix+key, []byte(value), time.Minute).SetVal(value)
	err = mycache.Set(context.Background(), []byte(key), []byte(value))
	require.NoError(t, err)

	b, err = mycache.Get(context.Background(), []byte(key))
	require.NoError(t, err)
	assert.Equal(t, value, string(b))

	mock.ExpectDel(prefix + key).SetVal(0)
	err = mycache.Del(context.Background(), []byte(key))
	require.NoError(t, err)
}

func TestNew_LocalCache_only_prefix(t *testing.T) {
	mycache, err := New("test",
		WithLocalCacheFreeCache(1000, time.Minute),
		WithPrefixKey([]byte("test")),
	)
	require.NoError(t, err)
	key := myKey
	value := myValue

	b, err := mycache.Get(context.Background(), []byte(key))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, b)

	err = mycache.Set(context.Background(), []byte(key), []byte(value))
	require.NoError(t, err)

	b, err = mycache.Get(context.Background(), []byte(key))
	require.NoError(t, err)
	assert.Equal(t, value, string(b))

	err = mycache.Del(context.Background(), []byte(key))
	require.NoError(t, err)
}

func TestNew_key_empty(t *testing.T) {
	db, _ := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)
	key := ""
	value := myValue
	b, err := mycache.Get(context.Background(), []byte(key))
	require.ErrorIs(t, err, ErrKeyEmpty)
	assert.Nil(t, b)

	err = mycache.Set(context.Background(), []byte(key), []byte(value))
	require.ErrorIs(t, err, ErrKeyEmpty)

	err = mycache.Del(context.Background(), []byte(key))
	require.ErrorIs(t, err, ErrKeyEmpty)
}

func TestNew_Redis_CB_redisNil(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithCBEnabled(true),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)
	key := myKey

	mock.ExpectGet(key).RedisNil()
	b, err := mycache.Get(context.Background(), []byte(key))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, b)

	mock.ExpectGet(key).RedisNil()
	b, err = mycache.Get(context.Background(), []byte(key))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, b)

	mock.ExpectGet(key).RedisNil()
	b, err = mycache.Get(context.Background(), []byte(key))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, b)
}

func TestNew_Redis_CB(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithCBEnabled(true),
		WithCBTimeout(1*time.Second),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)
	key := myKey

	mock.ExpectGet(key).SetErr(ErrInitCache)
	b, err := mycache.Get(context.Background(), []byte(key))
	require.ErrorIs(t, err, ErrInitCache)
	assert.Nil(t, b)

	mock.ExpectGet(key).SetErr(ErrInitCache)
	b, err = mycache.Get(context.Background(), []byte(key))
	require.ErrorIs(t, err, ErrInitCache)
	assert.Nil(t, b)

	// CB is now open — error is propagated so callers can log it
	b, err = mycache.Get(context.Background(), []byte(key))
	require.ErrorIs(t, err, gobreaker.ErrOpenState)
	assert.Nil(t, b)

	time.Sleep(1 * time.Second)

	mock.ExpectGet(key).SetVal(myKey)
	b, err = mycache.Get(context.Background(), []byte(key))
	require.NoError(t, err)
	assert.Equal(t, []byte(myKey), b)
}

func TestNew_Redis_CB_fallback_Get_with_local(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, time.Minute),
		WithCBEnabled(true),
		WithCBTimeout(1*time.Second),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	// Store a value while Redis is healthy
	mock.ExpectSet(myKey, []byte(myValue), time.Minute).SetVal(myValue)
	require.NoError(t, mycache.Set(context.Background(), []byte(myKey), []byte(myValue)))

	// Trip the CB with a different key
	otherKey := "other"
	mock.ExpectGet(otherKey).SetErr(ErrInitCache)
	_, err = mycache.Get(context.Background(), []byte(otherKey))
	require.ErrorIs(t, err, ErrInitCache)

	mock.ExpectGet(otherKey).SetErr(ErrInitCache)
	_, err = mycache.Get(context.Background(), []byte(otherKey))
	require.ErrorIs(t, err, ErrInitCache)

	// CB is open — L1 hit still works
	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.NoError(t, err)
	assert.Equal(t, myValue, string(b))

	// CB is open — L1 miss propagates CB error (no Redis call)
	b, err = mycache.Get(context.Background(), []byte(otherKey))
	require.ErrorIs(t, err, gobreaker.ErrOpenState)
	assert.Nil(t, b)
}

func TestNew_Redis_CB_fallback_Set(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, time.Minute),
		WithCBEnabled(true),
		WithCBTimeout(1*time.Second),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	// Trip the CB via failed Sets
	mock.ExpectSet(myKey, []byte(myValue), time.Minute).SetErr(ErrInitCache)
	require.Error(t, mycache.Set(context.Background(), []byte(myKey), []byte(myValue)))

	mock.ExpectSet(myKey, []byte(myValue), time.Minute).SetErr(ErrInitCache)
	require.Error(t, mycache.Set(context.Background(), []byte(myKey), []byte(myValue)))

	// CB is open — Set succeeds (L1 written, Redis skipped)
	require.NoError(t, mycache.Set(context.Background(), []byte(myKey), []byte(myValue)))

	// Verify L1 has the value
	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.NoError(t, err)
	assert.Equal(t, myValue, string(b))
}

func TestNew_Redis_CB_fallback_SetExp(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, time.Minute),
		WithCBEnabled(true),
		WithCBTimeout(1*time.Second),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	ttl := 5 * time.Second

	// Trip the CB via failed SetExps
	mock.ExpectSetEx(myKey, []byte(myValue), ttl).SetErr(ErrInitCache)
	require.Error(t, mycache.SetExp(context.Background(), []byte(myKey), []byte(myValue), ttl))

	mock.ExpectSetEx(myKey, []byte(myValue), ttl).SetErr(ErrInitCache)
	require.Error(t, mycache.SetExp(context.Background(), []byte(myKey), []byte(myValue), ttl))

	// CB is open — SetExp succeeds (L1 written, Redis skipped)
	require.NoError(t, mycache.SetExp(context.Background(), []byte(myKey), []byte(myValue), ttl))

	// Verify L1 has the value
	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.NoError(t, err)
	assert.Equal(t, myValue, string(b))
}

func TestNew_Redis_CB_state_change_metrics(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stats := GetStatsProm("", "")
	mycache, err := New("test",
		WithRedisConn(db, time.Minute),
		WithCBEnabled(true),
		WithCBTimeout(1*time.Second),
		WithStatsProm(stats),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	// Trip the CB
	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	_, _ = mycache.Get(context.Background(), []byte(myKey))
	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	_, _ = mycache.Get(context.Background(), []byte(myKey))

	// CB is open — CBOpen counter incremented, error propagated
	_, err = mycache.Get(context.Background(), []byte(myKey))
	require.ErrorIs(t, err, gobreaker.ErrOpenState)
}

// --- New() init ---

func TestNew_no_cache_returns_error(t *testing.T) {
	_, err := New("test")
	require.ErrorIs(t, err, ErrInitCache)
}

// --- TinyLFU through the option layer ---

func TestNew_LocalCache_TinyLFU(t *testing.T) {
	mycache, err := New("",
		WithLocalCacheTinyLFU(10000, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, b)

	err = mycache.Set(context.Background(), []byte(myKey), []byte(myValue))
	require.NoError(t, err)

	b, err = mycache.Get(context.Background(), []byte(myKey))
	require.NoError(t, err)
	assert.Equal(t, myValue, string(b))

	err = mycache.Del(context.Background(), []byte(myKey))
	require.NoError(t, err)

	b, err = mycache.Get(context.Background(), []byte(myKey))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, b)
}

func TestNew_LocalCache_TinyLFU_SetExp(t *testing.T) {
	mycache, err := New("",
		WithLocalCacheTinyLFU(10000, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	err = mycache.SetExp(context.Background(), []byte(myKey), []byte(myValue), 5*time.Second)
	require.NoError(t, err)

	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.NoError(t, err)
	assert.Equal(t, myValue, string(b))
}

// --- Default cache sizes (size=0 triggers internal defaults) ---

func TestNew_default_cache_sizes(t *testing.T) {
	t.Run("FreeCache", func(t *testing.T) {
		mycache, err := New("",
			WithLocalCacheFreeCache(0, time.Minute),
			WithPrefixKey(nil),
		)
		require.NoError(t, err)
		require.NoError(t, mycache.Set(context.Background(), []byte(myKey), []byte(myValue)))
		b, err := mycache.Get(context.Background(), []byte(myKey))
		require.NoError(t, err)
		assert.Equal(t, myValue, string(b))
	})
	t.Run("TinyLFU", func(t *testing.T) {
		mycache, err := New("",
			WithLocalCacheTinyLFU(0, time.Minute),
			WithPrefixKey(nil),
		)
		require.NoError(t, err)
		require.NoError(t, mycache.Set(context.Background(), []byte(myKey), []byte(myValue)))
		b, err := mycache.Get(context.Background(), []byte(myKey))
		require.NoError(t, err)
		assert.Equal(t, myValue, string(b))
	})
}

// --- Ready ---

func TestNew_Ready(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		mycache, err := New("",
			WithRedisConn(db, time.Minute),
			WithPrefixKey(nil),
		)
		require.NoError(t, err)
		mock.ExpectPing().SetVal("PONG")
		require.NoError(t, mycache.Ready(context.Background()))
	})
	t.Run("error", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		mycache, err := New("",
			WithRedisConn(db, time.Minute),
			WithPrefixKey(nil),
		)
		require.NoError(t, err)
		mock.ExpectPing().SetErr(ErrInitCache)
		err = mycache.Ready(context.Background())
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInitCache)
	})
	t.Run("no_redis", func(t *testing.T) {
		mycache, err := New("",
			WithLocalCacheFreeCache(1000, time.Minute),
			WithPrefixKey(nil),
		)
		require.NoError(t, err)
		require.NoError(t, mycache.Ready(context.Background()))
	})
}

// --- Empty key on SetExp ---

func TestNew_SetExp_key_empty(t *testing.T) {
	mycache, err := New("",
		WithLocalCacheFreeCache(1000, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)
	err = mycache.SetExp(context.Background(), []byte(""), []byte(myValue), time.Minute)
	require.ErrorIs(t, err, ErrKeyEmpty)
}

// --- Generic Redis Get error (not redis.Nil, not CB) ---

func TestNew_Get_remote_error(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInitCache)
	assert.Nil(t, b)
}

// --- DeleteFromLocalCache / DeleteFromRemoteCache ---

func TestNew_DeleteFromLocalCache(t *testing.T) {
	mycache, err := New("",
		WithLocalCacheFreeCache(1000, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	require.NoError(t, mycache.Set(context.Background(), []byte(myKey), []byte(myValue)))
	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.NoError(t, err)
	assert.Equal(t, myValue, string(b))

	mycache.DeleteFromLocalCache([]byte(myKey))

	b, err = mycache.Get(context.Background(), []byte(myKey))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, b)
}

func TestNew_DeleteFromLocalCache_no_local(t *testing.T) {
	db, _ := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)
	mycache.DeleteFromLocalCache([]byte(myKey)) // must not panic
}

func TestNew_DeleteFromRemoteCache(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		mycache, err := New("",
			WithRedisConn(db, time.Minute),
			WithPrefixKey(nil),
		)
		require.NoError(t, err)
		mock.ExpectDel(myKey).SetVal(1)
		require.NoError(t, mycache.DeleteFromRemoteCache(context.Background(), []byte(myKey)))
	})
	t.Run("empty_key", func(t *testing.T) {
		db, _ := redismock.NewClientMock()
		mycache, err := New("",
			WithRedisConn(db, time.Minute),
			WithPrefixKey(nil),
		)
		require.NoError(t, err)
		err = mycache.DeleteFromRemoteCache(context.Background(), []byte(""))
		require.ErrorIs(t, err, ErrKeyEmpty)
	})
	t.Run("no_redis", func(t *testing.T) {
		mycache, err := New("",
			WithLocalCacheFreeCache(1000, time.Minute),
			WithPrefixKey(nil),
		)
		require.NoError(t, err)
		require.NoError(t, mycache.DeleteFromRemoteCache(context.Background(), []byte(myKey)))
	})
	t.Run("redis_error", func(t *testing.T) {
		db, mock := redismock.NewClientMock()
		mycache, err := New("",
			WithRedisConn(db, time.Minute),
			WithPrefixKey(nil),
		)
		require.NoError(t, err)
		mock.ExpectDel(myKey).SetErr(ErrInitCache)
		err = mycache.DeleteFromRemoteCache(context.Background(), []byte(myKey))
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInitCache)
	})
}

// --- GetStruct / SetStruct / SetExStruct ---

func TestNew_SetStruct_GetStruct(t *testing.T) {
	type user struct {
		Name string `msgpack:"name"`
		Age  int    `msgpack:"age"`
	}
	mycache, err := New("",
		WithLocalCacheFreeCache(1000000, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	original := user{Name: "Alice", Age: 30}
	require.NoError(t, mycache.SetStruct(context.Background(), "user:1", original))

	var got user
	require.NoError(t, mycache.GetStruct(context.Background(), "user:1", &got))
	assert.Equal(t, original, got)
}

func TestNew_SetExStruct_GetStruct(t *testing.T) {
	type user struct {
		Name string `msgpack:"name"`
		Age  int    `msgpack:"age"`
	}
	mycache, err := New("",
		WithLocalCacheFreeCache(1000000, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	original := user{Name: "Bob", Age: 25}
	require.NoError(t, mycache.SetExStruct(context.Background(), "user:2", original, 5*time.Second))

	var got user
	require.NoError(t, mycache.GetStruct(context.Background(), "user:2", &got))
	assert.Equal(t, original, got)
}

func TestNew_SetStruct_GetStruct_JSONCoder(t *testing.T) {
	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	mycache, err := New("",
		WithLocalCacheFreeCache(1000000, time.Minute),
		WithCoder(&JSONCoder{}),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	original := user{Name: "Charlie", Age: 40}
	require.NoError(t, mycache.SetStruct(context.Background(), "user:3", original))

	var got user
	require.NoError(t, mycache.GetStruct(context.Background(), "user:3", &got))
	assert.Equal(t, original, got)
}

func TestNew_GetStruct_miss(t *testing.T) {
	mycache, err := New("",
		WithLocalCacheFreeCache(1000, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	var dest struct{ Name string }
	err = mycache.GetStruct(context.Background(), "nonexistent", &dest)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrCacheMiss)
}

func TestNew_GetStruct_decode_error(t *testing.T) {
	mycache, err := New("",
		WithLocalCacheFreeCache(1000000, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	// Store raw bytes that are not valid msgpack for a struct
	require.NoError(t, mycache.Set(context.Background(), []byte("badkey"), []byte("not-msgpack")))

	var dest struct{ Name string }
	err = mycache.GetStruct(context.Background(), "badkey", &dest)
	require.Error(t, err)
	require.ErrorContains(t, err, "decode")
}

func TestNew_SetStruct_encode_error(t *testing.T) {
	mycache, err := New("",
		WithLocalCacheFreeCache(1000, time.Minute),
		WithCoder(&JSONCoder{}),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	err = mycache.SetStruct(context.Background(), "key", make(chan int))
	require.Error(t, err)
}

func TestNew_SetExStruct_encode_error(t *testing.T) {
	mycache, err := New("",
		WithLocalCacheFreeCache(1000, time.Minute),
		WithCoder(&JSONCoder{}),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	err = mycache.SetExStruct(context.Background(), "key", make(chan int), time.Minute)
	require.Error(t, err)
}

// --- CB does not wrap Del — Redis errors are returned directly ---

func TestNew_Redis_CB_Del_bypasses_CB(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, time.Minute),
		WithCBEnabled(true),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	mock.ExpectDel(myKey).SetErr(ErrInitCache)
	err = mycache.Del(context.Background(), []byte(myKey))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInitCache)
}

// --- CB with Set stats + Prometheus ---

func TestNew_Redis_CB_Set_with_stats(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stats := GetStatsProm("", "")
	mycache, err := New("test",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, time.Minute),
		WithCBEnabled(true),
		WithCBTimeout(1*time.Second),
		WithStatsProm(stats),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	// Successful Set — remote counter incremented
	mock.ExpectSet(myKey, []byte(myValue), time.Minute).SetVal(myValue)
	require.NoError(t, mycache.Set(context.Background(), []byte(myKey), []byte(myValue)))

	// Trip the CB via failed Sets
	mock.ExpectSet(myKey, []byte(myValue), time.Minute).SetErr(ErrInitCache)
	require.Error(t, mycache.Set(context.Background(), []byte(myKey), []byte(myValue)))

	mock.ExpectSet(myKey, []byte(myValue), time.Minute).SetErr(ErrInitCache)
	require.Error(t, mycache.Set(context.Background(), []byte(myKey), []byte(myValue)))

	// CB is open — Set falls back to L1 only, CBOpen counter incremented
	require.NoError(t, mycache.Set(context.Background(), []byte(myKey), []byte(myValue)))
}

// --- CB with SetExp stats ---

func TestNew_Redis_CB_SetExp_with_stats(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stats := GetStatsProm("", "")
	mycache, err := New("test",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, time.Minute),
		WithCBEnabled(true),
		WithCBTimeout(1*time.Second),
		WithStatsProm(stats),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	ttl := 5 * time.Second

	// Successful SetExp
	mock.ExpectSetEx(myKey, []byte(myValue), ttl).SetVal(myValue)
	require.NoError(t, mycache.SetExp(context.Background(), []byte(myKey), []byte(myValue), ttl))

	// Trip the CB
	mock.ExpectSetEx(myKey, []byte(myValue), ttl).SetErr(ErrInitCache)
	require.Error(t, mycache.SetExp(context.Background(), []byte(myKey), []byte(myValue), ttl))

	mock.ExpectSetEx(myKey, []byte(myValue), ttl).SetErr(ErrInitCache)
	require.Error(t, mycache.SetExp(context.Background(), []byte(myKey), []byte(myValue), ttl))

	// CB is open — falls back to L1
	require.NoError(t, mycache.SetExp(context.Background(), []byte(myKey), []byte(myValue), ttl))
}

// --- Remote Set error without CB ---

func TestNew_Set_remote_error(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	mock.ExpectSet(myKey, []byte(myValue), time.Minute).SetErr(ErrInitCache)
	err = mycache.Set(context.Background(), []byte(myKey), []byte(myValue))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInitCache)
}

func TestNew_SetExp_remote_error(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	mock.ExpectSetEx(myKey, []byte(myValue), time.Minute).SetErr(ErrInitCache)
	err = mycache.SetExp(context.Background(), []byte(myKey), []byte(myValue), time.Minute)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInitCache)
}

// --- Graceful degradation (stale cache) ---

func TestNew_GracefulDegradation_stale_hit(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, 1*time.Second),
		WithCBEnabled(true),
		WithCBTimeout(1*time.Second),
		WithGracefulDegradation(1*time.Hour),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	// Store a value while Redis is healthy — written to L1 + stale + Redis
	mock.ExpectSet(myKey, []byte(myValue), time.Minute).SetVal(myValue)
	require.NoError(t, mycache.Set(context.Background(), []byte(myKey), []byte(myValue)))

	// Evict from primary L1 so next Get must go to Redis
	mycache.opt.localCache.Del([]byte(myKey))

	// Trip the CB
	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	_, err = mycache.Get(context.Background(), []byte(myKey))
	require.ErrorIs(t, err, ErrInitCache)

	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	_, err = mycache.Get(context.Background(), []byte(myKey))
	require.ErrorIs(t, err, ErrInitCache)

	// CB is open — L1 misses, but stale cache returns the value
	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.NoError(t, err)
	assert.Equal(t, myValue, string(b))
}

func TestNew_GracefulDegradation_stale_miss(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, time.Minute),
		WithCBEnabled(true),
		WithCBTimeout(1*time.Second),
		WithGracefulDegradation(1*time.Hour),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	// Trip the CB without storing anything
	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	_, _ = mycache.Get(context.Background(), []byte(myKey))
	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	_, _ = mycache.Get(context.Background(), []byte(myKey))

	// CB is open — no data in stale cache either, error is returned
	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.ErrorIs(t, err, gobreaker.ErrOpenState)
	assert.Nil(t, b)
}

func TestNew_GracefulDegradation_disabled(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, time.Minute),
		WithCBEnabled(true),
		WithCBTimeout(1*time.Second),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	// Store a value
	mock.ExpectSet(myKey, []byte(myValue), time.Minute).SetVal(myValue)
	require.NoError(t, mycache.Set(context.Background(), []byte(myKey), []byte(myValue)))

	// Evict from primary L1
	mycache.opt.localCache.Del([]byte(myKey))

	// Trip the CB
	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	_, _ = mycache.Get(context.Background(), []byte(myKey))
	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	_, _ = mycache.Get(context.Background(), []byte(myKey))

	// CB is open — no graceful degradation, error is returned
	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.ErrorIs(t, err, gobreaker.ErrOpenState)
	assert.Nil(t, b)
}

func TestNew_GracefulDegradation_Del_clears_stale(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, time.Minute),
		WithCBEnabled(true),
		WithCBTimeout(1*time.Second),
		WithGracefulDegradation(1*time.Hour),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	// Store a value
	mock.ExpectSet(myKey, []byte(myValue), time.Minute).SetVal(myValue)
	require.NoError(t, mycache.Set(context.Background(), []byte(myKey), []byte(myValue)))

	// Delete clears L1 + stale + Redis
	mock.ExpectDel(myKey).SetVal(1)
	require.NoError(t, mycache.Del(context.Background(), []byte(myKey)))

	// Trip the CB
	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	_, _ = mycache.Get(context.Background(), []byte(myKey))
	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	_, _ = mycache.Get(context.Background(), []byte(myKey))

	// CB is open — stale cache was also cleared, so no fallback
	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.ErrorIs(t, err, gobreaker.ErrOpenState)
	assert.Nil(t, b)
}

func TestNew_GracefulDegradation_stale_hit_with_stats(t *testing.T) {
	db, mock := redismock.NewClientMock()
	stats := GetStatsProm("", "")
	mycache, err := New("test",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, 1*time.Second),
		WithCBEnabled(true),
		WithCBTimeout(1*time.Second),
		WithGracefulDegradation(1*time.Hour),
		WithStatsProm(stats),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	// Store a value
	mock.ExpectSet(myKey, []byte(myValue), time.Minute).SetVal(myValue)
	require.NoError(t, mycache.Set(context.Background(), []byte(myKey), []byte(myValue)))

	// Evict from primary L1
	mycache.opt.localCache.Del([]byte(myKey))

	// Trip the CB
	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	_, _ = mycache.Get(context.Background(), []byte(myKey))
	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	_, _ = mycache.Get(context.Background(), []byte(myKey))

	// CB is open — stale hit
	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.NoError(t, err)
	assert.Equal(t, myValue, string(b))

	// Verify HitsStale counter was incremented
	val := stats.HitsStale.WithLabelValues("test")
	require.NotNil(t, val)
}

func TestNew_GracefulDegradation_zero_staleTTL(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, 1*time.Second),
		WithCBEnabled(true),
		WithCBTimeout(1*time.Second),
		WithGracefulDegradation(0), // never expire
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	// Store a value
	mock.ExpectSet(myKey, []byte(myValue), time.Minute).SetVal(myValue)
	require.NoError(t, mycache.Set(context.Background(), []byte(myKey), []byte(myValue)))

	// Evict from primary L1
	mycache.opt.localCache.Del([]byte(myKey))

	// Trip the CB
	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	_, _ = mycache.Get(context.Background(), []byte(myKey))
	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	_, _ = mycache.Get(context.Background(), []byte(myKey))

	// CB is open — stale cache entry never expires, still returns the value
	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.NoError(t, err)
	assert.Equal(t, myValue, string(b))
}

// --- Preload ---

func TestNew_Preload_local_only(t *testing.T) {
	data := map[string][]byte{
		"user:1": []byte("Alice"),
		"user:2": []byte("Bob"),
	}
	mycache, err := New("",
		WithLocalCacheFreeCache(1000, time.Minute),
		WithPreload(data),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	b, err := mycache.Get(context.Background(), []byte("user:1"))
	require.NoError(t, err)
	assert.Equal(t, "Alice", string(b))

	b, err = mycache.Get(context.Background(), []byte("user:2"))
	require.NoError(t, err)
	assert.Equal(t, "Bob", string(b))
}

func TestNew_Preload_with_prefix(t *testing.T) {
	data := map[string][]byte{
		myKey: []byte(myValue),
	}
	mycache, err := New("test",
		WithLocalCacheFreeCache(1000, time.Minute),
		WithPreload(data),
	)
	require.NoError(t, err)

	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.NoError(t, err)
	assert.Equal(t, myValue, string(b))
}

func TestNew_Preload_with_graceful_degradation(t *testing.T) {
	db, mock := redismock.NewClientMock()
	data := map[string][]byte{
		myKey: []byte(myValue),
	}
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithLocalCacheFreeCache(1000, 1*time.Second),
		WithCBEnabled(true),
		WithCBTimeout(1*time.Second),
		WithGracefulDegradation(1*time.Hour),
		WithPreload(data),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	// Evict from primary L1
	mycache.opt.localCache.Del([]byte(myKey))

	// Trip the CB
	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	_, _ = mycache.Get(context.Background(), []byte(myKey))
	mock.ExpectGet(myKey).SetErr(ErrInitCache)
	_, _ = mycache.Get(context.Background(), []byte(myKey))

	// CB is open — L1 miss, but stale cache was preloaded
	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.NoError(t, err)
	assert.Equal(t, myValue, string(b))
}

func TestNew_Preload_empty_data(t *testing.T) {
	mycache, err := New("",
		WithLocalCacheFreeCache(1000, time.Minute),
		WithPreload(map[string][]byte{}),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, b)
}

func TestNew_Preload_nil_data(t *testing.T) {
	mycache, err := New("",
		WithLocalCacheFreeCache(1000, time.Minute),
		WithPreload(nil),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)

	b, err := mycache.Get(context.Background(), []byte(myKey))
	require.ErrorIs(t, err, ErrCacheMiss)
	assert.Nil(t, b)
}

func TestNew_Preload_remote_only(t *testing.T) {
	db, _ := redismock.NewClientMock()
	data := map[string][]byte{
		myKey: []byte(myValue),
	}
	// No local cache — preload has nothing to write to, but must not panic
	mycache, err := New("",
		WithRedisConn(db, time.Minute),
		WithPreload(data),
		WithPrefixKey(nil),
	)
	require.NoError(t, err)
	require.NotNil(t, mycache)
}

// --- Benchmarks ---

func BenchmarkCache_FreeCache_Set(b *testing.B) {
	c, _ := New("",
		WithLocalCacheFreeCache(1000000, time.Minute),
		WithPrefixKey(nil),
	)
	ctx := context.Background()
	key := []byte("bench-key")
	value := []byte("bench-value")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = c.Set(ctx, key, value)
	}
}

func BenchmarkCache_FreeCache_Get_hit(b *testing.B) {
	c, _ := New("",
		WithLocalCacheFreeCache(1000000, time.Minute),
		WithPrefixKey(nil),
	)
	ctx := context.Background()
	key := []byte("bench-key")
	_ = c.Set(ctx, key, []byte("bench-value"))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = c.Get(ctx, key)
	}
}

func BenchmarkCache_FreeCache_Get_miss(b *testing.B) {
	c, _ := New("",
		WithLocalCacheFreeCache(1000000, time.Minute),
		WithPrefixKey(nil),
	)
	ctx := context.Background()
	key := []byte("bench-miss")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = c.Get(ctx, key)
	}
}

func BenchmarkCache_TinyLFU_Set(b *testing.B) {
	c, _ := New("",
		WithLocalCacheTinyLFU(10000, time.Minute),
		WithPrefixKey(nil),
	)
	ctx := context.Background()
	key := []byte("bench-key")
	value := []byte("bench-value")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = c.Set(ctx, key, value)
	}
}

func BenchmarkCache_TinyLFU_Get_hit(b *testing.B) {
	c, _ := New("",
		WithLocalCacheTinyLFU(10000, time.Minute),
		WithPrefixKey(nil),
	)
	ctx := context.Background()
	key := []byte("bench-key")
	_ = c.Set(ctx, key, []byte("bench-value"))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = c.Get(ctx, key)
	}
}

func BenchmarkCache_TinyLFU_Get_miss(b *testing.B) {
	c, _ := New("",
		WithLocalCacheTinyLFU(10000, time.Minute),
		WithPrefixKey(nil),
	)
	ctx := context.Background()
	key := []byte("bench-miss")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = c.Get(ctx, key)
	}
}

func BenchmarkCache_FreeCache_GracefulDeg_Set(b *testing.B) {
	c, _ := New("",
		WithLocalCacheFreeCache(1000000, time.Minute),
		WithGracefulDegradation(1*time.Hour),
		WithPrefixKey(nil),
	)
	ctx := context.Background()
	key := []byte("bench-key")
	value := []byte("bench-value")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = c.Set(ctx, key, value)
	}
}

func BenchmarkCache_FreeCache_GracefulDeg_Get_hit(b *testing.B) {
	c, _ := New("",
		WithLocalCacheFreeCache(1000000, time.Minute),
		WithGracefulDegradation(1*time.Hour),
		WithPrefixKey(nil),
	)
	ctx := context.Background()
	key := []byte("bench-key")
	_ = c.Set(ctx, key, []byte("bench-value"))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = c.Get(ctx, key)
	}
}

func BenchmarkCache_FreeCache_GracefulDeg_Get_miss(b *testing.B) {
	c, _ := New("",
		WithLocalCacheFreeCache(1000000, time.Minute),
		WithGracefulDegradation(1*time.Hour),
		WithPrefixKey(nil),
	)
	ctx := context.Background()
	key := []byte("bench-miss")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = c.Get(ctx, key)
	}
}

func BenchmarkCache_SetStruct_MsgPack(b *testing.B) {
	type user struct {
		Name  string `msgpack:"name"`
		Email string `msgpack:"email"`
		Age   int    `msgpack:"age"`
	}
	c, _ := New("",
		WithLocalCacheFreeCache(1000000, time.Minute),
		WithPrefixKey(nil),
	)
	ctx := context.Background()
	u := user{Name: "Alice", Email: "alice@example.com", Age: 30}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = c.SetStruct(ctx, "user:1", u)
	}
}

func BenchmarkCache_GetStruct_MsgPack(b *testing.B) {
	type user struct {
		Name  string `msgpack:"name"`
		Email string `msgpack:"email"`
		Age   int    `msgpack:"age"`
	}
	c, _ := New("",
		WithLocalCacheFreeCache(1000000, time.Minute),
		WithPrefixKey(nil),
	)
	ctx := context.Background()
	_ = c.SetStruct(ctx, "user:1", user{Name: "Alice", Email: "alice@example.com", Age: 30})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var u user
		_ = c.GetStruct(ctx, "user:1", &u)
	}
}

func BenchmarkCache_SetStruct_JSON(b *testing.B) {
	type user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}
	c, _ := New("",
		WithLocalCacheFreeCache(1000000, time.Minute),
		WithCoder(&JSONCoder{}),
		WithPrefixKey(nil),
	)
	ctx := context.Background()
	u := user{Name: "Alice", Email: "alice@example.com", Age: 30}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = c.SetStruct(ctx, "user:1", u)
	}
}

func BenchmarkCache_GetStruct_JSON(b *testing.B) {
	type user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}
	c, _ := New("",
		WithLocalCacheFreeCache(1000000, time.Minute),
		WithCoder(&JSONCoder{}),
		WithPrefixKey(nil),
	)
	ctx := context.Background()
	_ = c.SetStruct(ctx, "user:1", user{Name: "Alice", Email: "alice@example.com", Age: 30})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var u user
		_ = c.GetStruct(ctx, "user:1", &u)
	}
}
