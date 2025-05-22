package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/log"
	"packwiz-web/internal/services/packwiz_svc"
	"packwiz-web/internal/types"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

type PackwizController struct {
	packwizSvc *packwiz_svc.PackwizService
}

func NewPackwizController(db *gorm.DB) *PackwizController {
	return &PackwizController{packwizSvc: packwiz_svc.NewPackwizService(db)}
}

// -----------------------------------------------------------------------------

func (pc *PackwizController) GetAllPacks(c *gin.Context) {

	user, err := mustBindCurrentUser(c)
	if pc.abortWithError(c, err) {
		return
	}

	var query dto.AllPacksQuery

	err = mustBindQuery(c, &query)
	if pc.abortWithError(c, err) {
		return
	}

	packs, err := pc.packwizSvc.GetPacksWithPerms(query, user.Id)
	if pc.abortWithError(c, err) {
		return
	}

	if packs == nil {
		// get the response to return an empty array instead of null
		packs = make([]dto.PackResponse, 0)
	}

	allPacks := dto.AllPacksResponse{
		Packs: packs,
	}

	dataOK(c, allPacks)
}

func (pc *PackwizController) NewPack(c *gin.Context) {
	author, err := mustBindCurrentUser(c)
	if pc.abortWithError(c, err) {
		return
	}

	if !author.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"msg": "not authorized"})
		return
	}

	var request dto.NewPackRequest
	err = mustBindJson(c, &request)
	if pc.abortWithError(c, err) {
		return
	}

	err = pc.packwizSvc.NewPack(request, author)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) PackHead(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	c.Header("Content-Type", "application/json")

	if pc.packwizSvc.PackExists(slug, true) {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusNotFound)
	}
}

func (pc *PackwizController) GetOnePack(c *gin.Context) {
	user, err := mustBindCurrentUser(c)
	if pc.abortWithError(c, err) {
		return
	}

	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, slug, true) {
		return
	}

	var query dto.GetPackQuery
	err = mustBindQuery(c, &query)
	if pc.abortWithError(c, err) {
		return
	}

	pack, err := pc.packwizSvc.GetPackWithPerms(slug, user.Id)
	if pc.abortWithError(c, err) {
		return
	}

	dataOK(c, pack)
}

func (pc *PackwizController) AddMod(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, slug, false) {
		return
	}

	user, err := mustBindCurrentUser(c)
	if pc.abortWithError(c, err) {
		return
	}

	var request dto.AddModRequest
	err = mustBindJson(c, &request)
	if pc.abortWithError(c, err) {
		return
	}

	err = pc.packwizSvc.AddMod(slug, request, user)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) ListMissingDependencies(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, slug, false) {
		return
	}

	var request dto.AddModRequest
	err = mustBindJson(c, &request)
	if pc.abortWithError(c, err) {
		return
	}

	missing, err := pc.packwizSvc.GetMissingModDependencies(slug, request)
	if pc.abortWithError(c, err) {
		return
	}

	var data []dto.ModDependency

	for _, mod := range missing {
		data = append(data, dto.ModDependency{
			Slug:     mod.Slug,
			Name:     mod.Name,
			FileName: mod.FileName,
			ModType:  mod.ModType,
			Side:     mod.Side,
			Url:      mod.Download.URL,
		})
	}

	dataOK(c, gin.H{"missing": data})
}

