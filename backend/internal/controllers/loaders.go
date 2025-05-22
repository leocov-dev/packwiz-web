package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/leocov-dev/packwiz-nxt/core"
	"packwiz-web/internal/types/response"
)

type LoadersController struct {
}

func NewLoadersController() *LoadersController {
	return &LoadersController{}
}

type McVersionInfo struct {
	Latest         string   `json:"latest"`
	LatestSnapshot string   `json:"snapshot"`
	Versions       []string `json:"versions"`
}

type minecraftVersions struct {
	Latest         string   `json:"latest"`
	LatestSnapshot string   `json:"snapshot"`
	Versions       []string `json:"versions"`
}

type loaderData struct {
	Fabric     []string            `json:"fabric"`
	Forge      map[string][]string `json:"forge"`
	Liteloader []string            `json:"liteloader"`
	Quilt      []string            `json:"quilt"`
	Neoforge   map[string][]string `json:"neoforge"`
}

type VersionData struct {
	Minecraft minecraftVersions `json:"minecraft"`
	Loaders   loaderData        `json:"loaders"`
}

// -----------------------------------------------------------------------------

func (lc *LoadersController) GetLoaderVersions(c *gin.Context) {
	lv := core.GetLoaderCache()
	if lv.IsEmpty() {
		if err := lv.RefreshCache(); err != nil {
			response.Wrap(err).JSON(c)
			return
		}
	}

	mcv, err := core.GetMinecraftVersions()
	if err != nil {
		response.Wrap(err).JSON(c)
		return
	}

	dataOK(
		c,
		VersionData{
			Minecraft: minecraftVersions{
				Latest:         mcv.Latest,
				LatestSnapshot: mcv.LatestSnapshot,
				Versions:       mcv.Versions,
			},
			Loaders: loaderData{
				Fabric:     lv.Fabric,
				Forge:      lv.Forge,
				Liteloader: lv.Liteloader,
				Quilt:      lv.Quilt,
				Neoforge:   lv.Neoforge,
			},
		},
	)
}
