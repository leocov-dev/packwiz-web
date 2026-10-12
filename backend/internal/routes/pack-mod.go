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
	"packwiz-web/internal/services/compat_svc"
)

func RegisterPackModRoutes(router gin.IRouter, db *gorm.DB, riverClient *river.Client[*sql.Tx], compat *compat_svc.CompatService, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	packModController := controllers.NewPackwizModController(db, riverClient, compat)

	modGroup := router.Group("mod", handlers...)

	authz := authz_svc.NewService(db)
	can := func(permission string) gin.HandlerFunc {
		return middleware.RequirePackPermission(authz, permission)
	}

	modGroup.POST("", can(authz_svc.PackModAdd), packModController.AddMod)
	modGroup.POST("remove", can(authz_svc.PackModRemove), packModController.RemoveMods)
	modGroup.POST("missing-dependencies", can(authz_svc.PackModAdd), packModController.ListMissingDependencies)
	modGroup.GET("search", can(authz_svc.PackModAdd), packModController.SearchModrinthMods)
	modGroup.GET("search/modrinth", can(authz_svc.PackModAdd), packModController.SearchModrinthMods)
	modGroup.GET("search/curseforge", can(authz_svc.PackModAdd), packModController.SearchCurseforgeMods)
	modGroup.GET("search/curseforge/status", can(authz_svc.PackModAdd), packModController.CurseforgeStatus)

	bulkGroup := modGroup.Group("bulk")
	{
		bulkGroup.POST("", can(authz_svc.PackModAdd), packModController.StartBulkAdd)
		bulkGroup.GET("latest", can(authz_svc.PackModAdd), packModController.GetLatestBulkAdd)
		bulkGroup.GET(fmt.Sprintf(":%s", params.BulkRunId), can(authz_svc.PackModAdd), packModController.GetBulkAdd)
		bulkGroup.POST(fmt.Sprintf(":%s/item/:%s", params.BulkRunId, params.BulkItemId), can(authz_svc.PackModAdd), packModController.ResolveBulkItem)
	}

	modIdGroup := modGroup.Group(fmt.Sprintf(":%s", params.ModId))
	{
		modIdGroup.GET("", can(authz_svc.PackView), packModController.GetOneMod)
		modIdGroup.DELETE("", can(authz_svc.PackModRemove), packModController.RemoveMod)
		modIdGroup.PATCH("update", can(authz_svc.PackModUpdate), packModController.UpdateMod)
		modIdGroup.PATCH("side", can(authz_svc.PackModConfigure), packModController.ChangeModSide)
		modIdGroup.PATCH("option", can(authz_svc.PackModConfigure), packModController.ChangeModOption)
		modIdGroup.PATCH("pin", can(authz_svc.PackModConfigure), packModController.PinMod)
		modIdGroup.PATCH("unpin", can(authz_svc.PackModConfigure), packModController.UnPinMod)
		modIdGroup.PATCH("uplift", can(authz_svc.PackModConfigure), packModController.UpliftMod)
	}

	return modGroup
}
