# go-cache

Two-level cache (local L1 + Redis L2) with circuit breaker, Prometheus/OpenTelemetry metrics, and structured value support.

## Install

```bash
go get github.com/treussart/go-cache
```

## Architecture

```
Get request
    │
    ▼
┌────────┐  hit   ┌────────────┐
│ L1     │───────▶│ return     │
│ (local)│        └────────────┘
└───┬────┘
    │ miss
    ▼
┌────────┐  hit   ┌────────────┐
│ L2     │───────▶│ write to L1│──▶ return
│ (Redis)│        └────────────┘
└───┬────┘
    │ miss
    ▼
ErrCacheMiss
```

Either layer is optional — you can use L1-only, L2-only, or both.

## Usage

Cache instances are created with functional options:

```go
import (
    cache "github.com/treussart/go-cache"
)

c, err := cache.New("my-service",
    cache.WithRedisConn(redisClient, 5*time.Minute),
    cache.WithLocalCacheTinyLFU(10000, time.Minute),
    cache.WithCBEnabled(true),
)
```

### Raw bytes

```go
err = c.Set(ctx, []byte("user:1"), []byte("data"))

val, err := c.Get(ctx, []byte("user:1"))

err = c.SetExp(ctx, []byte("user:1"), []byte("data"), 30*time.Second)

err = c.Del(ctx, []byte("user:1"))
```

### Structured values

`GetStruct` / `SetStruct` / `SetExStruct` encode and decode values automatically
using the configured `Coder` (MsgPack by default):

```go
type User struct {
    Name string
    Age  int
}

err = c.SetStruct(ctx, "user:1", &User{Name: "Alice", Age: 30})

var u User
err = c.GetStruct(ctx, "user:1", &u)

err = c.SetExStruct(ctx, "user:1", &User{Name: "Alice"}, 30*time.Second)
```

## Functional options

| Option | Description |
|--------|-------------|
| `WithRedisConn(conn, ttl)` | Set the Redis connection (`redis.UniversalClient`) and default TTL |
| `WithLocalCacheTinyLFU(size, ttl)` | L1 using TinyLFU eviction (size = max items, default 10 000) |
| `WithLocalCacheFreeCache(size, ttl)` | L1 using FreeCache (size = bytes, default 1 MB) |
| `WithPrefixKey(prefix)` | Override key namespace prefix (default: `name + ":"`) |
| `WithCoder(coder)` | Serializer for `*Struct` methods (default: `MsgPackCoder`) |
| `WithStatsProm(stats)` | Attach Prometheus counters from `GetStatsProm()` |
| `WithStatsOTEL(stats)` | Attach OpenTelemetry counters from `GetStatsOTEL()` |
| `WithCBEnabled(bool)` | Enable circuit breaker on Redis calls |
| `WithCBTimeout(d)` | Time in open state before half-open probe (default: 4 min) |
| `WithCBMaxRequests(n)` | Max requests allowed in half-open state (default: 1) |
| `WithCBConsecutiveFailures(n)` | Consecutive failures before tripping (default: 2) |
| `WithGracefulDegradation(staleTTL)` | Enable stale cache fallback when CB is open (see below) |

## Interface

```go
type Cacher interface {
    Get(ctx context.Context, key []byte) ([]byte, error)
    Set(ctx context.Context, key, value []byte) error
    SetExp(ctx context.Context, key, value []byte, ttl time.Duration) error
    Del(ctx context.Context, key []byte) error
    DeleteFromLocalCache(key []byte)
    DeleteFromRemoteCache(ctx context.Context, key []byte) error
    Ready(ctx context.Context) error
    GetStruct(ctx context.Context, key string, dest any) error
    SetStruct(ctx context.Context, key string, value any) error
    SetExStruct(ctx context.Context, key string, value any, ttl time.Duration) error
}
```

A `Mocked` struct (testify mock) implementing `Cacher` is provided for unit tests.

## Circuit breaker

