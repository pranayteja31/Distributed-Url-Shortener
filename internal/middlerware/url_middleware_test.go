package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	return gin.New()
}
func TestRateLimiter_AllowsRequest(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer redisClient.Close()

	router := setupTestRouter()

	router.Use(RateLimiter(redisClient))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}
func TestRateLimiter_LimitExceeded(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer redisClient.Close()

	router := setupTestRouter()

	router.Use(RateLimiter(redisClient))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	})

	for i := 1; i <= requestLimit+1; i++ {
		req := httptest.NewRequest(
			http.MethodGet,
			"/test",
			nil,
		)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if i <= requestLimit {
			if rec.Code != http.StatusOK {
				t.Fatalf(
					"request %d: expected status %d, got %d",
					i,
					http.StatusOK,
					rec.Code,
				)
			}
		}

		if i == requestLimit+1 {
			if rec.Code != http.StatusTooManyRequests {
				t.Fatalf(
					"request %d: expected status %d, got %d",
					i,
					http.StatusTooManyRequests,
					rec.Code,
				)
			}
		}
	}
}
func TestRateLimiter_AbortsRequest(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer redisClient.Close()

	router := setupTestRouter()

	router.Use(RateLimiter(redisClient))

	handlerCalled := false

	router.GET("/test", func(c *gin.Context) {
		handlerCalled = true

		c.JSON(http.StatusOK, gin.H{
			"message": "handler called",
		})
	})

	// Consume the complete limit.
	for i := 0; i < requestLimit; i++ {
		req := httptest.NewRequest(
			http.MethodGet,
			"/test",
			nil,
		)

		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
	}

	handlerCalled = false

	// 31st request.
	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusTooManyRequests,
			rec.Code,
		)
	}

	if handlerCalled {
		t.Fatal("handler should not be called after rate limit is exceeded")
	}
}
func TestPrometheusMiddleware(t *testing.T) {
	router := setupTestRouter()

	router.Use(PrometheusMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}
func TestPrometheusMiddleware_Methods(t *testing.T) {
	router := setupTestRouter()

	router.Use(PrometheusMiddleware())

	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{
			"message": "created",
		})
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/test",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}
}
