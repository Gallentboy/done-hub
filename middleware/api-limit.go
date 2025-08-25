package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

const (
	LIMIT_KEY               = "api-limiter:%d"
	INTERNAL                = 1 * time.Minute
	RATE_LIMIT_EXCEEDED_MSG = "您的速率达到上限，请稍后再试。"
	SERVER_ERROR_MSG        = "Server error"
)

func DynamicRedisRateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
