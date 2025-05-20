package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/log"
	"packwiz-web/internal/params"
	"packwiz-web/internal/services/packwiz_svc"
	"strings"
)

type SerializeController struct {
	pw *packwiz_svc.PackwizService
}

func NewTomlController(db *gorm.DB) SerializeController {
	return SerializeController{
		pw: packwiz_svc.NewPackwizService(db),
	}
}

func (tc *SerializeController) RenderPackToml(c *gin.Context) {
	packSlug := c.Param(string(params.PackSlug))
	if packSlug == "" {
		c.Status(http.StatusNotFound)
		return
	}

	pack, err := tc.pw.GetPack(packSlug)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	packMeta := pack.AsMeta()
	packToml, tomlerr := packMeta.AsPackToml()
	if tomlerr != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Render(http.StatusOK, TomlStr{text: packToml})
}
func (tc *SerializeController) RenderIndexToml(c *gin.Context) {
	packSlug := c.Param(string(params.PackSlug))
	if packSlug == "" {
		c.Status(http.StatusNotFound)
		return
	}

	pack, err := tc.pw.GetPack(packSlug)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	packMeta := pack.AsMeta()
	indexToml, _, tomlErr := packMeta.AsIndexToml()
	if tomlErr != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Render(http.StatusOK, TomlStr{text: indexToml})
}

func (tc *SerializeController) RenderModToml(c *gin.Context) {
	log.Debug(c.Params)
	packSlug := c.Param(string(params.PackSlug))
	modSlug := strings.TrimSuffix(c.Param(string(params.ModSlug)), ".pw.toml")
	if packSlug == "" || modSlug == "" {
		log.Warn("invalid mod or slug", packSlug, modSlug)
		c.Status(http.StatusNotFound)
		return
	}

	pack, err := tc.pw.GetPack(packSlug)
	if err != nil {
		log.Warn("pack not found", packSlug, err)
		c.Status(http.StatusNotFound)
		return
	}

	packMeta := pack.AsMeta()
	modToml, tomlerr := packMeta.AsModToml(modSlug)
	if tomlerr != nil {
		log.Warn("error", tomlerr)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Render(http.StatusOK, TomlStr{text: modToml})
}

type TomlStr struct {
	text string
}

func (r TomlStr) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)

	_, err := w.Write([]byte(r.text))
	return err
}

func (r TomlStr) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"text/plain; charset=utf-8"}
	}
}
