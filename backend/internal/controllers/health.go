package controllers

import (
	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// Status
// @Summary service healthcheck
// @Produce json
// @Router /health [get]
func (hc *HealthController) Status(c *gin.Context) {
	isOK(c)
}
