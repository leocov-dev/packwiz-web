package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"strconv"
)

func mustBindParam(c *gin.Context, name string) (status int, response interface{}) {
	value := c.Param(name)
	if value == "" {
		return http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("invalid %s", name)}
	}
	return 0, value
}

func mustBindIdParam(c *gin.Context, name string) (status int, response interface{}) {
	raw, err := strconv.Atoi(c.Param(name))
	if err != nil || raw <= 0 {
		return http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("invalid %s", name)}
	}
	return 0, uint(raw)
}

func mustBindCurrentUser(c *gin.Context) (status int, response tables.User) {
	return 0, c.MustGet("user").(tables.User)
}

func mustBindBody(c *gin.Context, request dto.Request) (status int, response interface{}) {
	if err := c.ShouldBindBodyWithJSON(request); err != nil {
		return http.StatusInternalServerError, gin.H{"msg": "failed to parse request"}
	}

	if err := request.Validate(); err != nil {
		return http.StatusBadRequest, gin.H{"msg": err.Error()}
	}

	return 0, request
}
