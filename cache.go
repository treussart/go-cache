// Package cache provides a two-level cache (local L1 + Redis L2) with
// circuit breaker, Prometheus/OpenTelemetry metrics, and structured value support.
package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	pkgName                      = "cache"
	defaultCBTimeout             = 4 * time.Minute
	defaultCBConsecutiveFailures = 2
	defaultCBMaxRequests         = 1
)

// Cacher defines the interface for a two-level cache (local + Redis).
type Cacher interface {
	Del(ctx context.Context, key []byte) error
	Set(ctx context.Context, key, value []byte) error
	SetExp(ctx context.Context, key, value []byte, ttl time.Duration) error
	Get(ctx context.Context, key []byte) ([]byte, error)
	DeleteFromLocalCache(key []byte)
	DeleteFromRemoteCache(ctx context.Context, key []byte) error
	Ready(ctx context.Context) error
	GetStruct(context.Context, string, any) error
	SetStruct(context.Context, string, any) error
	SetExStruct(context.Context, string, any, time.Duration) error
}

// Cache implements Cacher with optional local (L1) and Redis (L2) layers.
type Cache struct {
	opt        *customConfig
	labelValue string
	cb         *gobreaker.CircuitBreaker[[]byte]
	tracer     trace.Tracer
}

// New creates a new Cache. At least one of Redis or LocalCache must be set.
func New(name string, options ...CustomOption) (*Cache, error) {
	defaults := []CustomOption{
		WithCBEnabled(false),
		WithCBTimeout(defaultCBTimeout),
		WithCBMaxRequests(defaultCBMaxRequests),
		WithCBConsecutiveFailures(defaultCBConsecutiveFailures),
		WithPrefixKey([]byte(name + ":")),
		WithCoder(&MsgPackCoder{}),
	}
	var config customConfig
	for _, opt := range append(defaults, options...) {
		opt(&config)
	}

	if config.remoteCache == nil && config.localCache == nil {
		return nil, ErrInitCache
	}

	cbConf := gobreaker.Settings{
		Name:        "Redis Cache Circuit Breaker",
		Timeout:     config.cbTimeout,
		MaxRequests: config.cbMaxRequests,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= config.cbConsecutiveFailures
		},
		IsSuccessful: func(err error) bool {
			return err == nil || errors.Is(err, redis.Nil)
		},
		OnStateChange: func(_ string, _ gobreaker.State, to gobreaker.State) {
			if config.statsProm != nil && config.statsProm.CBState != nil {
				config.statsProm.CBState.WithLabelValues(name).Set(float64(to))
			}
			if config.statsOTEL != nil && config.statsOTEL.CBState != nil {
				config.statsOTEL.CBState.Record(context.Background(), float64(to),
					metric.WithAttributes(attribute.String(LabelName, name)))
			}
		},
	}

	return &Cache{
		opt:        &config,
		labelValue: name,
		cb:         gobreaker.NewCircuitBreaker[[]byte](cbConf),
		tracer:     otel.GetTracerProvider().Tracer(pkgName),
	}, nil
}

var _ Cacher = &Cache{}

// opHandle holds the state needed to finish an instrumented operation.
// It is returned by value from startOp so it stays on the stack.
type opHandle struct {
	span    trace.Span
	timer   *prometheus.Timer
	start   time.Time
	hasOTEL bool
	cache   *Cache
	cmd     string
	ctx     context.Context
}

// End finishes the span and records duration metrics. Must be deferred.
func (h opHandle) End() {
	h.span.End()
	if h.timer != nil {
		h.timer.ObserveDuration()
	}
	if h.hasOTEL {
		h.cache.opt.statsOTEL.Duration.Record(h.ctx, time.Since(h.start).Seconds(), metric.WithAttributes(
			attribute.String(LabelName, h.cache.labelValue), attribute.String(labelCmd, h.cmd),
		))
	}
}

// startOp begins an instrumented operation: it creates a trace span, starts
// a Prometheus timer (if configured), and captures the start time for the
// OTEL duration histogram.
func (c *Cache) startOp(ctx context.Context, spanName, cmd string) (context.Context, opHandle) {
	var h opHandle
	h.cache = c
	h.cmd = cmd
	if c.opt.statsProm != nil && c.opt.statsProm.Duration != nil {
		h.timer = prometheus.NewTimer(c.opt.statsProm.Duration.WithLabelValues(c.labelValue, cmd))
	}
	h.hasOTEL = c.opt.statsOTEL != nil && c.opt.statsOTEL.Duration != nil
	if h.hasOTEL {
		h.start = time.Now()
	}
	ctx, h.span = c.tracer.Start(ctx, spanName) //nolint:spancheck // span is closed by opHandle.End()
	h.ctx = ctx
	return ctx, h
}

