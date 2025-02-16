package server

import (
	"embed"
	"github.com/gin-gonic/gin"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
)

//go:embed public/*
var publicFiles embed.FS

// embeddedPublicHandler
// for all non-api routes or routes not captured by the packwiz_cli static files,
// we will try to load something from the pre-built single-page-application
// frontend that is bundled into the build by `embed`. the frontend build
// process must have run to generate these files.
func embeddedPublicHandler(c *gin.Context) {
	requestedPath := c.Request.URL.Path

	if requestedPath == "/" {
		requestedPath = "/index.html"
	}

	requestedPath = filepath.Join("public", requestedPath)

	info, err := fs.Stat(publicFiles, requestedPath)
	if err != nil || info.IsDir() {
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	file, err := publicFiles.Open(requestedPath)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, _ := file.Stat()
	ext := filepath.Ext(requestedPath)

	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.DataFromReader(http.StatusOK, stat.Size(), contentType, file, nil)
}
