package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/controllers"
	"packwiz-web/internal/middleware"
	"packwiz-web/internal/services/authz_svc"
)

func RegisterAdminRoutes(router gin.IRouter, db *gorm.DB, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	adminController := controllers.NewAdminController(db)
	accessController := controllers.NewPackAccessController(db)

	adminGroup := router.Group("admin", handlers...)
	{
		adminGroup.GET("users", middleware.RequirePermission(authz_svc.UserView), adminController.GetUsersPaginated)
		adminGroup.POST("users", middleware.RequirePermission(authz_svc.UserCreate), adminController.CreateUser)
		adminGroup.GET("users/:userId", middleware.RequirePermission(authz_svc.UserView), adminController.GetUserById)
		adminGroup.PATCH("users/:userId", middleware.RequirePermission(authz_svc.UserManage), adminController.UpdateUser)
		adminGroup.PATCH("users/:userId/deactivate", middleware.RequirePermission(authz_svc.UserManage), adminController.DeactivateUser)
		adminGroup.PATCH("users/:userId/reactivate", middleware.RequirePermission(authz_svc.UserManage), adminController.ReactivateUser)
		adminGroup.POST("users/:userId/reset-password", middleware.RequirePermission(authz_svc.UserManage), adminController.ResetUserPassword)
		adminGroup.GET("roles", middleware.RequirePermission(authz_svc.UserView), adminController.ListRoles)
		adminGroup.GET("permissions", middleware.RequirePermission(authz_svc.UserView), adminController.ListPermissions)
		adminGroup.PUT("users/:userId/roles", middleware.RequirePermission(authz_svc.UserRolesAssign), adminController.SetUserRoles)
		adminGroup.GET("audits", middleware.RequirePermission(authz_svc.AuditView), adminController.GetAuditsPaginated)
		adminGroup.GET("pack-access", middleware.RequirePermission(authz_svc.AuditView), accessController.SystemSummary)
		adminGroup.GET("pack-access/recent", middleware.RequirePermission(authz_svc.AuditView), accessController.SystemRecent)
		adminGroup.GET("instance-downloads", middleware.RequirePermission(authz_svc.AuditView), accessController.SystemDownloadsSummary)
		adminGroup.GET("instance-downloads/recent", middleware.RequirePermission(authz_svc.AuditView), accessController.SystemDownloadsRecent)
	}

	return adminGroup
}
