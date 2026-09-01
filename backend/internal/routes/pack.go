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
	"packwiz-web/internal/types"
)

func RegisterPackRoutes(router gin.IRouter, db *gorm.DB, riverClient *river.Client[*sql.Tx], handlers ...gin.HandlerFunc) *gin.RouterGroup {
	packwizController := controllers.NewPackwizController(db, riverClient)

	packGroup := router.Group("pack", handlers...)
	{
		packGroup.GET("", packwizController.GetAllPacks)
		packGroup.POST("", packwizController.NewPack)

		// -----------------------------------------------------
		canViewPackGuard := middleware.PackPermissionGuard(types.PackPermissionView, db)
		canEditPackGuard := middleware.PackPermissionGuard(types.PackPermissionEdit, db)

		packIdGroup := packGroup.Group(fmt.Sprintf(":%s", params.PackId), canViewPackGuard)
		{
			packIdGroup.HEAD("", packwizController.PackHead)
			packIdGroup.GET("", packwizController.GetOnePack)
			packIdGroup.GET("link", packwizController.GetPersonalizedLink)

			editPackGroup := packIdGroup.Group("", canEditPackGuard)
			{
				editPackGroup.DELETE("", packwizController.ArchivePack)
				editPackGroup.PATCH("unarchive", packwizController.UnArchivePack)
				editPackGroup.PATCH("publish", packwizController.PublishPack)
				editPackGroup.PATCH("draft", packwizController.ConvertToDraft)
				editPackGroup.PATCH("public", packwizController.MakePublic)
				editPackGroup.PATCH("private", packwizController.MakePrivate)
				editPackGroup.PATCH("edit", packwizController.EditPackInfo)
				editPackGroup.PATCH("update-all", packwizController.UpdateAll)
				editPackGroup.PATCH("rehash", packwizController.RehashAll)
				editPackGroup.PATCH("migrate", packwizController.MigratePack)
				editPackGroup.POST("migrate/dry-run", packwizController.MigrateDryRun)
				editPackGroup.GET(fmt.Sprintf("migrate/job/:%s", params.JobId), packwizController.MigrateJobStatus)
				editPackGroup.GET("users", packwizController.GetPackUsers)
				editPackGroup.GET("users/search", packwizController.SearchPackUsers)
				editPackGroup.POST("users", packwizController.AddPackUser)
				editPackGroup.DELETE(fmt.Sprintf("users/:%s", params.UserID), packwizController.RemovePackUser)
				editPackGroup.PATCH(fmt.Sprintf("users/:%s", params.UserID), packwizController.EditUserAccess)

			}
			RegisterPackModRoutes(editPackGroup, db)
		}

	}

	return packGroup
}
