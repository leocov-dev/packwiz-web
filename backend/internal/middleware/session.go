package middleware

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"net/http"
	"packwiz-web/internal/config"
	"packwiz-web/internal/log"
)

func SessionStore() gin.HandlerFunc {
	store := cookie.NewStore(config.C.SessionSecret)

	isDevelopment := config.C.Mode == "development"

	var sameSite http.SameSite
	if isDevelopment {
		sameSite = http.SameSiteLaxMode
	} else {
		sameSite = http.SameSiteStrictMode
	}

	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   !isDevelopment,
		SameSite: sameSite,
	})
	return sessions.Sessions("session", store)
}

func ClearSession(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		log.Warn(fmt.Sprintf("Failed to clear and save session: %s", err))
	}
}
