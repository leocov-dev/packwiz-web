package controllers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/riverqueue/river"
	"gorm.io/gorm"
	"strings"

	"packwiz-web/internal/params"
	"packwiz-web/internal/services/packwiz_svc"
	"packwiz-web/internal/types/response"
)

// PublicPackController serves the unauthenticated public pack page data.
type PublicPackController struct {
	packwizSvc *packwiz_svc.PackwizService
}

func NewPublicPackController(db *gorm.DB, riverClient *river.Client[*sql.Tx]) *PublicPackController {
	return &PublicPackController{packwizSvc: packwiz_svc.NewPackwizService(db, riverClient)}
}

// GetPublicPack returns the public view of a pack, or 404 if it is not public.
func (pc *PublicPackController) GetPublicPack(c *gin.Context) {
	slug, err := mustBindParam(c, params.PackSlug)
	if pc.abortWithError(c, err) {
		return
	}

	pack, err := pc.packwizSvc.GetPublicPack(slug, requestScheme(c), c.Request.Host)
	if pc.abortWithError(c, err) {
		return
	}

	dataOK(c, pack)
}

func (pc *PublicPackController) abortWithError(c *gin.Context, err response.ServerError) bool {
	if err != nil {
		err.JSON(c)
		return true
	}
	return false
}

// requestScheme is the scheme the client used, honoring a reverse proxy.
func requestScheme(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = strings.Split(proto, ",")[0]
	}
	return scheme
}