func (pc *PackwizController) ArchivePack(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, slug, true) {
		return
	}

	err = pc.packwizSvc.ArchivePack(slug)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) UnArchivePack(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, slug, true) {
		return
	}

	err = pc.packwizSvc.UnArchivePack(slug)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) PublishPack(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, slug, false) {
		return
	}

	if pc.packwizSvc.IsPackPublished(slug) {
		c.JSON(http.StatusAccepted, gin.H{"msg": "pack is already published"})
		return
	}

	err = pc.packwizSvc.SetPackStatus(slug, types.PackStatusPublished)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) ConvertToDraft(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, slug, false) {
		return
	}

	if !pc.packwizSvc.IsPackPublished(slug) {
		c.JSON(http.StatusAccepted, gin.H{"msg": "pack is already a draft"})
		return
	}

	err = pc.packwizSvc.SetPackStatus(slug, types.PackStatusDraft)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) MakePublic(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, slug, false) {
		return
	}

	if pc.packwizSvc.IsPackPublic(slug) {
		c.JSON(http.StatusAccepted, gin.H{"msg": "pack is already public"})
		return
	}

	err = pc.packwizSvc.MakePackPublic(slug)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) MakePrivate(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, slug, false) {
		return
	}

	if !pc.packwizSvc.IsPackPublic(slug) {
		c.JSON(http.StatusAccepted, gin.H{"msg": "pack is already private"})
		return
	}

	err = pc.packwizSvc.MakePackPrivate(slug)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) EditPackInfo(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, slug, false) {
		return
	}

	var request dto.EditPackRequest
	err = mustBindJson(c, &request)
	if pc.abortWithError(c, err) {
		return
	}

	err = pc.packwizSvc.EditPack(slug, request)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) UpdateAll(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, slug, false) {
		return
	}

	err = pc.packwizSvc.UpdateAll(slug)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) GetOneMod(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	mod, err := mustBindParam(c, "mod")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfModNotExist(c, slug, mod) {
		return
	}

	modData, err := pc.packwizSvc.GetMod(slug, mod)
	if pc.abortWithError(c, err) {
		return
	}

	dataOK(c, &modData)
}

func (pc *PackwizController) RemoveMod(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	mod, err := mustBindParam(c, "mod")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfModNotExist(c, slug, mod) {
		return
	}

	err = pc.packwizSvc.RemoveMod(slug, mod)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) UpdateMod(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	mod, err := mustBindParam(c, "mod")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfModNotExist(c, slug, mod) {
		return
	}

	err = pc.packwizSvc.UpdateMod(slug, mod)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) ChangeModSide(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	mod, err := mustBindParam(c, "mod")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfModNotExist(c, slug, mod) {
		return
	}

	var request dto.ChangeModSideRequest
	err = mustBindJson(c, &request)
	if pc.abortWithError(c, err) {
		return
	}

	err = pc.packwizSvc.ChangeModSide(slug, mod, request.Side)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) PinMod(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	mod, err := mustBindParam(c, "mod")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfModNotExist(c, slug, mod) {
		return
	}

	data, err := pc.packwizSvc.GetMod(slug, mod)
	if pc.abortWithError(c, err) {
		return
	}

	if data.Pinned {
		c.JSON(http.StatusAccepted, gin.H{"msg": "mod is already pinned"})
		return
	}

	err = pc.packwizSvc.SetModPinnedValue(slug, mod, true)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) UnPinMod(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	mod, err := mustBindParam(c, "mod")
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfModNotExist(c, slug, mod) {
		return
	}

	data, err := pc.packwizSvc.GetMod(slug, mod)
	if pc.abortWithError(c, err) {
		return
	}

	if !data.Pinned {
		c.JSON(http.StatusAccepted, gin.H{"msg": "mod is already unpinned"})
		return
	}

	err = pc.packwizSvc.SetModPinnedValue(slug, mod, false)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) GetPersonalizedLink(c *gin.Context) {
	slug, err := mustBindParam(c, "slug")
	if pc.abortWithError(c, err) {
		return
	}

	user, err := mustBindCurrentUser(c)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, slug, false) {
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	link, err := pc.packwizSvc.GetPersonalLink(user, slug, scheme, c.Request.Host)
	if pc.abortWithError(c, err) {
		return
	}

	dataOK(c, gin.H{"link": link.String()})
}

func (pc *PackwizController) GetPackUsers(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"msg": "not implemented"})
}

func (pc *PackwizController) AddPackUser(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"msg": "not implemented"})
}

func (pc *PackwizController) RemovePackUser(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"msg": "not implemented"})
}

func (pc *PackwizController) EditUserAccess(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"msg": "not implemented"})
}

// -----------------------------------------------------------------------------

// abortWithError
// exit the request if the given error is not nil
func (pc *PackwizController) abortWithError(c *gin.Context, err response.ServerError) bool {
	if err != nil {
		log.Debug(err)
		err.JSON(c)
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
