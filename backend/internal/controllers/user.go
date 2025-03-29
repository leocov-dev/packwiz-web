package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/services/user_svc"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
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

// -----------------------------------------------------------------------------

func (uc *UserController) GetCurrentUser(c *gin.Context) {
	user, err := mustBindCurrentUser(c)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, user)
}

func (uc *UserController) ChangePassword(c *gin.Context) {
	user := c.MustGet("user").(tables.User)

	var form dto.ChangePasswordForm
	if err := mustBindData(c, &form); err != nil {
		err.JSON(c)
		return
	}

	if err := uc.svc.ChangePassword(user, form); err != nil {
		err.JSON(c)
		return
	}

	_ = uc.svc.InvalidateUserSessions(user.Id)
	_ = clearSession(c)

	isOK(c)
}

func (uc *UserController) UpdateUser(c *gin.Context) {
	currentUser, err := mustBindCurrentUser(c)
	if err != nil {
		err.JSON(c)
		return
	}

	var request dto.EditUserRequest
	if err := mustBindData(c, &request); err != nil {
		err.JSON(c)
		return
	}

	// admin can never change username
	if currentUser.Username == "admin" {
		request.Username = "admin"
	}

	if err := uc.svc.UpdateUser(currentUser.Id, request); err != nil {
		err.JSON(c)
		return
	}

	isOK(c)
}

func (uc *UserController) InvalidateCurrentUserSessions(c *gin.Context) {
	currentUser, err := mustBindCurrentUser(c)
	if err != nil {
		err.JSON(c)
		return
	}

	if err := uc.svc.InvalidateUserSessions(currentUser.Id); err != nil {
		err.JSON(c)
		return
	}

	currentUser.SessionKey = uc.svc.NewSessionKey(currentUser.Id)

	if err := newSession(c, currentUser); err != nil {
		err.JSON(c)
		return
	}

	isOK(c)
}
