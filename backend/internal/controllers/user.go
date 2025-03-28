package controllers

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"packwiz-web/internal/services/user_svc"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"strings"

	"gorm.io/gorm"
	"net/http"
)

type UserController struct {
	db  *gorm.DB
	svc *user_svc.UserService
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{
		db:  db,
		svc: user_svc.NewUserService(db),
	}
}

func (uc *UserController) CurrentUser(c *gin.Context) {
	user := c.MustGet("user").(tables.User)
	c.JSON(http.StatusOK, user)
}

func (uc *UserController) ChangePassword(c *gin.Context) {
	user := c.MustGet("user").(tables.User)

	if user.Username == "admin" {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "admin password set via PWW_ADMIN_PASSWORD env var",
		})
		return
	}

	type ChangePasswordForm struct {
		OldPassword string `form:"oldPassword" binding:"required"`
		NewPassword string `form:"newPassword" binding:"required"`
	}

	var form ChangePasswordForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid form data"})
		return
	}

	form.OldPassword = strings.TrimSpace(form.OldPassword)
	form.NewPassword = strings.TrimSpace(form.NewPassword)

	if !user.CheckPassword(form.OldPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid current password"})
		return
	}

	if !uc.svc.CheckPasswordLength(form.NewPassword, 10, 64) {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "Password must be from 10 t0 64 characters long",
		})
		return
	}

	if !uc.svc.CheckPasswordComplexity(form.NewPassword) {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "Password must contain at least one letter and one number",
		})
		return
	}

	if err := uc.svc.ChangePassword(user, form.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Failed to change password, internal error",
		})
		return
	}

	session := sessions.Default(c)
	session.Clear()
	_ = session.Save()

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (uc *UserController) UpdateUser(c *gin.Context) {
	currentUser := c.MustGet("user").(tables.User)

	var request dto.EditUserRequest
	if err := c.ShouldBindBodyWithJSON(&request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": fmt.Sprintf("Failed to parse request: %s", err)})
		return
	}

	if err := request.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	// there can be only one admin
	if strings.ToLower(request.Username) == "admin" && currentUser.Username != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "all forms of 'admin' username are reserved"})
		return
	}

	// admin can never change username
	if currentUser.Username == "admin" {
		request.Username = "admin"
	}

	updatedUser, err := uc.svc.UpdateUser(currentUser.Id, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

func (uc *UserController) InvalidateCurrentUserSessions(c *gin.Context) {
	currentUser := c.MustGet("user").(tables.User)
	currentUser.SessionKey = ""

	if err := uc.svc.InvalidateUserSessions(currentUser.Id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to invalidate sessions"})
		return
	}

	session := sessions.Default(c)
	session.Set("userId", currentUser.Id)
	session.Set("sessionKey", uc.svc.GetOrMakeSessionKey(&currentUser))
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to save session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}
