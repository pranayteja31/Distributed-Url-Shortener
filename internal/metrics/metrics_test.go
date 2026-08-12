package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

// ============================================================
// CACHE HITS
// ============================================================

func TestCacheHits(t *testing.T) {
	before := testutil.ToFloat64(CacheHits)

	CacheHits.Inc()

	after := testutil.ToFloat64(CacheHits)

	if after != before+1 {
		t.Fatalf(
			"expected CacheHits to increase by 1, before=%v after=%v",
			before,
			after,
		)
	}
}

// ============================================================
// CACHE MISSES
// ============================================================

func TestCacheMisses(t *testing.T) {
	before := testutil.ToFloat64(CacheMisses)

	CacheMisses.Inc()

	after := testutil.ToFloat64(CacheMisses)

	if after != before+1 {
		t.Fatalf(
			"expected CacheMisses to increase by 1, before=%v after=%v",
			before,
			after,
		)
	}
}

// ============================================================
// REDIS LATENCY
// ============================================================

func TestObserveRedisLatency(t *testing.T) {
	start := time.Now()

	time.Sleep(1 * time.Millisecond)

	ObserveRedisLatency(start)

	// Histograms cannot be checked using ToFloat64().
	// CollectAndCount verifies that the histogram can be
	// collected successfully.
	count := testutil.CollectAndCount(
		RedisGetLatency,
	)

	if count != 1 {
		t.Fatalf(
			"expected Redis latency metric to be collected once, got %d",
			count,
		)
	}
}

// ============================================================
// POSTGRES LATENCY
// ============================================================

func TestObservePostgresLatency(t *testing.T) {
	start := time.Now()

	time.Sleep(1 * time.Millisecond)

	ObservePostgresLatency(start)

	count := testutil.CollectAndCount(
		PostgresLookupLatency,
	)

	if count != 1 {
		t.Fatalf(
			"expected PostgreSQL latency metric to be collected once, got %d",
			count,
		)
	}
}

// ============================================================
// REQUEST LATENCY
// ============================================================

func TestObserveRequestLatency(t *testing.T) {
	start := time.Now()

	time.Sleep(1 * time.Millisecond)

	ObserveRequest(
		start,
		"GET",
		"/r/abc123",
		302,
	)

	count := testutil.CollectAndCount(
		RequestLatency,
	)

	if count != 1 {
		t.Fatalf(
			"expected request latency metric to be collected once, got %d",
			count,
		)
	}
}

// ============================================================
// REQUEST TOTAL
// ============================================================

func TestObserveRequestTotal(t *testing.T) {
	method := "GET"
	path := "/test-metrics"
	status := "200"

	before := testutil.ToFloat64(
		RequestsTotal.WithLabelValues(
			method,
			path,
			status,
		),
	)

	ObserveRequest(
		time.Now(),
		method,
		path,
		200,
	)

	after := testutil.ToFloat64(
		RequestsTotal.WithLabelValues(
			method,
			path,
			status,
		),
	)

	if after != before+1 {
		t.Fatalf(
			"expected request counter to increase by 1, before=%v after=%v",
			before,
			after,
		)
	}
}

// ============================================================
// REQUEST LABELS
// ============================================================

func TestObserveRequestLabels(t *testing.T) {
	method := "POST"
	path := "/api/urls"
	status := "201"

	ObserveRequest(
		time.Now(),
		method,
		path,
		201,
	)

	value := testutil.ToFloat64(
		RequestsTotal.WithLabelValues(
			method,
			path,
			status,
		),
	)

	if value <= 0 {
		t.Fatal(
			"expected request counter for specified labels to be greater than zero",
		)
	}
}

// ============================================================
// DIFFERENT STATUS LABELS
// ============================================================

func TestObserveRequestDifferentStatuses(t *testing.T) {
	method := "GET"
	path := "/api/test"

	ObserveRequest(
		time.Now(),
		method,
		path,
		200,
	)

	ObserveRequest(
		time.Now(),
		method,
		path,
		404,
	)

	okCount := testutil.ToFloat64(
		RequestsTotal.WithLabelValues(
			method,
			path,
			"200",
		),
	)

	notFoundCount := testutil.ToFloat64(
		RequestsTotal.WithLabelValues(
			method,
			path,
			"404",
		),
	)

	if okCount <= 0 {
		t.Fatal("expected 200 status metric")
	}

	if notFoundCount <= 0 {
		t.Fatal("expected 404 status metric")
	}
}
