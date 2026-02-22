package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/controllers"
	"packwiz-web/internal/middleware"
	"packwiz-web/internal/middleware/meta"
)

func RegisterAuthRoutes(router gin.IRouter, db *gorm.DB, handlers ...gin.HandlerFunc) *gin.RouterGroup {

	authController := controllers.NewAuthController(db)

	authGroup := router.Group("auth", handlers...)
	{
		authGroup.POST("login", middleware.RateLimiter(), meta.Tag(meta.CategoryLogin), authController.Login)
		authGroup.POST("logout", authController.Logout, middleware.SkipAudit)
	}

	return authGroup
}
