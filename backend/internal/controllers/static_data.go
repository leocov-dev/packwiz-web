package controllers

import (
	"github.com/gin-gonic/gin"
	"packwiz-web/internal/config"
)

type StaticDataController struct{}

func NewStaticDataController() *StaticDataController {
	return &StaticDataController{}
}

func (sdc *StaticDataController) GetStaticData(c *gin.Context) {
	dataOK(c, gin.H{
		"version": config.C.Version,
	})
}
