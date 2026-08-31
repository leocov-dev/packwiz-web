package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/log"
	"packwiz-web/internal/params"
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

	packs, err := pc.packwizSvc.GetPacksWithPerms(query, user.ID)
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
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	c.Header("Content-Type", "application/json")

	if pc.packwizSvc.PackExists(packId, true) {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusNotFound)
	}
}

func (pc *PackwizController) GetOnePack(c *gin.Context) {
	log.Warn("GetOnePack")
	user, err := mustBindCurrentUser(c)
	if pc.abortWithError(c, err) {
		return
	}

	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, packId, true) {
		return
	}

	pack, err := pc.packwizSvc.GetPackWithPerms(packId, user.ID)
	if pc.abortWithError(c, err) {
		return
	}

	dataOK(c, pack)
}

func (pc *PackwizController) ArchivePack(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, packId, true) {
		return
	}

	err = pc.packwizSvc.ArchivePack(packId)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) UnArchivePack(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, packId, true) {
		return
	}

	err = pc.packwizSvc.UnArchivePack(packId)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) PublishPack(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, packId, false) {
		return
	}

	if pc.packwizSvc.IsPackPublished(packId) {
		c.JSON(http.StatusAccepted, gin.H{"msg": "pack is already published"})
		return
	}

	err = pc.packwizSvc.SetPackStatus(packId, types.PackStatusPublished)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) ConvertToDraft(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, packId, false) {
		return
	}

	if !pc.packwizSvc.IsPackPublished(packId) {
		c.JSON(http.StatusAccepted, gin.H{"msg": "pack is already a draft"})
		return
	}

	err = pc.packwizSvc.SetPackStatus(packId, types.PackStatusDraft)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) MakePublic(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, packId, false) {
		return
	}

	if pc.packwizSvc.IsPackPublicById(packId) {
		c.JSON(http.StatusAccepted, gin.H{"msg": "pack is already public"})
		return
	}

	err = pc.packwizSvc.MakePackPublic(packId)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) MakePrivate(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, packId, false) {
		return
	}

	if !pc.packwizSvc.IsPackPublicById(packId) {
		c.JSON(http.StatusAccepted, gin.H{"msg": "pack is already private"})
		return
	}

	err = pc.packwizSvc.MakePackPrivate(packId)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) EditPackInfo(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	if pc.abortIfPackNotExist(c, packId, false) {
		return
	}

	var request dto.EditPackRequest
	err = mustBindJson(c, &request)
	if pc.abortWithError(c, err) {
		return
	}

	err = pc.packwizSvc.EditPack(packId, request)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) UpdateAll(c *gin.Context) {
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

	err = pc.packwizSvc.UpdateAll(packId, user)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) MigratePack(c *gin.Context) {
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

	var request dto.MigratePackRequest
	err = mustBindJson(c, &request)
	if pc.abortWithError(c, err) {
		return
	}

	err = pc.packwizSvc.Migrate(packId, request, user)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) GetPersonalizedLink(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	user, err := mustBindCurrentUser(c)
	if pc.abortWithError(c, err) {
		return
	}

	pack, err := pc.packwizSvc.GetPackById(packId)
	if pc.abortWithError(c, err) {
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	link, err := pc.packwizSvc.GetPersonalLink(user, pack.ID, scheme, c.Request.Host)
	if pc.abortWithError(c, err) {
		return
	}

	dataOK(c, gin.H{"link": link.String()})
}

func (pc *PackwizController) GetPackUsers(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	users, err := pc.packwizSvc.ListPackUsers(packId)
	if pc.abortWithError(c, err) {
		return
	}

	if users == nil {
		users = make([]packwiz_svc.PackUserInfo, 0)
	}

	dataOK(c, gin.H{"users": users})
}

func (pc *PackwizController) SearchPackUsers(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	var request dto.SearchPackUsersQuery
	err = mustBindQuery(c, &request)
	if pc.abortWithError(c, err) {
		return
	}

	users, err := pc.packwizSvc.SearchPackUsers(packId, request.Query)
	if pc.abortWithError(c, err) {
		return
	}

	if users == nil {
		users = make([]packwiz_svc.PackUserSearchResult, 0)
	}

	dataOK(c, gin.H{"users": users})
}

func (pc *PackwizController) AddPackUser(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	var request dto.AddPackUserRequest
	err = mustBindJson(c, &request)
	if pc.abortWithError(c, err) {
		return
	}

	err = pc.packwizSvc.GrantPackUser(packId, request.UserID, request.Permission)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) RemovePackUser(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	userId, err := mustBindIdParam(c, params.UserID)
	if pc.abortWithError(c, err) {
		return
	}

	err = pc.packwizSvc.RevokePackUser(packId, userId)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
}

func (pc *PackwizController) EditUserAccess(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	userId, err := mustBindIdParam(c, params.UserID)
	if pc.abortWithError(c, err) {
		return
	}

	var request dto.EditUserAccessRequest
	err = mustBindJson(c, &request)
	if pc.abortWithError(c, err) {
		return
	}

	err = pc.packwizSvc.ChangePackUserPermission(packId, userId, request.Permission)
	if pc.abortWithError(c, err) {
		return
	}

	isOK(c)
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

func (pc *PackwizController) abortIfPackNotExist(c *gin.Context, packId uint, includeDeleted bool) bool {
	if !pc.packwizSvc.PackExists(packId, includeDeleted) {
		c.JSON(http.StatusNotFound, gin.H{"msg": fmt.Sprintf("pack %d not found", packId)})
		return true
	}
	return false
}

func (pc *PackwizController) abortIfModNotExist(c *gin.Context, packId, modId uint) bool {
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
