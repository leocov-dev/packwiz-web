package controllers

import (
	"embed"
	"github.com/gin-gonic/gin"
	"io/fs"
	"mime"
	"net/http"
	"packwiz-web/internal/logger"
	"path/filepath"
	"strings"
)

type FrontendController struct {
	frontendFiles *embed.FS
}

func NewFrontendController(publicFiles *embed.FS) *FrontendController {
	return &FrontendController{
		frontendFiles: publicFiles,
	}
}

// Handler
// for all non-api routes or routes not captured by the packwiz_cli static files,
// we will try to load something from the pre-built single-page-application
// frontend that is bundled into the build by `embed`. the frontend build
// process must have run to generate these files.
func (epc *FrontendController) Handler(c *gin.Context) {
	requestedPath := c.Request.URL.Path

	if strings.HasPrefix(requestedPath, "/api") {
		logger.Warn("api route requested, aborting: ", requestedPath)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if requestedPath == "/" {
		requestedPath = "/index.html"
	}

	requestedPath = filepath.Join("frontend", requestedPath)

	info, err := fs.Stat(epc.frontendFiles, requestedPath)
	if err != nil || info.IsDir() {
		logger.Warn("file info error", requestedPath, err)
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	file, err := epc.frontendFiles.Open(requestedPath)
	if err != nil {
		logger.Warn("file open error", requestedPath, err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, _ := file.Stat()
	ext := filepath.Ext(requestedPath)

	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		logger.Warn("no content type found for extension:", ext)
		contentType = "application/octet-stream"
	}

	logger.Debug("serving file:", requestedPath, "with content type:", contentType)

	c.DataFromReader(http.StatusOK, stat.Size(), contentType, file, nil)
}
