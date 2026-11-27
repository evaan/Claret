package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func RequestTimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func internalError(c *gin.Context, err error) {
	log.Printf("API %s: %v", c.FullPath(), err)
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func RateLimiterMiddleware(limit string) (gin.HandlerFunc, error) {
	rate, err := limiter.NewRateFromFormatted(limit)
	if err != nil {
		return nil, fmt.Errorf("invalid API_RATE_LIMIT: %w", err)
	}

	store := memory.NewStore()
	instance := limiter.New(store, rate)

	return func(c *gin.Context) {
		if c.FullPath() == "/health" {
			c.Next()
			return
		}
		key := c.ClientIP()

		context, err := instance.Get(c, key)
		if err != nil {
			internalError(c, err)
			return
		}

		c.Header("RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Header("RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Header("RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		if context.Reached {
			c.Header("Retry-After", strconv.FormatInt(max(1, context.Reset-time.Now().Unix()), 10))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}, nil
}
