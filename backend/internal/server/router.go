package server

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"packwiz-web/internal/controllers"
	"packwiz-web/internal/database"
	"packwiz-web/internal/middleware"
	"packwiz-web/internal/middleware/meta"
	"packwiz-web/internal/params"
	"packwiz-web/internal/types"
	"packwiz-web/public"
	"time"
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
		AllowHeaders: []string{
			"Content-Type",
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		MaxAge: 12 * time.Hour,
	}))

	db := database.GetClient()

	// -------------------------------------------------------------------------
	packwizFiles := router.Group(fmt.Sprintf("packwiz/:%s/:%s", params.Token, params.PackSlug))
	packwizFiles.Use(middleware.ConsumerAuthentication(db))
	packwizFiles.Use(middleware.PackwizAudit(db))
	{
		tomlController := controllers.NewTomlController(db)
		packwizFiles.GET("pack.toml", tomlController.RenderPackToml)
		packwizFiles.GET("index.toml", tomlController.RenderIndexToml)
		packwizFiles.GET(fmt.Sprintf(":%s/:%s", params.ModType, params.ModSlug), tomlController.RenderModToml)
	}

	// -------------------------------------------------------------------------
	api := router.Group("api")
	api.Use(middleware.SessionStore())
	api.Use(middleware.ApiAudit(db))
	{
		// ---------------------------------------------------------------------
		v1 := api.Group("v1")
		{
			authController := controllers.NewAuthController(db)

			v1.POST("login", middleware.RateLimiter(), meta.Tag(meta.CategoryLogin), authController.Login)
			v1.POST("logout", authController.Logout)

			protectedGroup := v1.Group("")
			protectedGroup.Use(middleware.ApiAuthentication(db))
			{

				// -------------------------------------------------------------
				healthGroup := protectedGroup.Group("health", middleware.SkipAudit)
				{
					healthController := controllers.NewHealthController()
					healthGroup.GET("", healthController.Status)
				}

				userController := controllers.NewUserController(db)

				// -------------------------------------------------------------
				// current user
				userGroup := protectedGroup.Group("user")
				{
					userGroup.GET("", userController.GetCurrentUser, middleware.SkipAudit)
					userGroup.POST("password",
						func(c *gin.Context) {
							if err := userController.ChangePassword; err != nil {

								return
							}
						})
					userGroup.POST("update", userController.UpdateUser)
					userGroup.POST("invalidate-sessions", userController.InvalidateCurrentUserSessions)
				}

				// -------------------------------------------------------------
				adminGroup := protectedGroup.Group("admin")
				{
					adminGroup.GET("users", userController.GetUsersPaginated)
				}

				// -------------------------------------------------------------
				packwizGroup := protectedGroup.Group("packwiz")
				{
					// ---------------------------------------------------------
					loadersController := controllers.NewLoadersController()

					packwizGroup.GET("loaders", loadersController.GetLoaderVersions, middleware.SkipAudit)
					//packwizGroup.GET("loaders/versions", loadersController.GetLoaderVersions)

					// ---------------------------------------------------------
					importController := controllers.NewImportController(db)

					packwizGroup.GET("upload", importController.UploadPackwizArchive)

					// ---------------------------------------------------------
					packwizController := controllers.NewPackwizController(db)

					packGroup := packwizGroup.Group("pack")
					{
						packGroup.GET("", packwizController.GetAllPacks)
						packGroup.POST("", packwizController.NewPack)

						// -----------------------------------------------------
						canViewPackGuard := middleware.PackPermissionGuard(types.PackPermissionView, db)
						canEditPackGuard := middleware.PackPermissionGuard(types.PackPermissionEdit, db)

						slugGroup := packGroup.Group(fmt.Sprintf(":%s", params.PackSlug))
						slugGroup.Use(canViewPackGuard)
						{
							slugGroup.HEAD("", packwizController.PackHead)
							slugGroup.GET("", packwizController.GetOnePack)
							slugGroup.GET("link", packwizController.GetPersonalizedLink)

							editPackGroup := slugGroup.Group("")
							editPackGroup.Use(canEditPackGuard)
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

								// ---------------------------------------------
								editPackGroup.POST("mod", packwizController.AddMod)
								modGroup := editPackGroup.Group(fmt.Sprintf("mod/:%s", params.ModSlug))
								{
									modGroup.GET("", packwizController.GetOneMod)
									modGroup.DELETE("", packwizController.RemoveMod)
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
