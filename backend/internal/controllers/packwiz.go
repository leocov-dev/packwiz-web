package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"packwiz-web/internal/log"
	"packwiz-web/internal/services/packwiz_svc"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types"
	"packwiz-web/internal/types/dto"
)

type PackwizController struct {
	packwizSvc *packwiz_svc.PackwizService
}

func NewPackwizController(db *gorm.DB) *PackwizController {
	return &PackwizController{packwizSvc: packwiz_svc.NewPackwizService(db)}
}

// -----------------------------------------------------------------------------

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

	if pc.packwizSvc.PackExists(request.Slug, true) {
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

	if pc.packwizSvc.PackExists(slug, true) {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusNotFound)
	}
}

func (pc *PackwizController) GetOnePack(c *gin.Context) {
	slug := c.Param("slug")

	if pc.abortIfPackNotExist(c, slug, true) {
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

	if pc.abortIfPackNotExist(c, slug, false) {
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

	if pc.abortIfPackNotExist(c, slug, true) {
		return
	}

	if err := pc.packwizSvc.ArchivePack(slug); pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (pc *PackwizController) UnArchivePack(c *gin.Context) {
	slug := c.Param("slug")

	if pc.abortIfPackNotExist(c, slug, true) {
		return
	}

	if err := pc.packwizSvc.UnArchivePack(slug); pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (pc *PackwizController) PublishPack(c *gin.Context) {
	slug := c.Param("slug")
	if pc.abortIfPackNotExist(c, slug, false) {
		return
	}

	if pc.packwizSvc.IsPackPublished(slug) {
		c.JSON(http.StatusAccepted, gin.H{"msg": "pack is already published"})
		return
	}

	if err := pc.packwizSvc.SetPackStatus(slug, types.PackStatusPublished); pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (pc *PackwizController) ConvertToDraft(c *gin.Context) {
	slug := c.Param("slug")
	if pc.abortIfPackNotExist(c, slug, false) {
		return
	}

	if !pc.packwizSvc.IsPackPublished(slug) {
		c.JSON(http.StatusAccepted, gin.H{"msg": "pack is already a draft"})
		return
	}

	if err := pc.packwizSvc.SetPackStatus(slug, types.PackStatusDraft); pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (pc *PackwizController) MakePublic(c *gin.Context) {
	slug := c.Param("slug")
	if pc.abortIfPackNotExist(c, slug, false) {
		return
	}

	if pc.packwizSvc.IsPackPublic(slug) {
		c.JSON(http.StatusAccepted, gin.H{"msg": "pack is already public"})
		return
	}

	if err := pc.packwizSvc.MakePackPublic(slug); pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (pc *PackwizController) MakePrivate(c *gin.Context) {
	slug := c.Param("slug")
	if pc.abortIfPackNotExist(c, slug, false) {
		return
	}

	if !pc.packwizSvc.IsPackPublic(slug) {
		c.JSON(http.StatusAccepted, gin.H{"msg": "pack is already private"})
		return
	}

	if err := pc.packwizSvc.MakePackPrivate(slug); pc.abortWithError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (pc *PackwizController) EditPackInfo(c *gin.Context) {
	slug := c.Param("slug")

	if pc.abortIfPackNotExist(c, slug, false) {
		return
	}

	var request dto.EditPackRequest
	if err := c.ShouldBindBodyWithJSON(&request); err != nil {
		log.Error("Failed to bind request json", err)
		return
	}

	if err := request.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	// TODO

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (pc *PackwizController) UpdateAll(c *gin.Context) {
	slug := c.Param("slug")

	if pc.abortIfPackNotExist(c, slug, false) {
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

	if pc.abortIfPackNotExist(c, slug, false) {
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
	c.JSON(http.StatusInternalServerError, gin.H{"msg": "not implemented"})
}

// -----------------------------------------------------------------------------

// abortWithError
// exit the request if the given error is not nil
func (pc *PackwizController) abortWithError(c *gin.Context, err error) bool {
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return true
	}
	return false
}

func (pc *PackwizController) abortIfPackNotExist(c *gin.Context, slug string, includeDeleted bool) bool {
	if !pc.packwizSvc.PackExists(slug, includeDeleted) {
		c.JSON(http.StatusNotFound, gin.H{"msg": fmt.Sprintf("pack %s not found", slug)})
		return true
	}
	return false
}

func (pc *PackwizController) abortIfModNotExist(c *gin.Context, slug string, mod string) bool {
	if !pc.packwizSvc.PackExists(slug, false) {
		c.JSON(http.StatusNotFound, gin.H{"msg": fmt.Sprintf("pack %s not found", slug)})
		return true
	}

	if !pc.packwizSvc.ModExists(slug, mod) {
		c.JSON(http.StatusNotFound, gin.H{"msg": fmt.Sprintf("pack %s with mod %s not found", slug, mod)})
		return true
	}
	return false
}
