package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/types/tables"
)

func ApiAuthentication(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userId := session.Get("userId")
		if userId == nil {
			ClearSession(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "unauthorized"})
			return
		}

		var user tables.User
		err := db.Where("id = ?", userId).First(&user).Error
		if err != nil {
			ClearSession(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "unauthorized"})
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

		var user tables.User
		if err := db.Where("username = ?", "admin").First(&user).Error; err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var packUser tables.PackUsers
		if err := db.Where("pack_slug = ? AND user_id = ?", slug, user.ID).First(&packUser).Error; err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()

	}
}
