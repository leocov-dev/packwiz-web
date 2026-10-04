package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"packwiz-web/internal/params"
	"packwiz-web/internal/services/pack_access_svc"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

const defaultAccessDays = 30

type PackAccessController struct {
	svc *pack_access_svc.PackAccessService
}

func NewPackAccessController(db *gorm.DB) *PackAccessController {
	return &PackAccessController{svc: pack_access_svc.NewPackAccessService(db)}
}

// PackSeries returns a pack's daily successful-access counts.
func (pc *PackAccessController) PackSeries(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if err != nil {
		err.JSON(c)
		return
	}

	query := dto.PackAccessSummaryQuery{Days: 7}
	if err := mustBindQuery(c, &query); err != nil {
		err.JSON(c)
		return
	}

	series, err := pc.svc.PackSeries(packId, query.Days)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, series)
}

// PackSummary returns a pack's access metrics (successful requests only).
func (pc *PackAccessController) PackSummary(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if err != nil {
		err.JSON(c)
		return
	}

	query := dto.PackAccessSummaryQuery{Days: defaultAccessDays}
	if err := mustBindQuery(c, &query); err != nil {
		err.JSON(c)
		return
	}

	summary, err := pc.svc.PackSummary(packId, query.Days)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, summary)
}

// PackRecent pages through a pack's successful accesses.
func (pc *PackAccessController) PackRecent(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if err != nil {
		err.JSON(c)
		return
	}

	query := dto.PackAccessRecentQuery{Days: defaultAccessDays, Page: 1, PageSize: 25}
	if err := mustBindQuery(c, &query); err != nil {
		err.JSON(c)
		return
	}

	records, total, err := pc.svc.PackRecent(packId, query)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, response.NewPaginated(records, query.Page, query.PageSize, total))
}

// SystemSummary returns system-wide pack.toml access metrics, including failures.
func (pc *PackAccessController) SystemSummary(c *gin.Context) {
	pc.systemSummary(c, pack_access_svc.SourcePackToml)
}

// SystemRecent pages through pack.toml accesses of all packs, including failures.
func (pc *PackAccessController) SystemRecent(c *gin.Context) {
	pc.systemRecent(c, pack_access_svc.SourcePackToml)
}

// SystemDownloadsSummary returns system-wide instance zip download metrics.
func (pc *PackAccessController) SystemDownloadsSummary(c *gin.Context) {
	pc.systemSummary(c, pack_access_svc.SourceInstanceZip)
}

// SystemDownloadsRecent pages through instance zip downloads of all packs.
func (pc *PackAccessController) SystemDownloadsRecent(c *gin.Context) {
	pc.systemRecent(c, pack_access_svc.SourceInstanceZip)
}

func (pc *PackAccessController) systemSummary(c *gin.Context, source pack_access_svc.Source) {
	query := dto.PackAccessSummaryQuery{Days: defaultAccessDays}
	if err := mustBindQuery(c, &query); err != nil {
		err.JSON(c)
		return
	}

	summary, err := pc.svc.SystemSummary(source, query.Days)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, summary)
}

func (pc *PackAccessController) systemRecent(c *gin.Context, source pack_access_svc.Source) {
	query := dto.PackAccessRecentQuery{Days: defaultAccessDays, Page: 1, PageSize: 25}
	if err := mustBindQuery(c, &query); err != nil {
		err.JSON(c)
		return
	}

	records, total, err := pc.svc.SystemRecent(source, query)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, response.NewPaginated(records, query.Page, query.PageSize, total))
}
