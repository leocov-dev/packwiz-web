package server

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"packwiz-web/internal/controllers"
	"packwiz-web/internal/database"
	"packwiz-web/internal/middleware"
	"packwiz-web/internal/params"
	"packwiz-web/internal/routes"
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
	api := router.Group("api", middleware.SessionStore(), middleware.ApiAudit(db))
	{
		// ---------------------------------------------------------------------
		v1 := api.Group("v1")
		{
			healthController := controllers.NewHealthController()
			v1.GET("healthcheck", healthController.Status, middleware.SkipAudit)

			routes.RegisterAuthRoutes(v1, db)

			protectedGroup := v1.Group("")
			protectedGroup.Use(middleware.ApiAuthentication(db))
			{
				routes.RegisterUserRoutes(protectedGroup, db, middleware.SkipAudit)

				routes.RegisterAdminRoutes(protectedGroup, db)

				routes.RegisterStaticDataRoutes(protectedGroup, db, middleware.SkipAudit)

				// -------------------------------------------------------------
				packwizGroup := routes.RegisterPackwizRoutes(protectedGroup, db)

				routes.RegisterPackRoutes(packwizGroup, db)
			}
		}
	}

	// ---------------------------------------------------------------------
	embeddedSPAController := controllers.NewFrontendController(public.GetFrontendFiles())
	router.NoRoute(embeddedSPAController.Handler)

	return router
}
