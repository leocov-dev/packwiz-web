package controllers

import (
	"mime"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"packwiz-web/internal/config"
	"packwiz-web/internal/middleware"
	"packwiz-web/internal/params"
	"packwiz-web/internal/services/multimc_svc"
	"packwiz-web/internal/services/packwiz_svc"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/response"
)

// MultiMCController serves importable MultiMC / Prism Launcher instance zips.
// It sits behind ConsumerAuthentication, next to pack.toml, so the zip URL
// works for a browser download and for a launcher's "import from URL" alike.
type MultiMCController struct {
	multimcSvc *multimc_svc.MultiMCService
}

func NewMultiMCController(db *gorm.DB) *MultiMCController {
	return &MultiMCController{
		multimcSvc: multimc_svc.NewMultiMCService(
			db,
			packwiz_svc.NewPackwizService(db, nil),
			config.C.PublicURL,
		),
	}
}

// DownloadInstance streams the instance zip for the authenticated consumer pack.
func (mc *MultiMCController) DownloadInstance(c *gin.Context) {
	pack, err := mustBindConsumerPack(c)
	if mc.abortWithError(c, err) {
		return
	}

	archive, err := mc.multimcSvc.BuildForConsumer(pack, c.Param(string(params.Token)), requestScheme(c), c.Request.Host)
	if mc.abortWithError(c, err) {
		return
	}

	// FormatMediaType falls back to RFC 2231 (filename*=utf-8'') for non-ASCII names
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": archive.Filename}))
	// the zip can embed a personal token, so it must never be cached
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "application/zip", archive.Data)
}

func mustBindConsumerPack(c *gin.Context) (tables.Pack, response.ServerError) {
	pack, ok := c.Get(middleware.ConsumerPackKey)
	if !ok {
		return tables.Pack{}, response.New(http.StatusNotFound, "pack not found")
	}
	return pack.(tables.Pack), nil
}

func (mc *MultiMCController) abortWithError(c *gin.Context, err response.ServerError) bool {
	if err != nil {
		err.JSON(c)
		return true
	}
	return false
}
