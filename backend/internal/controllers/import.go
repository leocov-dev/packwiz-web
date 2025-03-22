package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

type ImportController struct {
	db *gorm.DB
}

func NewImportController(db *gorm.DB) *ImportController {
	return &ImportController{
		db: db,
	}
}

func (ic *ImportController) UploadPackwizArchive(c *gin.Context) {
	// TODO
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "not implemented"})
}
