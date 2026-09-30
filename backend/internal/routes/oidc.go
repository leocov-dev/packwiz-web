package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/controllers"
	"packwiz-web/internal/middleware"
	"packwiz-web/internal/middleware/meta"
	"packwiz-web/internal/services/oidc_svc"
)

// RegisterOidcAuthRoutes mounts the public sign-in endpoints. Rate limiting is
// delegated to the edge proxy, and the public routes cannot be audited (no
// user yet), so none is applied here; the callback opts out explicitly.
func RegisterOidcAuthRoutes(
	router gin.IRouter,
	db *gorm.DB,
	providers *oidc_svc.ProviderService,
	flow *oidc_svc.FlowService,
	handlers ...gin.HandlerFunc,
) *gin.RouterGroup {
	controller := controllers.NewOidcAuthController(db, providers, flow)

	authGroup := router.Group("auth", handlers...)
	{
		authGroup.GET("config", middleware.SkipAudit, controller.Config)
		authGroup.GET("oidc/:slug/login", controller.Login)
		authGroup.GET("oidc/:slug/callback", middleware.SkipAudit, controller.Callback)
	}

	return authGroup
}

// RegisterOidcIdentityRoutes mounts the self-service linked-account endpoints
// for the signed-in user. They are deliberately not under RegisterUserRoutes,
// whose group skips auditing.
func RegisterOidcIdentityRoutes(
	router gin.IRouter,
	db *gorm.DB,
	flow *oidc_svc.FlowService,
	handlers ...gin.HandlerFunc,
) *gin.RouterGroup {
	controller := controllers.NewOidcIdentityController(db, flow)

	group := router.Group("user/identities", handlers...)
	{
		group.GET("", middleware.SkipAudit, controller.List)
		group.POST(":slug/link", meta.Tag(meta.CategoryOidcLink), controller.Link)
		group.DELETE(":identityId", meta.Tag(meta.CategoryOidcUnlink), controller.Unlink)
	}

	return group
}

// RegisterOidcAdminRoutes mounts provider management and admin-side identity
// management. Mount it behind AdminGuard.
func RegisterOidcAdminRoutes(
	router gin.IRouter,
	db *gorm.DB,
	providers *oidc_svc.ProviderService,
	handlers ...gin.HandlerFunc,
) *gin.RouterGroup {
	controller := controllers.NewOidcAdminController(db, providers)

	group := router.Group("admin", handlers...)
	{
		group.GET("oidc/providers", controller.ListProviders)
		group.POST("oidc/providers", controller.CreateProvider)
		group.POST("oidc/providers/test", controller.TestDiscovery)
		group.GET("oidc/providers/:providerId", controller.GetProvider)
		group.PUT("oidc/providers/:providerId", controller.UpdateProvider)
		group.DELETE("oidc/providers/:providerId", controller.DeleteProvider)
		group.GET("oidc/providers/:providerId/orphaned-users", controller.OrphanedUsers)

		group.GET("users/:userId/identities", controller.ListUserIdentities)
		group.DELETE("users/:userId/identities/:identityId", controller.UnlinkUserIdentity)
	}

	return group
}
