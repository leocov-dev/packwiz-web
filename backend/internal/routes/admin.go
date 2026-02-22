package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/controllers"
)

func RegisterAdminRoutes(router gin.IRouter, db *gorm.DB, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	adminController := controllers.NewAdminController(db)

	adminGroup := router.Group("admin")
	{
		adminGroup.GET("users", adminController.GetUsersPaginated)
	}

	return adminGroup
}
