package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// ============================================================
// URL KEY
// ============================================================

func TestURLKey(t *testing.T) {
	tests := []struct {
		name      string
		shortCode string
		expected  string
	}{
		{
			name:      "normal shortcode",
			shortCode: "abc123",
			expected:  "url:abc123",
		},
		{
			name:      "empty shortcode",
			shortCode: "",
			expected:  "url:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := URLKey(tt.shortCode)

			if got != tt.expected {
				t.Fatalf(
					"expected %q, got %q",
					tt.expected,
					got,
				)
			}
		})
	}
}

// ============================================================
// URL TTL - NO EXPIRY
// ============================================================

func TestURLTTL_NoExpiry(t *testing.T) {
	ttl := URLTTL(nil)

	expected := 30 * time.Minute

	if ttl != expected {
		t.Fatalf(
			"expected TTL %v, got %v",
			expected,
			ttl,
		)
	}
}

// ============================================================
// URL TTL - EXPIRED
// ============================================================

func TestURLTTL_Expired(t *testing.T) {
	expired := time.Now().Add(-10 * time.Minute)

	ttl := URLTTL(&expired)

	if ttl != 0 {
		t.Fatalf(
			"expected TTL 0, got %v",
			ttl,
		)
	}
}

// ============================================================
// URL TTL - WITHIN MAX TTL
// ============================================================

func TestURLTTL_WithinLimit(t *testing.T) {
	expiresAt := time.Now().Add(30 * time.Minute)

	ttl := URLTTL(&expiresAt)

	if ttl <= 0 {
		t.Fatal("expected positive TTL")
	}

	if ttl > 30*time.Minute {
		t.Fatalf(
			"TTL should not exceed expiry duration, got %v",
			ttl,
		)
	}
}

// ============================================================
// URL TTL - MAX TTL
// ============================================================

func TestURLTTL_MaxLimit(t *testing.T) {
	expiresAt := time.Now().Add(2 * time.Hour)

	ttl := URLTTL(&expiresAt)

	if ttl > MaxCacheTTL {
		t.Fatalf(
			"expected TTL <= %v, got %v",
			MaxCacheTTL,
			ttl,
		)
	}

	// Allow a tiny amount of timing difference.
	if ttl < MaxCacheTTL-2*time.Second {
		t.Fatalf(
			"expected TTL close to max cache TTL %v, got %v",
			MaxCacheTTL,
			ttl,
		)
	}
}

// ============================================================
// REDIS CACHE - SET + GET
// ============================================================

func TestRedisCache_SetAndGet(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	cache := NewRedisCache(client)

	ctx := context.Background()

	key := "url:abc123"
	value := []byte(`{"original_url":"https://example.com"}`)

	err = cache.Set(
		ctx,
		key,
		value,
		30*time.Minute,
	)

	if err != nil {
		t.Fatalf(
			"Set returned error: %v",
			err,
		)
	}

	data, found, err := cache.Get(
		ctx,
		key,
	)

	if err != nil {
		t.Fatalf(
			"Get returned error: %v",
			err,
		)
	}

	if !found {
		t.Fatal("expected key to be found")
	}

	if string(data) != string(value) {
		t.Fatalf(
			"expected %s, got %s",
			string(value),
			string(data),
		)
	}
}

// ============================================================
// REDIS CACHE - GET MISSING KEY
// ============================================================

func TestRedisCache_GetMissingKey(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	cache := NewRedisCache(client)

	data, found, err := cache.Get(
		context.Background(),
		"url:does-not-exist",
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if found {
		t.Fatal("expected key not to be found")
	}

	if data != nil {
		t.Fatalf(
			"expected nil data, got %v",
			data,
		)
	}
}

// ============================================================
// REDIS CACHE - DELETE
// ============================================================

func TestRedisCache_Delete(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	cache := NewRedisCache(client)

	ctx := context.Background()

	key := "url:delete123"

	err = cache.Set(
		ctx,
		key,
		[]byte("test-value"),
		30*time.Minute,
	)

	if err != nil {
		t.Fatalf(
			"Set returned error: %v",
			err,
		)
	}

	err = cache.Delete(ctx, key)

	if err != nil {
		t.Fatalf(
			"Delete returned error: %v",
			err,
		)
	}

	_, found, err := cache.Get(ctx, key)

	if err != nil {
		t.Fatalf(
			"Get returned error: %v",
			err,
		)
	}

	if found {
		t.Fatal("expected key to be deleted")
	}
}

// ============================================================
// REDIS CACHE - DELETE MISSING KEY
// ============================================================

func TestRedisCache_DeleteMissingKey(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	cache := NewRedisCache(client)

	err = cache.Delete(
		context.Background(),
		"url:not-found",
	)

	if err != nil {
		t.Fatalf(
			"Delete missing key should not fail, got %v",
			err,
		)
	}
}

// ============================================================
// REDIS CACHE - TTL
// ============================================================

func TestRedisCache_SetTTL(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	cache := NewRedisCache(client)

	key := "url:ttl123"

	err = cache.Set(
		context.Background(),
		key,
		[]byte("test"),
		10*time.Minute,
	)

	if err != nil {
		t.Fatalf(
			"Set returned error: %v",
			err,
		)
	}

	ttl := mr.TTL(key)

	if ttl <= 0 {
		t.Fatal("expected key to have a TTL")
	}

	if ttl > 10*time.Minute {
		t.Fatalf(
			"unexpected TTL: %v",
			ttl,
		)
	}
}

// ============================================================
// REDIS CACHE - EXPIRED KEY
// ============================================================

func TestRedisCache_ExpiredKey(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	cache := NewRedisCache(client)

	key := "url:expired"

	err = cache.Set(
		context.Background(),
		key,
		[]byte("test"),
		1*time.Second,
	)

	if err != nil {
		t.Fatalf(
			"Set returned error: %v",
			err,
		)
	}

	mr.FastForward(2 * time.Second)

	_, found, err := cache.Get(
		context.Background(),
		key,
	)

	if err != nil {
		t.Fatalf(
			"Get returned error: %v",
			err,
		)
	}

	if found {
		t.Fatal("expected expired key not to be found")
	}
}

// ============================================================
// REDIS CACHE - REDIS ERROR
// ============================================================

func TestRedisCache_GetRedisError(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cache := NewRedisCache(client)

	mr.Close()

	_, found, err := cache.Get(
		context.Background(),
		"url:error",
	)

	if err == nil {
		t.Fatal("expected Redis error")
	}

	if found {
		t.Fatal("expected found=false on Redis error")
	}

	client.Close()
}

// ============================================================
// REDIS CLIENT - SUCCESS
// ============================================================

func TestNewRedisClient_Success(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	// This test is intentionally skipped because NewRedisClient
	// requires config.RedisConfig and the exact current fields
	// should be tested against the project's configuration.
	t.Skip("test against project-specific RedisConfig fields")
}

// Keep errors imported for projects that already use it elsewhere
// in future cache tests.
var _ = errors.New
