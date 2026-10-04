package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"packwiz-web/internal/log"
	"packwiz-web/internal/middleware"
	"packwiz-web/internal/params"
	"packwiz-web/internal/services/packwiz_svc"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

type PackSnapshotController struct {
	packwizSvc *packwiz_svc.PackwizService
}

func NewPackSnapshotController(db *gorm.DB) *PackSnapshotController {
	return &PackSnapshotController{packwizSvc: packwiz_svc.NewPackwizService(db, nil)}
}

// ListSnapshots lists a pack's history.
func (pc *PackSnapshotController) ListSnapshots(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	query := dto.PackSnapshotListQuery{Page: 1, PageSize: 25}
	err = mustBindQuery(c, &query)
	if pc.abortWithError(c, err) {
		return
	}

	result, err := pc.packwizSvc.ListSnapshots(packId, query)
	if pc.abortWithError(c, err) {
		return
	}

	dataOK(c, result)
}

// GetSnapshot returns one snapshot with its diff.
func (pc *PackSnapshotController) GetSnapshot(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	snapshotId, err := mustBindIdParam(c, params.SnapshotId)
	if pc.abortWithError(c, err) {
		return
	}

	var query dto.PackSnapshotDetailQuery
	err = mustBindQuery(c, &query)
	if pc.abortWithError(c, err) {
		return
	}

	result, err := pc.packwizSvc.GetSnapshot(packId, snapshotId, query.Against)
	if pc.abortWithError(c, err) {
		return
	}

	dataOK(c, result)
}

// RevertToSnapshot restores the pack to a snapshot.
func (pc *PackSnapshotController) RevertToSnapshot(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	snapshotId, err := mustBindIdParam(c, params.SnapshotId)
	if pc.abortWithError(c, err) {
		return
	}

	user, err := mustBindCurrentUser(c)
	if pc.abortWithError(c, err) {
		return
	}

	result, err := pc.packwizSvc.RevertToSnapshot(packId, snapshotId, user)
	if pc.abortWithError(c, err) {
		return
	}

	dataOK(c, result)
}

// CloneFromSnapshot creates a new pack from a snapshot.
func (pc *PackSnapshotController) CloneFromSnapshot(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	snapshotId, err := mustBindIdParam(c, params.SnapshotId)
	if pc.abortWithError(c, err) {
		return
	}

	user, err := mustBindCurrentUser(c)
	if pc.abortWithError(c, err) {
		return
	}

	var request dto.CloneSnapshotRequest
	err = mustBindJson(c, &request)
	if pc.abortWithError(c, err) {
		return
	}

	newPackId, err := pc.packwizSvc.CloneFromSnapshot(packId, snapshotId, request, user)
	if pc.abortWithError(c, err) {
		return
	}

	pack, err := pc.packwizSvc.GetPackWithPerms(newPackId, user, middleware.SystemPermissions(c))
	if pc.abortWithError(c, err) {
		return
	}

	dataOK(c, pack)
}

// PruneSnapshots permanently deletes the pack's abandoned snapshots.
func (pc *PackSnapshotController) PruneSnapshots(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	result, err := pc.packwizSvc.PruneSnapshots(packId)
	if pc.abortWithError(c, err) {
		return
	}

	dataOK(c, result)
}

// RebaseOnSnapshot makes a snapshot the root of the pack's history and
// permanently deletes everything before it.
func (pc *PackSnapshotController) RebaseOnSnapshot(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	snapshotId, err := mustBindIdParam(c, params.SnapshotId)
	if pc.abortWithError(c, err) {
		return
	}

	result, err := pc.packwizSvc.RebaseOnSnapshot(packId, snapshotId)
	if pc.abortWithError(c, err) {
		return
	}

	dataOK(c, result)
}

// abortWithError exits the request if err is not nil.
func (pc *PackSnapshotController) abortWithError(c *gin.Context, err response.ServerError) bool {
	if err != nil {
		log.Debug(err)
		err.JSON(c)
		return true
	}
	return false
}

// GetChangelist returns a pack's readable changelist.
func (pc *PackSnapshotController) GetChangelist(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	result, err := pc.packwizSvc.GetChangelist(packId)
	if pc.abortWithError(c, err) {
		return
	}

	dataOK(c, result)
}

// GetPendingChanges returns what publishing the pack would release.
func (pc *PackSnapshotController) GetPendingChanges(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if pc.abortWithError(c, err) {
		return
	}

	result, err := pc.packwizSvc.GetPendingChanges(packId)
	if pc.abortWithError(c, err) {
		return
	}

	dataOK(c, result)
}
