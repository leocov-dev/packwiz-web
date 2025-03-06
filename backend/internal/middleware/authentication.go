package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/logger"
	"packwiz-web/internal/services/user_svc"
	tables "packwiz-web/internal/tables"
)

func ApiAuthentication(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userId := session.Get("userId")
		if userId == nil {
			ClearSession(c)
			logger.Warn("no user session")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "no user session"})
			return
		}

		userService := user_svc.NewUserService(db)

		user, err := userService.FindById(userId.(uint))
		if err != nil {
			ClearSession(c)
			logger.Warn("no user match")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "no user match"})
			return
		}

		c.Set("user", user)

		c.Next()
	}
}

func PackwizFileAuthentication(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

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

		token := c.Query("token")
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
		if err := db.Where("pack_slug = ? AND user_id = ?", slug, user.Id).First(&packUser).Error; err != nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()

	}
}
