package server

import (
	"embed"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"packwiz-web/internal/controllers"
	"packwiz-web/internal/database"
	"packwiz-web/internal/middleware"
)

func NewRouter(publicFiles *embed.FS) *gin.Engine {
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
	packwizFiles := router.Group("packwiz/:slug")
	packwizFiles.Use(middleware.PackwizFileAuthentication(db))
	{
		modpackController := &controllers.ModpackController{}
		packwizFiles.GET("*filepath", modpackController.ServeStatic)
	}

	// ---------------------------------------------------------------------
	api := router.Group("api")
	api.Use(middleware.SessionStore())
	{
		// ---------------------------------------------------------------------
		v1 := api.Group("v1")
		{
			userController := controllers.NewUserController(db)

			v1.POST("login", userController.Login)
			v1.POST("logout", userController.Logout)

			protectedGroup := v1.Group("")
			protectedGroup.Use(middleware.ApiAuthentication(db))
			{

				// ---------------------------------------------------------------------
				healthGroup := protectedGroup.Group("health")
				{
					healthController := controllers.NewHealthController()
					healthGroup.GET("", healthController.Status)
				}

				// ---------------------------------------------------------------------
				userGroup := protectedGroup.Group("user")
				{
					// TODO
					userGroup.GET("", func(c *gin.Context) { c.JSON(200, gin.H{}) })
				}

				// ---------------------------------------------------------------------
				adminGroup := protectedGroup.Group("admin")
				{
					// TODO
					adminGroup.GET("", func(c *gin.Context) { c.JSON(200, gin.H{}) })
				}

				// ---------------------------------------------------------------------
				packwizGroup := protectedGroup.Group("packwiz")
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
		}
	}

	// ---------------------------------------------------------------------
	embeddedSPAController := NewEmbeddedSPAController(publicFiles)
	router.NoRoute(embeddedSPAController.Handler)

	return router
}
