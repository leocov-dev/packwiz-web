package controllers

import (
	"github.com/gin-gonic/gin"
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
		err.JSON(c)
		return
	}

	minecraftVersions, err := loaders.GetMinecraftVersions()
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(
		c,
		VersionData{
			Minecraft: minecraftVersions,
			Loaders:   loaderData,
		},
	)
}
