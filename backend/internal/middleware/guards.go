package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"packwiz-web/internal/log"
	"packwiz-web/internal/params"
	"packwiz-web/internal/services/authz_svc"
	"packwiz-web/internal/tables"
)

const systemPermissionsKey = "systemPermissions"

// SystemPermissions returns the system permission set loaded by ApiAuthentication.
func SystemPermissions(c *gin.Context) authz_svc.Set {
	set, _ := c.Get(systemPermissionsKey)
	s, _ := set.(authz_svc.Set)
	return s
}

// RequirePermission guards a route behind a global permission.
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.MustGet("user").(tables.User)

		if !authz_svc.DecideGlobal(user.IsSuperuser, SystemPermissions(c), permission) {
			log.Warn("missing permission", user.ID, permission)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}

// RequirePackPermission guards a route behind a permission on the pack named
// by the :packId param. A system role holding the permission applies to every
// pack; a pack role applies to that pack only.
func RequirePackPermission(authz *authz_svc.Service, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		packId, err := mustBindIdParam(c, params.PackId)
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		user := c.MustGet("user").(tables.User)

		switch err := authz.CanOnPack(user, SystemPermissions(c), packId, permission); {
		case err == nil:
			c.Next()
		case errors.Is(err, authz_svc.ErrPackNotFound):
			c.AbortWithStatus(http.StatusNotFound)
		case errors.Is(err, authz_svc.ErrArchived):
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"msg": "pack is archived"})
		case errors.Is(err, authz_svc.ErrForbidden):
			log.Warn("no permission to access pack", packId, user.ID, permission)
			c.AbortWithStatus(http.StatusForbidden)
		default:
			log.Warn("permission check failed", err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}
}
