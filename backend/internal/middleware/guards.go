package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types"
)

func PermissionGuard(minPermission types.PackPermission, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		user := c.MustGet("user").(tables.User)

		if err := db.Where(
			"pack_slug = ? AND user_id = ? AND permission >= ?",
			slug, user.Id, minPermission,
		).First(&tables.PackUsers{}).Error; err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}
