package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"packwiz-web/internal/log"
	"packwiz-web/internal/services/packwiz_cli"
	"packwiz-web/internal/services/packwiz_svc"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
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

	user := c.MustGet("user").(tables.User)

	var query dto.AllPacksQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	if err := query.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	packs, err := pc.packwizSvc.GetPacks(query.Status, query.Archived, query.Search, user.Id)
	if pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"packs": packs})
}

func (pc *PackwizController) NewPack(c *gin.Context) {
	author := c.MustGet("user").(tables.User)

	if !author.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"msg": "not authorized"})
		return
	}

	var request dto.NewPackRequest
	if err := c.ShouldBindBodyWithJSON(&request); err != nil {
		log.Error("Failed to bind request json", err)
		return
	}

	if err := request.Validate(); err != nil {
		log.Error("request validation failed", err)
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	if pc.packwizSvc.PackExists(request.Slug) {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "pack already exists"})
		return
	}

	err := pc.packwizSvc.NewPack(request, author)
	if pc.abortWithError(c, err) {
		log.Error("error creating new pack", err)
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

	user := c.MustGet("user").(tables.User)

	pack, err := pc.packwizSvc.GetPack(slug, user.Id, true, true)
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

	var request dto.AddModRequest
	if err := c.ShouldBindBodyWithJSON(&request); err != nil {
		log.Error("Failed to bind request json", err)
		return
	}

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

func (pc *PackwizController) ArchivePack(c *gin.Context) {
	slug := c.Param("slug")

	if pc.abortIfPackNotExist(c, slug) {
		return
	}

	err := pc.packwizSvc.ArchivePack(slug)
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

	var request dto.SetAcceptableVersionsRequest
	if err := c.ShouldBindBodyWithJSON(&request); err != nil {
		log.Error("Failed to bind request json", err)
		return
	}

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

	var request dto.ChangeModSideRequest
	if err := c.ShouldBindBodyWithJSON(&request); err != nil {
		log.Error("Failed to bind request json", err)
		return
	}

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

func (pc *PackwizController) GetPersonalizedLink(c *gin.Context) {
	slug := c.Param("slug")

	if pc.abortIfPackNotExist(c, slug) {
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	link, err := url.Parse(fmt.Sprintf("%s://%s/packwiz/%s/pack.toml", scheme, c.Request.Host, slug))
	if pc.abortWithError(c, err) {
		return
	}

	user := c.MustGet("user").(tables.User)

	if !pc.packwizSvc.IsPackPublic(slug) {
		query := link.Query()
		query.Add("token", user.LinkToken)
		link.RawQuery = query.Encode()

	}

	c.JSON(http.StatusOK, gin.H{"link": link.String()})
}

func (pc *PackwizController) GetPackUsers(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "not implemented"})
}

func (pc *PackwizController) AddPackUser(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "not implemented"})
}

func (pc *PackwizController) RemovePackUser(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "not implemented"})
}

func (pc *PackwizController) EditUserAccess(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "not implemented"})
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
