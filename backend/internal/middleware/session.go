package middleware

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"packwiz-web/internal/config"
	"packwiz-web/internal/logger"
)

func SessionStore() gin.HandlerFunc {
	store := cookie.NewStore(config.C.SessionSecret)

	isDevelopment := config.C.Mode == "development"

	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   !isDevelopment,
	})
	return sessions.Sessions("default", store)
}

func ClearSession(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		logger.Warn(fmt.Sprintf("Failed to clear and save session: %s", err))
	}
}
