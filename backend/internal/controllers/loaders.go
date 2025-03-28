package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"packwiz-web/internal/services/loaders"
)

type LoadersController struct {
}

func NewLoadersController() *LoadersController {
	return &LoadersController{}
}

type VersionData struct {
	Minecraft loaders.McVersionInfo `json:"minecraft"`
	Loaders   loaders.Loaders       `json:"loaders"`
}

// -----------------------------------------------------------------------------

func (lc *LoadersController) GetLoaderVersions(c *gin.Context) {
	loaderData, err := loaders.GetLoadersAndVersions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	minecraftVersions, err := loaders.GetMinecraftVersions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
	}

	versionData := VersionData{
		Minecraft: minecraftVersions,
		Loaders:   loaderData,
	}

	c.JSON(http.StatusOK, versionData)
}
