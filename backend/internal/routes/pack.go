package routes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/controllers"
	"packwiz-web/internal/middleware"
	"packwiz-web/internal/params"
	"packwiz-web/internal/types"
)

func RegisterPackRoutes(router gin.IRouter, db *gorm.DB, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	packwizController := controllers.NewPackwizController(db)

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
				editPackGroup.GET("users", packwizController.GetPackUsers)
				editPackGroup.POST("users", packwizController.AddPackUser)
				editPackGroup.DELETE(fmt.Sprintf("users/:%s", params.UserID), packwizController.RemovePackUser)
				editPackGroup.PATCH(fmt.Sprintf("users/:%s", params.UserID), packwizController.EditUserAccess)

			}
			RegisterPackModRoutes(editPackGroup, db)
		}

	}

	return packGroup
}