// Del removes a key from both local and remote caches.
func (c *Cache) Del(ctx context.Context, key []byte) error {
	c.DeleteFromLocalCache(key)
	if err := c.DeleteFromRemoteCache(ctx, key); err != nil {
		return fmt.Errorf("DeleteFromRemoteCache: %w", err)
	}
	return nil
}

// DeleteFromLocalCache removes a key from the local cache only.
func (c *Cache) DeleteFromLocalCache(key []byte) {
	if c.opt.localCache != nil {
		c.opt.localCache.Del(c.addPrefixIfExist(key))
	}
}

// DeleteFromRemoteCache removes a key from the Redis remote cache only.
func (c *Cache) DeleteFromRemoteCache(ctx context.Context, key []byte) error {
	if len(key) == 0 {
		return ErrKeyEmpty
	}
	if c.opt.remoteCache != nil {
		err := c.opt.remoteCache.Del(ctx, string(c.addPrefixIfExist(key))).Err()
		if err != nil {
			return fmt.Errorf("remoteCache.Del: %w", err)
		}
	}
	return nil
}

// doSet is the unified private setter. A zero ttl means "use defaults"
// (local cache's built-in TTL and the configured remote TTL).
func (c *Cache) doSet(ctx context.Context, key, value []byte, ttl time.Duration) error {
	pkey := c.addPrefixIfExist(key)
	span := trace.SpanFromContext(ctx)

	if c.opt.localCache != nil {
		var err error
		if ttl > 0 {
			err = c.opt.localCache.SetExp(pkey, value, ttl)
		} else {
			err = c.opt.localCache.Set(pkey, value)
		}
		if err != nil {
			span.SetStatus(codes.Error, "localCache.Set")
			span.RecordError(err)
			return fmt.Errorf("localCache.Set: %w", err)
		}
		if c.opt.statsProm != nil {
			c.opt.statsProm.SetLocal.WithLabelValues(c.labelValue).Inc()
		}
		if c.opt.statsOTEL != nil {
			c.opt.statsOTEL.SetLocal.Add(ctx, 1, metric.WithAttributes(attribute.String(LabelName, c.labelValue)))
		}
	}

	if c.opt.remoteCache != nil {
		remoteKey := string(pkey)
		var err error
		switch {
		case c.opt.cbEnabled:
			_, err = c.cb.Execute(func() ([]byte, error) {
				if ttl > 0 {
					return nil, c.opt.remoteCache.SetEx(ctx, remoteKey, value, ttl).Err()
				}
				return nil, c.opt.remoteCache.Set(ctx, remoteKey, value, c.opt.remoteCacheTTL).Err()
			})
		case ttl > 0:
			err = c.opt.remoteCache.SetEx(ctx, remoteKey, value, ttl).Err()
		default:
			err = c.opt.remoteCache.Set(ctx, remoteKey, value, c.opt.remoteCacheTTL).Err()
		}
		if err != nil {
			if c.handleCBError(ctx, err) {
				if span.IsRecording() {
					span.SetAttributes(attribute.Bool("cb.fallback", true))
				}
			} else {
				span.SetStatus(codes.Error, "remoteCache.Set")
				span.RecordError(err)
				return fmt.Errorf("remoteCache.Set: %w", err)
			}
		} else {
			if c.opt.statsProm != nil {
				c.opt.statsProm.SetRemote.WithLabelValues(c.labelValue).Inc()
			}
			if c.opt.statsOTEL != nil {
				c.opt.statsOTEL.SetRemote.Add(ctx, 1, metric.WithAttributes(attribute.String(LabelName, c.labelValue)))
			}
		}
	}

	if c.opt.statsProm != nil {
		c.opt.statsProm.SetTotal.WithLabelValues(c.labelValue).Inc()
	}
	if c.opt.statsOTEL != nil {
		c.opt.statsOTEL.SetTotal.Add(ctx, 1, metric.WithAttributes(attribute.String(LabelName, c.labelValue)))
	}
	return nil
}

