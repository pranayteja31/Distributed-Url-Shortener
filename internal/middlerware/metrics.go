package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"pranayteja31/Urlshortener/internal/metrics"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now()

		c.Next()

		metrics.ObserveRequest(
			start,
			c.Request.Method,
			c.FullPath(),
			c.Writer.Status(),
		)
	}
}
