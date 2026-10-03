package routes

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/riverqueue/river"
	"gorm.io/gorm"
	"packwiz-web/internal/controllers"
	"packwiz-web/internal/middleware"
	"packwiz-web/internal/params"
	"packwiz-web/internal/services/authz_svc"
)

func RegisterPackRoutes(router gin.IRouter, db *gorm.DB, riverClient *river.Client[*sql.Tx], handlers ...gin.HandlerFunc) *gin.RouterGroup {
	packwizController := controllers.NewPackwizController(db, riverClient)
	authz := authz_svc.NewService(db)

	packGroup := router.Group("pack", handlers...)
	{
		packGroup.GET("", packwizController.GetAllPacks)
		packGroup.POST("", middleware.RequirePermission(authz_svc.PackCreate), packwizController.NewPack)
		packGroup.GET("roles", packwizController.GetPackRoles)

		// -----------------------------------------------------
		can := func(permission string) gin.HandlerFunc {
			return middleware.RequirePackPermission(authz, permission)
		}

		packIdGroup := packGroup.Group(fmt.Sprintf(":%s", params.PackId))
		{
			packIdGroup.HEAD("", can(authz_svc.PackView), packwizController.PackHead)
			packIdGroup.GET("", can(authz_svc.PackView), packwizController.GetOnePack)
			packIdGroup.GET("updates", can(authz_svc.PackView), packwizController.GetUpdateChecks)
			packIdGroup.GET("link", can(authz_svc.PackLink), packwizController.GetPersonalizedLink)

			packIdGroup.DELETE("", can(authz_svc.PackArchive), packwizController.ArchivePack)
			packIdGroup.PATCH("unarchive", can(authz_svc.PackArchive), packwizController.UnArchivePack)
			packIdGroup.PATCH("publish", can(authz_svc.PackPublish), packwizController.PublishPack)
			packIdGroup.PATCH("draft", can(authz_svc.PackPublish), packwizController.ConvertToDraft)
			packIdGroup.PATCH("public", can(authz_svc.PackVisibility), packwizController.MakePublic)
			packIdGroup.PATCH("private", can(authz_svc.PackVisibility), packwizController.MakePrivate)
			packIdGroup.PATCH("edit", can(authz_svc.PackInfoEdit), packwizController.EditPackInfo)
			packIdGroup.PATCH("update-all", can(authz_svc.PackModUpdate), packwizController.UpdateAll)
			packIdGroup.POST("updates/check", can(authz_svc.PackUpdatesCheck), packwizController.CheckForUpdates)
			packIdGroup.PATCH("rehash", can(authz_svc.PackRehash), packwizController.RehashAll)
			packIdGroup.PATCH("migrate", can(authz_svc.PackMigrate), packwizController.MigratePack)
			packIdGroup.POST("migrate/dry-run", can(authz_svc.PackMigrate), packwizController.MigrateDryRun)
			packIdGroup.GET(fmt.Sprintf("migrate/job/:%s", params.JobId), can(authz_svc.PackMigrate), packwizController.MigrateJobStatus)
			packIdGroup.GET("users", can(authz_svc.PackUsersView), packwizController.GetPackUsers)
			packIdGroup.GET("users/search", can(authz_svc.PackUsersManage), middleware.RequirePermission(authz_svc.UserLookup), packwizController.SearchPackUsers)
			packIdGroup.POST("users", can(authz_svc.PackUsersManage), packwizController.AddPackUser)
			packIdGroup.DELETE(fmt.Sprintf("users/:%s", params.UserID), can(authz_svc.PackUsersManage), packwizController.RemovePackUser)
			packIdGroup.PATCH(fmt.Sprintf("users/:%s", params.UserID), can(authz_svc.PackUsersManage), packwizController.EditUserAccess)

			RegisterPackModRoutes(packIdGroup, db)
		}

	}

	return packGroup
}