// nolint:funlen,nolintlint
func (c *Cache) get(ctx context.Context, key []byte) ([]byte, error) {
	pkey := c.addPrefixIfExist(key)
	span := trace.SpanFromContext(ctx)

	if c.opt.localCache != nil {
		b, err := c.opt.localCache.Get(pkey)
		if err == nil {
			if c.opt.statsProm != nil {
				c.opt.statsProm.HitsLocal.WithLabelValues(c.labelValue).Inc()
			}
			if span.IsRecording() {
				span.SetAttributes(attribute.Bool("hit.local", true))
			}
			return b, nil
		}
		if c.opt.statsProm != nil {
			c.opt.statsProm.MissesLocal.WithLabelValues(c.labelValue).Inc()
		}
		if c.opt.statsOTEL != nil {
			c.opt.statsOTEL.MissesLocal.Add(ctx, 1, metric.WithAttributes(attribute.String(LabelName, c.labelValue)))
		}
		if span.IsRecording() {
			span.SetAttributes(attribute.Bool("hit.local", false))
		}
	}

	if c.opt.remoteCache == nil {
		return nil, ErrCacheMiss
	}

	var err error
	var b []byte
	remoteKey := string(pkey)
	if c.opt.cbEnabled {
		b, err = c.cb.Execute(func() ([]byte, error) {
			return c.opt.remoteCache.Get(ctx, remoteKey).Bytes()
		})
	} else {
		b, err = c.opt.remoteCache.Get(ctx, remoteKey).Bytes()
	}
	if err != nil {
		if c.handleCBError(ctx, err) {
			if span.IsRecording() {
				span.SetAttributes(attribute.Bool("cb.fallback", true))
			}
			return nil, fmt.Errorf("remoteCache.Get: %w", err)
		}
		if errors.Is(err, redis.Nil) {
			if c.opt.statsProm != nil {
				c.opt.statsProm.MissesRemote.WithLabelValues(c.labelValue).Inc()
			}
			if c.opt.statsOTEL != nil {
				c.opt.statsOTEL.MissesRemote.Add(ctx, 1, metric.WithAttributes(attribute.String(LabelName, c.labelValue)))
			}
			if span.IsRecording() {
				span.SetAttributes(attribute.Bool("hit.remote", false))
			}
			return nil, ErrCacheMiss
		}
		span.SetStatus(codes.Error, "remoteCache.Get")
		span.RecordError(err)
		return nil, fmt.Errorf("remoteCache.Get: %w", err)
	}

	if c.opt.localCache != nil {
		if err = c.opt.localCache.Set(pkey, b); err != nil {
			span.SetStatus(codes.Error, "localCache.Set")
			span.RecordError(err)
			return nil, fmt.Errorf("localCache.Set: %w", err)
		}
	}

	if c.opt.statsProm != nil {
		c.opt.statsProm.HitsRemote.WithLabelValues(c.labelValue).Inc()
	}
	if c.opt.statsOTEL != nil {
		c.opt.statsOTEL.HitsRemote.Add(ctx, 1, metric.WithAttributes(attribute.String(LabelName, c.labelValue)))
	}
	if span.IsRecording() {
		span.SetAttributes(attribute.Bool("hit.remote", true))
	}
	return b, nil
}

// Ready checks whether the Redis connection is healthy.
func (c *Cache) Ready(ctx context.Context) error {
	if c.opt.remoteCache != nil {
		if err := c.opt.remoteCache.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("remoteCache.Ping: %w", err)
		}
	}
	return nil
}

// handleCBError checks whether err is a circuit breaker rejection (open or
// half-open/too-many-requests). When it is, the appropriate metric is recorded
// and true is returned so the caller can fall back to L1-only behaviour.
func (c *Cache) handleCBError(ctx context.Context, err error) bool {
	if errors.Is(err, gobreaker.ErrOpenState) {
		if c.opt.statsProm != nil {
			c.opt.statsProm.CBOpen.WithLabelValues(c.labelValue).Inc()
		}
		if c.opt.statsOTEL != nil {
			c.opt.statsOTEL.CBOpen.Add(ctx, 1, metric.WithAttributes(attribute.String(LabelName, c.labelValue)))
		}
		return true
	}
	if errors.Is(err, gobreaker.ErrTooManyRequests) {
		if c.opt.statsProm != nil {
			c.opt.statsProm.CBTooManyRequests.WithLabelValues(c.labelValue).Inc()
		}
		if c.opt.statsOTEL != nil {
			c.opt.statsOTEL.CBTooManyRequests.Add(ctx, 1, metric.WithAttributes(attribute.String(LabelName, c.labelValue)))
		}
		return true
	}
	return false
}

