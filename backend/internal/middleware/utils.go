package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"packwiz-web/internal/params"
	"packwiz-web/internal/types/response"
	"strconv"
)

func mustBindIdParam(c *gin.Context, name params.Param) (uint, response.ServerError) {
	raw, err := strconv.Atoi(c.Param(string(name)))
	if err != nil || raw <= 0 {
		return 0, response.New(http.StatusBadRequest, fmt.Sprintf("invalid %s", name))
	}
	return uint(raw), nil
}
