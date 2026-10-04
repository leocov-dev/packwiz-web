package server

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"packwiz-web/internal/config"
	"packwiz-web/internal/controllers"
	"packwiz-web/internal/database"
	"packwiz-web/internal/jobs"
	"packwiz-web/internal/log"
	"packwiz-web/internal/middleware"
	"packwiz-web/internal/params"
	"packwiz-web/internal/routes"
	"packwiz-web/internal/services/audit_svc"
	"packwiz-web/internal/services/oidc_svc"
	"packwiz-web/internal/services/pack_access_svc"
	"packwiz-web/internal/services/packwiz_svc"
	"packwiz-web/public"
	"strings"
	"time"
)

// Per-IP limits on consumer files. Unlike most of this app these are enforced
// by the server rather than left to a reverse proxy: pack.toml starts every
// sync, and the instance zip is generated on demand.
const (
	packTomlLimit    = "30-M"
	instanceZipLimit = "10-M"
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

	// jobResolver is only used to satisfy jobs.MigrateModsResolver for the
	// (unstarted, insert-only) client below - it never needs to enqueue jobs
	// itself, so it's built without a river client of its own.
	jobResolver := packwiz_svc.NewPackwizService(db, nil)
	packAccessSvc := pack_access_svc.NewPackAccessService(db)
	riverClient, err := jobs.NewClient(db, jobs.NewWorkers(jobResolver, jobResolver, audit_svc.NewAuditService(db), packAccessSvc))
	if err != nil {
		log.Error("failed to create river client:", err)
		panic(err)
	}

	// -------------------------------------------------------------------------
	packwizFiles := router.Group(fmt.Sprintf("packwiz/:%s/:%s", params.Token, params.PackSlug))
	// pack.toml is the entry point of every sync, so it is limited, ahead of
	// audit and auth so rejected requests count too. index.toml and the per-mod
	// files are fetched in bulk by the installer and stay unlimited; this is
	// load protection, not a defense against token guessing.
	packwizFiles.Use(middleware.RateLimitIf(packTomlLimit, func(c *gin.Context) bool {
		return strings.HasSuffix(c.FullPath(), "/pack.toml")
	}))
	packwizFiles.Use(middleware.RateLimitIf(instanceZipLimit, func(c *gin.Context) bool {
		return middleware.IsInstanceZipRoute(c)
	}))
	// audit runs first so it also records requests authentication rejects
	packwizFiles.Use(middleware.PackwizAudit(packAccessSvc))
	packwizFiles.Use(middleware.ConsumerAuthentication(db))
	{
		tomlController := controllers.NewTomlController(db)
		packwizFiles.GET("pack.toml", tomlController.RenderPackToml)
		packwizFiles.GET("index.toml", tomlController.RenderIndexToml)
		packwizFiles.GET(
			fmt.Sprintf("%s/:%s", params.InstanceZipDir, params.InstanceFile),
			controllers.NewMultiMCController(db).DownloadInstance,
		)
		packwizFiles.GET(fmt.Sprintf(":%s/:%s", params.ModType, params.ModSlug), tomlController.RenderModToml)
	}

	// shared so a provider edit invalidates the client cache the login flow uses
	oidcProviders := oidc_svc.NewProviderService(db, config.C.SessionSecret, config.C.PublicURL)
	oidcFlow := oidc_svc.NewFlowService(db, oidcProviders, config.C.SessionSecret)

	// -------------------------------------------------------------------------
	api := router.Group("api", middleware.SessionStore(), middleware.ApiAudit(db))
	{
		// ---------------------------------------------------------------------
		v1 := api.Group("v1")
		{
			healthController := controllers.NewHealthController()
			v1.GET("healthcheck", healthController.Status, middleware.SkipAudit)

			routes.RegisterAuthRoutes(v1, db)
			routes.RegisterOidcAuthRoutes(v1, db, oidcProviders, oidcFlow)
			routes.RegisterPublicRoutes(v1, db, riverClient)

			protectedGroup := v1.Group("")
			protectedGroup.Use(middleware.ApiAuthentication(db))
			{
				routes.RegisterUserRoutes(protectedGroup, db, middleware.SkipAudit)

				routes.RegisterOidcIdentityRoutes(protectedGroup, db, oidcFlow)

				routes.RegisterAdminRoutes(protectedGroup, db)
				routes.RegisterOidcAdminRoutes(protectedGroup, db, oidcProviders)

				routes.RegisterStaticDataRoutes(protectedGroup, db, middleware.SkipAudit)

				// -------------------------------------------------------------
				packwizGroup := routes.RegisterPackwizRoutes(protectedGroup, db)

				routes.RegisterPackRoutes(packwizGroup, db, riverClient)
			}
		}
	}

	// ---------------------------------------------------------------------
	embeddedSPAController := controllers.NewFrontendController(public.GetFrontendFiles())
	router.NoRoute(embeddedSPAController.Handler)

	return router
}