func (c *Cache) addPrefixIfExist(key []byte) []byte {
	if c.opt.prefixKey == nil {
		return key
	}
	buf := make([]byte, len(c.opt.prefixKey)+len(key))
	copy(buf, c.opt.prefixKey)
	copy(buf[len(c.opt.prefixKey):], key)
	return buf
}

// Set stores a key-value pair in both local and remote caches using the configured TTL.
func (c *Cache) Set(ctx context.Context, key, value []byte) error {
	ctx, op := c.startOp(ctx, pkgName+".Set", "Set")
	defer op.End()
	if op.span.IsRecording() {
		op.span.SetAttributes(attribute.String("key", string(key)))
	}
	if len(key) == 0 {
		op.span.SetStatus(codes.Error, "Set")
		op.span.RecordError(ErrKeyEmpty)
		return ErrKeyEmpty
	}
	return c.doSet(ctx, key, value, 0)
}

// SetExp stores a key-value pair in both local and remote caches with a custom TTL.
func (c *Cache) SetExp(ctx context.Context, key, value []byte, ttl time.Duration) error {
	ctx, op := c.startOp(ctx, pkgName+".SetExp", "SetExp")
	defer op.End()
	if op.span.IsRecording() {
		op.span.SetAttributes(attribute.String("key", string(key)))
	}
	if len(key) == 0 {
		op.span.SetStatus(codes.Error, "SetExp")
		op.span.RecordError(ErrKeyEmpty)
		return ErrKeyEmpty
	}
	return c.doSet(ctx, key, value, ttl)
}

// Get retrieves a value by key, checking the local cache first, then falling back to Redis.
func (c *Cache) Get(ctx context.Context, key []byte) ([]byte, error) {
	ctx, op := c.startOp(ctx, pkgName+".Get", "Get")
	defer op.End()
	if op.span.IsRecording() {
		op.span.SetAttributes(attribute.String("key", string(key)))
	}
	if len(key) == 0 {
		op.span.SetStatus(codes.Error, "Get")
		op.span.RecordError(ErrKeyEmpty)
		return nil, ErrKeyEmpty
	}
	return c.get(ctx, key)
}

// GetStruct retrieves a value by key and decodes it into dest.
func (c *Cache) GetStruct(ctx context.Context, key string, dest any) error {
	ctx, op := c.startOp(ctx, pkgName+".GetStruct", "GetStruct")
	defer op.End()
	if op.span.IsRecording() {
		op.span.SetAttributes(attribute.String("key", key))
	}
	if len(key) == 0 {
		op.span.SetStatus(codes.Error, "GetStruct")
		op.span.RecordError(ErrKeyEmpty)
		return ErrKeyEmpty
	}
	b, err := c.get(ctx, []byte(key))
	if err != nil {
		return fmt.Errorf("cache.Get: %w", err)
	}
	if err = c.opt.coder.Decode(b, dest); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}

// SetStruct encodes value and stores it in the cache under the given key.
func (c *Cache) SetStruct(ctx context.Context, key string, value any) error {
	ctx, op := c.startOp(ctx, pkgName+".SetStruct", "SetStruct")
	defer op.End()
	if op.span.IsRecording() {
		op.span.SetAttributes(attribute.String("key", key))
	}
	if len(key) == 0 {
		op.span.SetStatus(codes.Error, "SetStruct")
		op.span.RecordError(ErrKeyEmpty)
		return ErrKeyEmpty
	}
	data, err := c.opt.coder.Encode(value)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	if err = c.doSet(ctx, []byte(key), data, 0); err != nil {
		return fmt.Errorf("cache.Set: %w", err)
	}
	return nil
}

// SetExStruct encodes value and stores it in the cache with a custom TTL.
func (c *Cache) SetExStruct(ctx context.Context, key string, value any, ttl time.Duration) error {
	ctx, op := c.startOp(ctx, pkgName+".SetExStruct", "SetExStruct")
	defer op.End()
	if op.span.IsRecording() {
		op.span.SetAttributes(attribute.String("key", key))
	}
	if len(key) == 0 {
		op.span.SetStatus(codes.Error, "SetExStruct")
		op.span.RecordError(ErrKeyEmpty)
		return ErrKeyEmpty
	}
	data, err := c.opt.coder.Encode(value)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	if err = c.doSet(ctx, []byte(key), data, ttl); err != nil {
		return fmt.Errorf("cache.Set: %w", err)
	}
	return nil
}
