package routes

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/riverqueue/river"
	"gorm.io/gorm"

	"packwiz-web/internal/controllers"
	"packwiz-web/internal/middleware"
)

// RegisterPublicRoutes registers unauthenticated routes. Every handler here must
// itself refuse anything that is not explicitly public.
func RegisterPublicRoutes(router gin.IRouter, db *gorm.DB, riverClient *river.Client[*sql.Tx]) {
	publicController := controllers.NewPublicPackController(db, riverClient)

	// no user on these requests, so there is nothing for audit to attribute
	router.GET("public/packs/:packSlug", publicController.GetPublicPack, middleware.SkipAudit)
	router.GET("public/packs/:packSlug/changelist", publicController.GetPublicChangelist, middleware.SkipAudit)
}
