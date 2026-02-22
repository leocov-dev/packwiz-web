package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/controllers"
	"packwiz-web/internal/middleware"
)

func RegisterPackwizRoutes(router gin.IRouter, db *gorm.DB, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	packwizGroup := router.Group("packwiz", handlers...)
	{
		// ---------------------------------------------------------
		loadersController := controllers.NewLoadersController()

		packwizGroup.GET("loaders", loadersController.GetLoaderVersions, middleware.SkipAudit)

		// ---------------------------------------------------------
		importController := controllers.NewImportController(db)

		packwizGroup.GET("upload", importController.UploadPackwizArchive)

		// ---------------------------------------------------------

	}

	return packwizGroup
}
