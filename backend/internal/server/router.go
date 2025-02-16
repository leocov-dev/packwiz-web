package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"packwiz-web/internal/controllers"
	"packwiz-web/internal/database"
	"packwiz-web/internal/middleware"
)

func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	if gin.Mode() == gin.DebugMode {
		router.Use(cors.New(cors.Config{
			AllowAllOrigins: true,
		}))
	}

	db := database.GetClient()

	// ---------------------------------------------------------------------
	packwizFiles := router.Group("packwiz")
	{
		modpackController := &controllers.ModpackController{}
		packwizFiles.GET(":modpack/*filepath", modpackController.ServeStatic)
	}

	// ---------------------------------------------------------------------
	v1 := router.Group("api/v1")
	v1.Use(middleware.Authentication())
	{
		// ---------------------------------------------------------------------
		healthGroup := v1.Group("health")
		{
			healthController := controllers.NewHealthController()
			healthGroup.GET("", healthController.Status)
		}

		// ---------------------------------------------------------------------
		userGroup := v1.Group("user")
		{
			// TODO
			userGroup.GET("", func(c *gin.Context) { c.JSON(200, gin.H{}) })
		}

		// ---------------------------------------------------------------------
		adminGroup := v1.Group("admin")
		{
			// TODO
			adminGroup.GET("", func(c *gin.Context) { c.JSON(200, gin.H{}) })
		}

		// ---------------------------------------------------------------------
		packwizGroup := v1.Group("packwiz")
		{
			packwizController := controllers.NewPackwizController(db)

			// ---------------------------------------------------------------------
			packwizGroup.GET("loaders", packwizController.ListLoaders)

			// ---------------------------------------------------------------------
			packwizGroup.GET("upload", packwizController.UploadPackwizArchive)

			// ---------------------------------------------------------------------
			packGroup := packwizGroup.Group("pack")
			{
				packGroup.GET("", packwizController.GetAllPacks)
				packGroup.POST("", packwizController.NewPack)

				// ---------------------------------------------------------------------
				slugGroup := packGroup.Group(":slug")
				{
					slugGroup.HEAD("", packwizController.PackHead)
					slugGroup.GET("", packwizController.GetOnePack)
					slugGroup.POST("", packwizController.AddMod)
					slugGroup.DELETE("", packwizController.RemovePack)
					slugGroup.PATCH("acceptable-versions", packwizController.SetAcceptableVersions)
					slugGroup.PATCH("update", packwizController.UpdateAll)
					slugGroup.PATCH("rename", packwizController.RenamePack)

					// ---------------------------------------------------------------------
					modGroup := slugGroup.Group("mod/:mod")
					{
						modGroup.DELETE("", packwizController.RemoveMod)
						modGroup.PATCH("rename", packwizController.RenameMod)
						modGroup.PATCH("update", packwizController.UpdateMod)
						modGroup.PATCH("side", packwizController.ChangeModSide)
						modGroup.PATCH("pin", packwizController.PinMod)
						modGroup.PATCH("unpin", packwizController.UnPinMod)
					}
				}
			}
		}
	}

	// ---------------------------------------------------------------------
	router.NoRoute(embeddedPublicHandler)

	return router
}