When enabled, all Redis operations go through a [gobreaker](https://github.com/sony/gobreaker) circuit breaker:

- **Closed** — requests flow normally to Redis.
- **Open** — Redis calls are skipped. `Get` returns the error; `Set`/`SetExp`/`Del` silently fall back to L1.
- **Half-open** — a limited number of probe requests are sent to Redis.

`redis.Nil` (cache miss) is treated as a success and does not count toward tripping.

## Graceful degradation

When Redis goes down and the circuit breaker opens, `Get` normally returns an error for any key that is no longer in L1 (expired or evicted). With graceful degradation enabled, a **stale cache** (a separate in-memory TinyLFU with a longer TTL) is consulted before returning an error:

```
CB open + L1 miss
    │
    ▼
┌────────────┐  hit   ┌─────────────────┐
│ Stale cache│───────▶│ return stale data│
└───┬────────┘        └─────────────────┘
    │ miss
    ▼
return error
```

Enable it with:

```go
c, err := cache.New("my-service",
    cache.WithRedisConn(redisClient, 5*time.Minute),
    cache.WithLocalCacheTinyLFU(10000, time.Minute),
    cache.WithCBEnabled(true),
    cache.WithGracefulDegradation(1*time.Hour), // stale TTL
)
```

- The stale cache is written on every `Set`/`SetExp`/`SetStruct`/`SetExStruct`, not only during degradation.
- Only circuit breaker errors (open / too-many-requests) trigger a stale lookup. Regular Redis errors are surfaced normally.
- `Del` clears both the primary L1 and the stale cache to maintain consistency.
- A `cache_stale_hit_total` metric is emitted on each stale hit (both Prometheus and OpenTelemetry).

## Coders

The `Coder` interface controls serialization for `GetStruct`/`SetStruct`/`SetExStruct`:

| Implementation | Format |
|----------------|--------|
| `MsgPackCoder` | MessagePack (default) |
| `JSONCoder` | JSON |

Implement the `Coder` interface for custom serialization:

```go
type Coder interface {
    Encode(value any) ([]byte, error)
    Decode(data []byte, value any) error
}
```

## Observability

### Prometheus

```go
c, _ := cache.New("my-cache",
    cache.WithStatsProm(cache.GetStatsProm("namespace", "subsystem")),
    // ...
)
```

### OpenTelemetry

```go
stats, _ := cache.GetStatsOTEL("my-cache")
c, _ := cache.New("my-cache",
    cache.WithStatsOTEL(stats),
    // ...
)
```

Exposed metrics: `cache_local_hit_total`, `cache_local_miss_total`, `cache_remote_hit_total`, `cache_remote_miss_total`, `cache_stale_hit_total`, `cache_local_set_total`, `cache_remote_set_total`, `cache_set_total`, `cache_cb_open_total`, `cache_cb_too_many_requests_total`, `cache_cb_state`, `cache_duration_seconds`.

All metrics are labelled with `cache_name`.

## Benchmarks

Run with `go test -bench=. -benchmem ./...`

**Apple M1 Pro — Go 1.26 — arm64**

### Coder (serialization only)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|------:|-----:|----------:|
| MsgPackCoder Encode | 503 | 608 | 5 |
| MsgPackCoder Decode | 712 | 312 | 9 |
| JSONCoder Encode ([goccy/go-json](https://github.com/goccy/go-json)) | 281 | 320 | 2 |
| JSONCoder Decode ([goccy/go-json](https://github.com/goccy/go-json)) | 453 | 416 | 6 |

### Cache operations (L1 only, raw bytes)

| Benchmark | ns/op | B/op | allocs/op | Δ ns | Δ allocs |
|-----------|------:|-----:|----------:|-----:|---------:|
| FreeCache Set | 206 | 240 | 3 | −26 % | −40 % |
| FreeCache Get (hit) | 217 | 256 | 4 | −38 % | −43 % |
| FreeCache Get (miss) | 257 | 304 | 5 | −36 % | −38 % |
| TinyLFU Set | 434 | 440 | 7 | −22 % | −22 % |
| TinyLFU Get (hit) | 216 | 240 | 3 | −38 % | −50 % |
| TinyLFU Get (miss) | 155 | 240 | 3 | −46 % | −50 % |

	### Cache operations with graceful degradation (L1 + stale cache, raw bytes)

| Benchmark | ns/op | B/op | allocs/op | vs. without GD |
|-----------|------:|-----:|----------:|---------------:|
| FreeCache Set | 521 | 440 | 7 | +4 allocs (stale TinyLFU write) |
| FreeCache Get (hit) | 260 | 256 | 4 | 0 allocs overhead |
| FreeCache Get (miss) | 288 | 304 | 5 | 0 allocs overhead |

Set pays for the extra TinyLFU write into the stale cache. Get paths are unaffected because the stale cache is only consulted when the circuit breaker rejects a request.

### Cache operations (L1 only, struct roundtrip)

| Benchmark | ns/op | B/op | allocs/op | Δ ns | Δ allocs |
|-----------|------:|-----:|----------:|-----:|---------:|
| SetStruct MsgPack | 403 | 360 | 6 | −42 % | −50 % |
| GetStruct MsgPack | 516 | 424 | 9 | −41 % | −44 % |
| SetStruct JSON | 312 | 312 | 5 | −48 % | −55 % |
| GetStruct JSON | 390 | 424 | 7 | −45 % | −50 % |
