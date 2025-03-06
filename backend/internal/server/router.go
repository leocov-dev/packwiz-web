package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"packwiz-web/internal/controllers"
	"packwiz-web/internal/database"
	"packwiz-web/internal/middleware"
	"packwiz-web/internal/types"
	"packwiz-web/public"
)

func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:8080",
		},
		AllowCredentials: true,
	}))

	db := database.GetClient()

	// ---------------------------------------------------------------------
	packwizFiles := router.Group("packwiz/:slug")
	packwizFiles.Use(middleware.PackwizFileAuthentication(db))
	packwizFiles.Use(middleware.PackwizAudit(db))
	{
		modpackController := controllers.NewModpackController(db)
		packwizFiles.GET("*filepath", modpackController.ServeStatic)
	}

	// ---------------------------------------------------------------------
	api := router.Group("api")
	api.Use(middleware.SessionStore())
	api.Use(middleware.ApiAudit(db))
	{
		// ---------------------------------------------------------------------
		v1 := api.Group("v1")
		{
			authController := controllers.NewAuthController(db)

			v1.POST("login", authController.Login)
			v1.POST("logout", authController.Logout)

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
					userController := controllers.NewUserController(db)
					// TODO
					userGroup.GET("", func(c *gin.Context) { c.JSON(200, gin.H{}) })
					userGroup.GET("current", userController.CurrentUser)
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
						canViewPackGuard := middleware.PackPermissionGuard(types.PackPermissionView, db)
						canEditPackGuard := middleware.PackPermissionGuard(types.PackPermissionEdit, db)

						slugGroup := packGroup.Group(":slug")
						slugGroup.Use(canViewPackGuard)
						{
							slugGroup.HEAD("", packwizController.PackHead)
							slugGroup.GET("", packwizController.GetOnePack)
							slugGroup.GET("link", packwizController.GetPersonalizedLink)

							editPackGroup := slugGroup.Group("")
							editPackGroup.Use(canEditPackGuard)
							{
								editPackGroup.POST("", packwizController.AddMod)
								editPackGroup.DELETE("", packwizController.ArchivePack)
								editPackGroup.PATCH("acceptable-versions", packwizController.SetAcceptableVersions)
								editPackGroup.PATCH("update", packwizController.UpdateAll)
								editPackGroup.PATCH("rename", packwizController.RenamePack)
								editPackGroup.GET("users", packwizController.GetPackUsers)
								editPackGroup.POST("users", packwizController.AddPackUser)
								editPackGroup.DELETE("users/:userId", packwizController.RemovePackUser)
								editPackGroup.PATCH("users/:userId", packwizController.EditUserAccess)

								// ---------------------------------------------------------------------
								modGroup := editPackGroup.Group("mod/:mod")
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
	}

	// ---------------------------------------------------------------------
	embeddedSPAController := controllers.NewFrontendController(public.GetFrontendFiles())
	router.NoRoute(embeddedSPAController.Handler)

	return router
}
