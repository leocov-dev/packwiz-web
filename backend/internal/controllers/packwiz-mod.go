package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/log"
	"packwiz-web/internal/params"
	"packwiz-web/internal/services/packwiz_svc"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

type PackwizModController struct {
	packwizSvc *packwiz_svc.PackwizService
}

func NewPackwizModController(db *gorm.DB) *PackwizModController {
	return &PackwizModController{packwizSvc: packwiz_svc.NewPackwizService(db)}
}

func (pc *PackwizModController) AddMod(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, packId, false) {
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

	err = pc.packwizSvc.AddMod(packId, request, user)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizModController) ListMissingDependencies(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, packId, false) {
		return
	}

	var request dto.AddModRequest
	err = mustBindJson(c, &request)
	if pc.abortWithError(c, err) {
		return
	}

	missing, err := pc.packwizSvc.GetMissingModDependencies(packId, request)
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

func (pc *PackwizModController) SearchModrinthMods(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, packId, false) {
		return
	}

	var query dto.SearchModsQuery
	err = mustBindQuery(c, &query)
	if pc.abortWithError(c, err) {
		return
	}

	results, err := pc.packwizSvc.SearchModrinthProjects(query.Query, query.Versions)
	if pc.abortWithError(c, err) {
		return
	}

	dataOK(c, gin.H{"results": results})
}

func (pc *PackwizModController) GetOneMod(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	modId, err := mustBindIdParam(c, params.ModId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfModNotExist(c, packId, modId) {
		return
	}

	modData, err := pc.packwizSvc.GetMod(modId)
	if pc.abortWithError(c, err) {
		return
	}

	dataOK(c, &modData)
}

func (pc *PackwizModController) RemoveMod(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	modId, err := mustBindIdParam(c, params.ModId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfModNotExist(c, packId, modId) {
		return
	}

	err = pc.packwizSvc.RemoveModById(modId)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizModController) UpdateMod(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	modId, err := mustBindIdParam(c, params.ModId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfModNotExist(c, packId, modId) {
		return
	}

	user, err := mustBindCurrentUser(c)
	if pc.abortWithError(c, err) {
		return
	}

	err = pc.packwizSvc.UpdateMod(modId, user)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizModController) ChangeModSide(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	modId, err := mustBindIdParam(c, params.ModId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfModNotExist(c, packId, modId) {
		return
	}

	var request dto.ChangeModSideRequest
	err = mustBindJson(c, &request)
	if pc.abortWithError(c, err) {
		return
	}

	err = pc.packwizSvc.ChangeModSide(modId, request.Side)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizModController) PinMod(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	modId, err := mustBindIdParam(c, params.ModId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfModNotExist(c, packId, modId) {
		return
	}

	data, err := pc.packwizSvc.GetMod(modId)
	if pc.abortWithError(c, err) {
		return
	}

	if data.Pinned {
		c.JSON(http.StatusAccepted, gin.H{"msg": "mod is already pinned"})
		return
	}

	err = pc.packwizSvc.SetModPinnedValue(modId, true)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizModController) UnPinMod(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	modId, err := mustBindIdParam(c, params.ModId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfModNotExist(c, packId, modId) {
		return
	}

	data, err := pc.packwizSvc.GetMod(modId)
	if pc.abortWithError(c, err) {
		return
	}

	if !data.Pinned {
		c.JSON(http.StatusAccepted, gin.H{"msg": "mod is already unpinned"})
		return
	}

	err = pc.packwizSvc.SetModPinnedValue(modId, false)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

// -----------------------------------------------------------------------------

// abortWithError
// exit the request if the given error is not nil
func (pc *PackwizModController) abortWithError(c *gin.Context, err response.ServerError) bool {
	if err != nil {
		log.Debug(err)
		err.JSON(c)
		return true
	}
	return false
}

func (pc *PackwizModController) abortIfPackNotExist(c *gin.Context, packId uint, includeDeleted bool) bool {
	if !pc.packwizSvc.PackExists(packId, includeDeleted) {
		c.JSON(http.StatusNotFound, gin.H{"msg": fmt.Sprintf("pack %d not found", packId)})
		return true
	}
	return false
}

func (pc *PackwizModController) abortIfModNotExist(c *gin.Context, packId, modId uint) bool {
	if !pc.packwizSvc.PackExists(packId, false) {
		c.JSON(http.StatusNotFound, gin.H{"msg": fmt.Sprintf("pack %d not found", packId)})
		return true
	}

	if !pc.packwizSvc.ModExistsById(modId) {
		c.JSON(http.StatusNotFound, gin.H{"msg": fmt.Sprintf("pack %d with mod %d not found", packId, modId)})
		return true
	}
	return false
}
