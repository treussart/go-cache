package cache

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const (
	// LabelName is the Prometheus/OTEL label key used to identify the cache instance.
	LabelName = "cache_name"
	labelCmd  = "cmd"
)

// StatsProm contains accumulated stats.
type StatsProm struct {
	HitsLocal         *prometheus.CounterVec
	MissesLocal       *prometheus.CounterVec
	HitsRemote        *prometheus.CounterVec
	MissesRemote      *prometheus.CounterVec
	HitsStale         *prometheus.CounterVec
	SetLocal          *prometheus.CounterVec
	SetRemote         *prometheus.CounterVec
	SetTotal          *prometheus.CounterVec
	CBOpen            *prometheus.CounterVec
	CBTooManyRequests *prometheus.CounterVec
	CBState           *prometheus.GaugeVec
	Duration          *prometheus.HistogramVec
}

// GetStatsProm creates a new StatsProm instance with Prometheus counter vectors for the given namespace and subsystem.
func GetStatsProm(metricNamespace, metricSubSystem string) *StatsProm {
	return &StatsProm{
		HitsLocal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: metricNamespace,
				Subsystem: metricSubSystem,
				Name:      "cache_local_hit_total",
				Help:      "Number of times a key was found in the local cache",
			},
			[]string{LabelName},
		),
		MissesLocal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: metricNamespace,
				Subsystem: metricSubSystem,
				Name:      "cache_local_miss_total",
				Help:      "Number of times a key was not found in the local cache",
			},
			[]string{LabelName},
		),
		HitsRemote: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: metricNamespace,
				Subsystem: metricSubSystem,
				Name:      "cache_remote_hit_total",
				Help:      "Number of times a key was found in the remote cache",
			},
			[]string{LabelName},
		),
		MissesRemote: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: metricNamespace,
				Subsystem: metricSubSystem,
				Name:      "cache_remote_miss_total",
				Help:      "Number of times a key was not found in the remote cache",
			},
			[]string{LabelName},
		),
		HitsStale: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: metricNamespace,
				Subsystem: metricSubSystem,
				Name:      "cache_stale_hit_total",
				Help:      "Number of times a stale cache hit saved a request during circuit breaker open state",
			},
			[]string{LabelName},
		),
		SetLocal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: metricNamespace,
				Subsystem: metricSubSystem,
				Name:      "cache_local_set_total",
				Help:      "Number of times a key was set in the local cache",
			},
			[]string{LabelName},
		),
		SetRemote: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: metricNamespace,
				Subsystem: metricSubSystem,
				Name:      "cache_remote_set_total",
				Help:      "Number of times a key was set in the remote cache",
			},
			[]string{LabelName},
		),
		SetTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: metricNamespace,
				Subsystem: metricSubSystem,
				Name:      "cache_set_total",
				Help:      "Number of times a key was set in the cache",
			},
			[]string{LabelName},
		),
		CBOpen: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: metricNamespace,
				Subsystem: metricSubSystem,
				Name:      "cache_cb_open_total",
				Help:      "Number of times the cache circuit breaker rejected a request (open state)",
			},
			[]string{LabelName},
		),
		CBTooManyRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: metricNamespace,
				Subsystem: metricSubSystem,
				Name:      "cache_cb_too_many_requests_total",
				Help:      "Number of times the cache circuit breaker rejected a request (too many requests in half-open)",
			},
			[]string{LabelName},
		),
		CBState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: metricNamespace,
				Subsystem: metricSubSystem,
				Name:      "cache_cb_state",
				Help:      "Current state of the cache circuit breaker (0=closed, 1=half-open, 2=open)",
			},
			[]string{LabelName},
		),
		Duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: metricNamespace,
				Subsystem: metricSubSystem,
				Name:      "cache_duration_seconds",
				Help:      "Duration of cache",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{LabelName, labelCmd},
		),
	}
}

// StatsOTEL holds OpenTelemetry counters for cache hit, miss, and set operations.
type StatsOTEL struct {
	HitsLocal         metric.Float64Counter
	MissesLocal       metric.Float64Counter
	HitsRemote        metric.Float64Counter
	MissesRemote      metric.Float64Counter
	HitsStale         metric.Float64Counter
	SetLocal          metric.Float64Counter
	SetRemote         metric.Float64Counter
	SetTotal          metric.Float64Counter
	CBOpen            metric.Float64Counter
	CBTooManyRequests metric.Float64Counter
	CBState           metric.Float64Gauge
	Duration          metric.Float64Histogram
}

// GetStatsOTEL creates a new StatsOTEL instance with OpenTelemetry counters for the given meter name.
func GetStatsOTEL(name string) (*StatsOTEL, error) {
	meter := otel.GetMeterProvider().Meter(name)
	hitsLocal, err := meter.Float64Counter("cache_local_hit_total",
		metric.WithDescription("Number of times a key was found in the local cache."),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Counter: %w", err)
	}
	missesLocal, err := meter.Float64Counter("cache_local_miss_total",
		metric.WithDescription("Number of times a key was not found in the local cache."),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Counter: %w", err)
	}
	hitsRemote, err := meter.Float64Counter("cache_remote_hit_total",
		metric.WithDescription("Number of times a key was found in the remote cache."),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Counter: %w", err)
	}
	missesRemote, err := meter.Float64Counter("cache_remote_miss_total",
		metric.WithDescription("Number of times a key was not found in the remote cache."),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Counter: %w", err)
	}
	hitsStale, err := meter.Float64Counter("cache_stale_hit_total",
		metric.WithDescription("Number of times a stale cache hit saved a request during circuit breaker open state."),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Counter: %w", err)
	}
	setLocal, err := meter.Float64Counter("cache_local_set_total",
		metric.WithDescription("Number of times a key was set in the local cache"),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Counter: %w", err)
	}
	setRemote, err := meter.Float64Counter("cache_remote_set_total",
		metric.WithDescription("Number of times a key was set in the remote cache."),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Counter: %w", err)
	}
	setTotal, err := meter.Float64Counter("cache_set_total",
		metric.WithDescription("Number of times a key was set in the cache."),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Counter: %w", err)
	}

	cbOpen, err := meter.Float64Counter("cache_cb_open_total",
		metric.WithDescription("Number of times the cache circuit breaker rejected a request (open state)."),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Counter: %w", err)
	}
	cbTooManyRequests, err := meter.Float64Counter("cache_cb_too_many_requests_total",
		metric.WithDescription("Number of times the cache circuit breaker rejected a request (too many requests in half-open)."),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Counter: %w", err)
	}
	cbState, err := meter.Float64Gauge("cache_cb_state",
		metric.WithDescription("Current state of the cache circuit breaker (0=closed, 1=half-open, 2=open)."),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Gauge: %w", err)
	}

	duration, err := meter.Float64Histogram(
		"cache_duration_seconds",
		metric.WithDescription("The duration in seconds of the event."),
		metric.WithExplicitBucketBoundaries(.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10),
		metric.WithUnit("seconds"),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Histogram: %w", err)
	}

	return &StatsOTEL{
		HitsLocal:         hitsLocal,
		MissesLocal:       missesLocal,
		HitsRemote:        hitsRemote,
		MissesRemote:      missesRemote,
		HitsStale:         hitsStale,
		SetLocal:          setLocal,
		SetRemote:         setRemote,
		SetTotal:          setTotal,
		CBOpen:            cbOpen,
		CBTooManyRequests: cbTooManyRequests,
		CBState:           cbState,
		Duration:          duration,
	}, nil
}
