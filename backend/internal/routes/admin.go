package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/controllers"
)

func RegisterAdminRoutes(router gin.IRouter, db *gorm.DB, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	adminController := controllers.NewAdminController(db)

	adminGroup := router.Group("admin", handlers...)
	{
		adminGroup.GET("users", adminController.GetUsersPaginated)
		adminGroup.POST("users", adminController.CreateUser)
		adminGroup.GET("users/:userId", adminController.GetUserById)
		adminGroup.PATCH("users/:userId", adminController.UpdateUser)
		adminGroup.PATCH("users/:userId/deactivate", adminController.DeactivateUser)
		adminGroup.PATCH("users/:userId/reactivate", adminController.ReactivateUser)
		adminGroup.GET("audits", adminController.GetAuditsPaginated)
	}

	return adminGroup
}
