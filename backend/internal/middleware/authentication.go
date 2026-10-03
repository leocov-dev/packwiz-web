package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/log"
	"packwiz-web/internal/params"
	"packwiz-web/internal/services/authz_svc"
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
			return
		}

		if !user.IsActive {
			ClearSession(c)
			log.Warn("user account is deactivated")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "account is deactivated"})
			return
		}

		system, err := authz_svc.NewService(db).SystemPermissions(user)
		if err != nil {
			log.Warn("failed to load permissions", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "failed to load permissions"})
			return
		}

		c.Set("user", user)
		c.Set(systemPermissionsKey, system)

		c.Next()
	}
}

// Context keys ConsumerAuthentication sets for PackwizAudit.
const (
	ConsumerPackKey = "consumerPack"
	ConsumerUserKey = "consumerUser"
)

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

		c.Set(ConsumerPackKey, pack)

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

		if !user.IsActive {
			log.Warn("deactivated user attempted pack link access")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set(ConsumerUserKey, user)

		authz := authz_svc.NewService(db)
		system, err := authz.SystemPermissions(user)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if err := authz.CanOnPack(user, system, pack.ID, authz_svc.PackConsume); err != nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}
