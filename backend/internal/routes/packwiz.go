package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/controllers"
	"packwiz-web/internal/middleware"
	"packwiz-web/internal/services/authz_svc"
	"packwiz-web/internal/services/packwiz_svc"
)

func RegisterPackwizRoutes(router gin.IRouter, db *gorm.DB, scans packwiz_svc.ScanRequester, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	packwizGroup := router.Group("packwiz", handlers...)
	{
		// ---------------------------------------------------------
		loadersController := controllers.NewLoadersController()

		packwizGroup.GET("loaders", loadersController.GetLoaderVersions, middleware.SkipAudit)

		// ---------------------------------------------------------
		importController := controllers.NewImportController(db, scans)

		packwizGroup.POST("import", middleware.RequirePermission(authz_svc.PackCreate), importController.ImportPack)

		// ---------------------------------------------------------

	}

	return packwizGroup
}
