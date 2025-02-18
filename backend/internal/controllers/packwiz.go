package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/services/packwiz_cli"
	"packwiz-web/internal/services/packwiz_svc"
	dto2 "packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/tables"
)

type PackwizController struct {
	packwizSvc *packwiz_svc.PackwizService
}

func NewPackwizController(db *gorm.DB) *PackwizController {
	return &PackwizController{packwizSvc: packwiz_svc.NewPackwizService(db)}
}

// -----------------------------------------------------------------------------

// ListLoaders
// list the possible loader options
func (pc *PackwizController) ListLoaders(c *gin.Context) {
	loaders := []string{
		string(packwiz_cli.FabricLoader),
		string(packwiz_cli.ForgeLoader),
		string(packwiz_cli.LiteLoader),
		string(packwiz_cli.QuiltLoader),
		string(packwiz_cli.NeoForgeLoader),
	}

	c.JSON(http.StatusOK, gin.H{"loaders": loaders})
}

func (pc *PackwizController) UploadPackwizArchive(c *gin.Context) {
	// TODO
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "not implemented"})
}

func (pc *PackwizController) GetAllPacks(c *gin.Context) {
	packs, err := pc.packwizSvc.GetPacks()
	if pc.abortWithError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"packs": packs})
}

func (pc *PackwizController) NewPack(c *gin.Context) {
	author := c.MustGet("user").(tables.User)

	var request dto2.NewPackRequest
	c.BindJSON(&request)

	if err := request.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	if pc.packwizSvc.PackExists(request.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "pack already exists"})
		return
	}

	err := pc.packwizSvc.NewPack(request, author)
	if pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (pc *PackwizController) PackHead(c *gin.Context) {
	slug := c.Param("slug")

	c.Header("Content-Type", "application/json")

	if pc.packwizSvc.PackExists(slug) {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusNotFound)
	}
}

func (pc *PackwizController) RenamePack(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "not implemented"})
}

func (pc *PackwizController) GetOnePack(c *gin.Context) {
	slug := c.Param("slug")

	if pc.abortIfPackNotExist(c, slug) {
		return
	}

	pack, err := pc.packwizSvc.GetPack(slug, true, true)
	if pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, pack)
}
func (pc *PackwizController) AddMod(c *gin.Context) {
	slug := c.Param("slug")

	if pc.abortIfPackNotExist(c, slug) {
		return
	}

	var request dto2.AddModRequest
	c.BindJSON(&request)

	if err := request.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	err := pc.packwizSvc.AddMod(slug, request)
	if pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (pc *PackwizController) RemovePack(c *gin.Context) {
	slug := c.Param("slug")

	if pc.abortIfPackNotExist(c, slug) {
		return
	}

	err := pc.packwizSvc.RemovePack(slug)
	if pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (pc *PackwizController) SetAcceptableVersions(c *gin.Context) {
	slug := c.Param("slug")

	if pc.abortIfPackNotExist(c, slug) {
		return
	}

	var request dto2.SetAcceptableVersionsRequest
	c.BindJSON(&request)

	if err := request.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	err := pc.packwizSvc.SetAcceptableVersions(slug, request)
	if pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}
func (pc *PackwizController) UpdateAll(c *gin.Context) {
	slug := c.Param("slug")

	if pc.abortIfPackNotExist(c, slug) {
		return
	}

	err := pc.packwizSvc.UpdateAll(slug)
	if pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (pc *PackwizController) RemoveMod(c *gin.Context) {
	slug := c.Param("slug")
	mod := c.Param("mod")
	if pc.abortIfModNotExist(c, slug, mod) {
		return
	}

	err := pc.packwizSvc.RemoveMod(slug, mod)
	if pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (pc *PackwizController) RenameMod(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "not implemented"})
}

func (pc *PackwizController) UpdateMod(c *gin.Context) {
	slug := c.Param("slug")
	mod := c.Param("mod")
	if pc.abortIfModNotExist(c, slug, mod) {
		return
	}

	data, err := pc.packwizSvc.GetMod(slug, mod)
	if pc.abortWithError(c, err) {
		return
	}

	if data.Pinned {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "mod is pinned"})
		return
	}

	err = pc.packwizSvc.UpdateMod(slug, mod)
	if pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (pc *PackwizController) ChangeModSide(c *gin.Context) {
	slug := c.Param("slug")
	mod := c.Param("mod")
	if pc.abortIfModNotExist(c, slug, mod) {
		return
	}

	var request dto2.ChangeModSideRequest
	c.BindJSON(&request)

	if err := request.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	err := pc.packwizSvc.ChangeModSide(slug, mod, request.Side)
	if pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (pc *PackwizController) PinMod(c *gin.Context) {
	slug := c.Param("slug")
	mod := c.Param("mod")
	if pc.abortIfModNotExist(c, slug, mod) {
		return
	}

	data, err := pc.packwizSvc.GetMod(slug, mod)
	if pc.abortWithError(c, err) {
		return
	}

	if data.Pinned {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "mod is already pinned"})
		return
	}

	err = pc.packwizSvc.PinMod(slug, mod)
	if pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (pc *PackwizController) UnPinMod(c *gin.Context) {
	slug := c.Param("slug")
	mod := c.Param("mod")
	if pc.abortIfModNotExist(c, slug, mod) {
		return
	}

	data, err := pc.packwizSvc.GetMod(slug, mod)
	if pc.abortWithError(c, err) {
		return
	}

	if !data.Pinned {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "mod is already unpinned"})
		return
	}

	err = pc.packwizSvc.UnpinMod(slug, mod)
	if pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

// -----------------------------------------------------------------------------

// abortWithError
// exit the request if the given error is not nil
func (pc *PackwizController) abortWithError(c *gin.Context, err error) bool {
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return true
	}
	return false
}

func (pc *PackwizController) abortIfPackNotExist(c *gin.Context, slug string) bool {
	if !pc.packwizSvc.PackExists(slug) {
		c.JSON(http.StatusNotFound, gin.H{"msg": fmt.Sprintf("pack %s not found", slug)})
		return true
	}
	return false
}

func (pc *PackwizController) abortIfModNotExist(c *gin.Context, slug string, mod string) bool {
	if !pc.packwizSvc.ModExists(slug, mod) {
		c.JSON(http.StatusNotFound, gin.H{"msg": fmt.Sprintf("pack %s with mod %s not found", slug, mod)})
		return true
	}
	return false
}
