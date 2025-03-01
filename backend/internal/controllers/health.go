package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
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
	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}
