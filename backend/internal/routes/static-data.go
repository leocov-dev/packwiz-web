package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/controllers"
)

func RegisterStaticDataRoutes(router gin.IRouter, db *gorm.DB, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	staticDataController := controllers.NewStaticDataController()

	staticDataGroup := router.Group("static-data", handlers...)
	{
		staticDataGroup.GET("", staticDataController.GetStaticData)
	}

	return staticDataGroup
}
