package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"packwiz-web/internal/config"
)

type StaticDataController struct {
}

func NewStaticDataController() *StaticDataController {
	return &StaticDataController{}
}

func (sdc *StaticDataController) GetStaticData(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		gin.H{
			"version": config.C.Version,
		},
	)
}
