package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/log"
	"packwiz-web/internal/params"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types"
)

func PackPermissionGuard(minPermission types.PackPermission, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		packId, err := mustBindIdParam(c, params.PackId)
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		user := c.MustGet("user").(tables.User)

		if err := db.Where(
			"pack_id = ? AND user_id = ? AND permission >= ?",
			packId, user.ID, minPermission,
		).First(&tables.PackUsers{}).Error; err != nil {
			log.Warn("no permission to access pack", packId, user.ID, minPermission)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}
