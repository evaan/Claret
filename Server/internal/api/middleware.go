package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func RateLimiterMiddleware(limit string) gin.HandlerFunc {
	rate, err := limiter.NewRateFromFormatted(limit)
	if err != nil {
		panic(err)
	}

	store := memory.NewStore()
	instance := limiter.New(store, rate)

	return func(c *gin.Context) {
		key := c.ClientIP()

		context, err := instance.Get(c, key)
		if err != nil {
			fmt.Printf("Error on rate limiter: %s\n", err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Header("RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Header("RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Header("RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		if context.Reached {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
		}

		c.Next()
	}
}
