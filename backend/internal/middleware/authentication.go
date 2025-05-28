package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/log"
	"packwiz-web/internal/params"
	"packwiz-web/internal/services/user_svc"
	"packwiz-web/internal/tables"
)

func ApiAuthentication(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userId := session.Get("userId")
		if userId == nil {
			ClearSession(c)
			log.Warn("no user session")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "no user session"})
			return
		}

		userService := user_svc.NewUserService(db)

		user, err := userService.FindById(userId.(uint))
		if err != nil {
			ClearSession(c)
			log.Warn("no user match")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "no user match"})
			return
		}

		sessionKey := session.Get("sessionKey")
		if sessionKey == nil || sessionKey != user.SessionKey {
			ClearSession(c)
			log.Warn("session key mismatch")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "session is invalid, log in again"})
		}

		c.Set("user", user)

		c.Next()
	}
}

func ConsumerAuthentication(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param(string(params.PackSlug))

		if slug == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		var pack tables.Pack
		if err := db.Where("slug = ?", slug).First(&pack).Error; err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		if pack.IsPublic {
			c.Next()
			return
		}

		token := c.Param(string(params.Token))
		if token == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var user tables.User
		if err := db.Where("link_token = ?", token).First(&user).Error; err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var packUser tables.PackUsers
		if err := db.Where("pack_slug = ? AND user_id = ?", slug, user.ID).First(&packUser).Error; err != nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}
