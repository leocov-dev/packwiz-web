package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"packwiz-web/internal/log"
)

// RateLimiter is the strict limiter used on login.
func RateLimiter() gin.HandlerFunc {
	return RateLimit("6-M")
}

// RateLimit limits requests per client IP, e.g. "10-M" for 10 per minute. Each
// call gets its own counter, so routes do not share a budget. Disabled in debug mode.
func RateLimit(formatted string) gin.HandlerFunc {
	if gin.Mode() == gin.DebugMode {
		log.Warn("Rate limiter disabled in debug mode")
		return func(c *gin.Context) {
			c.Next()
		}
	}

	rate, err := limiter.NewRateFromFormatted(formatted)
	if err != nil {
		log.Panic(err)
	}

	store := memory.NewStore()

	instance := limiter.New(store, rate)

	return mgin.NewMiddleware(instance)
}

// RateLimitIf applies a RateLimit only to requests for which match is true.
// It lets one limit cover a single route of a group whose shared middleware
// (like authentication) must run after the limit.
func RateLimitIf(formatted string, match func(*gin.Context) bool) gin.HandlerFunc {
	limit := RateLimit(formatted)
	return func(c *gin.Context) {
		if match(c) {
			limit(c)
			return
		}
		c.Next()
	}
}
