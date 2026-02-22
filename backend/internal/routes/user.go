package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/controllers"
)

func RegisterUserRoutes(router gin.IRouter, db *gorm.DB, handlers ...gin.HandlerFunc) *gin.RouterGroup {

	userController := controllers.NewUserController(db)

	userGroup := router.Group("user", handlers...)
	{
		userGroup.GET("", userController.GetCurrentUser)
		userGroup.POST("password", userController.ChangePassword)
		userGroup.POST("update", userController.UpdateUser)
		userGroup.POST("invalidate-sessions", userController.InvalidateCurrentUserSessions)
	}

	return userGroup
}
