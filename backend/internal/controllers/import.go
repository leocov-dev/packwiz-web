package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"packwiz-web/internal/log"
	"packwiz-web/internal/services/import_svc"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

type ImportController struct {
	importSvc *import_svc.ImportService
}

func NewImportController(db *gorm.DB) *ImportController {
	return &ImportController{importSvc: import_svc.NewImportService(db)}
}

// ImportPack imports a live packwiz pack, given its pack.toml url, as a new pack.
func (ic *ImportController) ImportPack(c *gin.Context) {
	author, err := mustBindCurrentUser(c)
	if ic.abortWithError(c, err) {
		return
	}

	var request dto.ImportPackRequest
	err = mustBindJson(c, &request)
	if ic.abortWithError(c, err) {
		return
	}

	result, err := ic.importSvc.ImportPack(c.Request.Context(), request, author)
	if ic.abortWithError(c, err) {
		return
	}

	dataOK(c, result)
}

// abortWithError
// exit the request if the given error is not nil
func (ic *ImportController) abortWithError(c *gin.Context, err response.ServerError) bool {
	if err != nil {
		log.Debug(err)
		err.JSON(c)
		return true
	}
	return false
}
