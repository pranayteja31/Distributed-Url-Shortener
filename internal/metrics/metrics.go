package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	CacheHits = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "url_shortener_cache_hits_total",
			Help: "Total number of Redis cache hits.",
		},
	)

	CacheMisses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "url_shortener_cache_misses_total",
			Help: "Total number of Redis cache misses.",
		},
	)

	RedisGetLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "url_shortener_redis_get_latency_seconds",
			Help:    "Redis GET latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
	)

	PostgresLookupLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "url_shortener_postgres_lookup_latency_seconds",
			Help:    "PostgreSQL lookup latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
	)

	RequestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "url_shortener_http_request_latency_seconds",
			Help:    "Overall HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_shortener_http_requests_total",
			Help: "Total HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)
)

func Register() {
	prometheus.MustRegister(
		CacheHits,
		CacheMisses,
		RedisGetLatency,
		PostgresLookupLatency,
		RequestLatency,
		RequestsTotal,
	)
}

func ObserveRedisLatency(start time.Time) {
	RedisGetLatency.Observe(time.Since(start).Seconds())
}

func ObservePostgresLatency(start time.Time) {
	PostgresLookupLatency.Observe(time.Since(start).Seconds())
}

func ObserveRequest(start time.Time, method string, path string, status int) {
	statusCode := strconv.Itoa(status)

	RequestLatency.
		WithLabelValues(method, path, statusCode).
		Observe(time.Since(start).Seconds())

	RequestsTotal.
		WithLabelValues(method, path, statusCode).
		Inc()
}