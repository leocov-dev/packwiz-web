package controllers

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
	"strconv"
)

func mustBindParam(c *gin.Context, name string) (string, response.ServerError) {
	value := c.Param(name)
	if value == "" {
		return "", response.New(http.StatusBadRequest, fmt.Sprintf("invalid %s", name))
	}
	return value, nil
}

func mustBindIdParam(c *gin.Context, name string) (uint, response.ServerError) {
	raw, err := strconv.Atoi(c.Param(name))
	if err != nil || raw <= 0 {
		return 0, response.New(http.StatusBadRequest, fmt.Sprintf("invalid %s", name))
	}
	return uint(raw), nil
}

func mustBindQuery(c *gin.Context, request dto.Request) response.ServerError {
	if err := c.ShouldBindQuery(request); err != nil {
		return response.Wrap(err)
	}

	if err := request.Validate(); err != nil {
		return response.New(
			http.StatusBadRequest,
			fmt.Sprintf("invalid request: %s", err.Error()),
		)
	}

	return nil
}

func mustBindCurrentUser(c *gin.Context) (tables.User, response.ServerError) {
	user, ok := c.Get("user")
	if !ok {
		return tables.User{}, response.New(http.StatusUnauthorized, "no user session")
	}

	return user.(tables.User), nil
}

func mustBindJson(c *gin.Context, request dto.Request) response.ServerError {
	if err := c.ShouldBindBodyWithJSON(request); err != nil {
		return response.Wrap(err)
	}

	if err := request.Validate(); err != nil {
		return response.New(
			http.StatusBadRequest,
			fmt.Sprintf("invalid request: %s", err.Error()),
		)
	}

	return nil
}

func mustBindForm(c *gin.Context, request dto.Request) response.ServerError {
	if err := c.ShouldBindWith(request, binding.Form); err != nil {
		return response.Wrap(err)
	}

	if err := request.Validate(); err != nil {
		return response.New(
			http.StatusBadRequest,
			fmt.Sprintf("invalid request: %s", err.Error()),
		)
	}

	return nil
}

func clearSession(c *gin.Context) response.ServerError {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		return response.New(
			http.StatusInternalServerError,
			"failed to clear and save session",
		)
	}
	return nil
}

func newSession(c *gin.Context, user tables.User) response.ServerError {
	session := sessions.Default(c)
	session.Set("userId", user.Id)
	session.Set("sessionKey", user.SessionKey)
	if err := session.Save(); err != nil {
		return response.New(http.StatusInternalServerError, "failed to save session")
	}
	return nil
}

func isOK(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func dataOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}
