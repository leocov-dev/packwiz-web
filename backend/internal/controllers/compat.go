package controllers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"packwiz-web/internal/jobs"
	"packwiz-web/internal/log"
	"packwiz-web/internal/params"
	"packwiz-web/internal/services/compat_svc"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

// compatService is what CompatController needs from the compat service.
type compatService interface {
	GetCompat(ctx context.Context, packID uint) (compat_svc.Report, error)
	RequestManualScan(ctx context.Context, packID uint) (compat_svc.ManualScanResult, error)
}

// CompatController serves the jar compatibility check of a pack.
type CompatController struct {
	compatSvc compatService
}

// NewCompatController builds the controller around the compat service.
func NewCompatController(compatSvc compatService) *CompatController {
	return &CompatController{compatSvc: compatSvc}
}

// GetCompat returns the stored compatibility state of a pack. A pack that was
// never scanned has status unknown and no findings.
func (cc *CompatController) GetCompat(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if cc.abortWithError(c, err) {
		return
	}

	report, svcErr := cc.compatSvc.GetCompat(c.Request.Context(), packId)
	if cc.abortWithError(c, compatServerError(svcErr)) {
		return
	}

	dataOK(c, dto.NewCompatResponse(string(report.Status), report.Run, report.Findings))
}

// ScanCompat requests a rescan: 202 when a scan was enqueued, 200 with the
// current status when one is already active or ran too recently.
func (cc *CompatController) ScanCompat(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if cc.abortWithError(c, err) {
		return
	}

	result, svcErr := cc.compatSvc.RequestManualScan(c.Request.Context(), packId)
	if cc.abortWithError(c, compatServerError(svcErr)) {
		return
	}

	c.JSON(scanCompatStatus(result.Enqueued), dto.CompatScanResponse{
		Status:   string(result.Status),
		Enqueued: result.Enqueued,
	})
}

// scanCompatStatus maps the service outcome to an HTTP status: 202 when a scan
// was enqueued, 200 when the current state is returned instead.
func scanCompatStatus(enqueued bool) int {
	if enqueued {
		return http.StatusAccepted
	}
	return http.StatusOK
}

// compatServerError translates a compat service error to an HTTP error. It
// returns nil for a nil error.
func compatServerError(err error) response.ServerError {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return response.New(http.StatusNotFound, "pack not found")
	case errors.Is(err, compat_svc.ErrPackArchived):
		return response.New(http.StatusConflict, "pack is archived")
	case errors.Is(err, jobs.ErrClientUnavailable):
		return response.New(http.StatusServiceUnavailable, "compatibility check is not available")
	}
	log.Error(err)
	return response.Wrap(err)
}

// abortWithError exits the request if err is not nil.
func (cc *CompatController) abortWithError(c *gin.Context, err response.ServerError) bool {
	if err != nil {
		log.Debug(err)
		err.JSON(c)
		return true
	}
	return false
}
