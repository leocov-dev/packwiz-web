package routes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/controllers"
	"packwiz-web/internal/params"
)

func RegisterPackModRoutes(router gin.IRouter, db *gorm.DB, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	packModController := controllers.NewPackwizModController(db)

	modGroup := router.Group("mod", handlers...)

	modGroup.POST("", packModController.AddMod)
	modGroup.POST("missing-dependencies", packModController.ListMissingDependencies)
	modGroup.GET("search", packModController.SearchModrinthMods)

	modIdGroup := modGroup.Group(fmt.Sprintf(":%s", params.ModId))
	{
		modIdGroup.GET("", packModController.GetOneMod)
		modIdGroup.DELETE("", packModController.RemoveMod)
		modIdGroup.PATCH("update", packModController.UpdateMod)
		modIdGroup.PATCH("side", packModController.ChangeModSide)
		modIdGroup.PATCH("pin", packModController.PinMod)
		modIdGroup.PATCH("unpin", packModController.UnPinMod)
	}

	return modGroup
}
