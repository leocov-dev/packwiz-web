package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"packwiz-web/internal/log"
)

func RateLimiter() gin.HandlerFunc {
	if gin.Mode() == gin.DebugMode {
		log.Warn("Rate limiter disabled in debug mode")
		return func(c *gin.Context) {
			c.Next()
		}
	}

	rate, err := limiter.NewRateFromFormatted("6-M")
	if err != nil {
		log.Panic(err)
	}

	store := memory.NewStore()

	instance := limiter.New(store, rate)

	return mgin.NewMiddleware(instance)
}
