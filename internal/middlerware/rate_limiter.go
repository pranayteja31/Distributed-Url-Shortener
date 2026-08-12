package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	requestLimit = 30
	window       = time.Minute
)

func RateLimiter(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		clientIP := c.ClientIP()

		key := fmt.Sprintf("rate_limiter:%s", clientIP)

		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}
		//set 1 min window expiry
		if count == 1 {
			if err := redisClient.Expire(ctx, key, window).Err(); err != nil {
				c.Next()
				return
			}
		}
		//request limit exceeded
		if count > requestLimit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate Limit Exceeded",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
