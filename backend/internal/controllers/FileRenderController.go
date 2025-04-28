package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/log"
	"packwiz-web/internal/services/file_render"
	"strings"
)

type FileRenderController struct {
	db  *gorm.DB
	svc *file_render.FileRenderService
}

func NewFileRenderController(db *gorm.DB) *FileRenderController {
	return &FileRenderController{
		db:  db,
		svc: file_render.NewFileRenderService(db),
	}
}

func (frc *FileRenderController) RenderPackFile(c *gin.Context) {
	slug := c.Param("slug")

	frc.svc.UpdateHashes(slug)

	packFile, err := frc.svc.BuildPackFile(slug)
	if err != nil {
		log.Error(err)
		c.Status(http.StatusNotFound)
		return
	}

	c.TOML(http.StatusOK, packFile)
}

func (frc *FileRenderController) RenderIndexFile(c *gin.Context) {
	slug := c.Param("slug")

	frc.svc.UpdateHashes(slug)

	indexFile, err := frc.svc.BuildIndexFile(slug)
	if err != nil {
		log.Error(err)
		c.Status(http.StatusNotFound)
		return
	}

	c.TOML(http.StatusOK, indexFile)
}

func (frc *FileRenderController) RenderModFile(c *gin.Context) {
	slug := c.Param("slug")
	//modType := c.Param("type")
	modFileName := c.Param("mod")

	modSlug := strings.TrimSuffix(modFileName, ".pw.toml")

	modFile, err := frc.svc.BuildModFile(slug, modSlug)
	if err != nil {
		log.Error(err)
		c.Status(http.StatusNotFound)
		return
	}

	c.TOML(http.StatusOK, modFile)
}
