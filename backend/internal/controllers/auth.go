package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/services/auth_svc"
	"packwiz-web/internal/services/user_svc"
	"packwiz-web/internal/types/dto"
)

type AuthController struct {
	db   *gorm.DB
	user *user_svc.UserService
	auth *auth_svc.AuthService
}

func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{
		db:   db,
		user: user_svc.NewUserService(db),
		auth: auth_svc.NewAuthService(db),
	}
}

// -----------------------------------------------------------------------------

func (ac *AuthController) Login(c *gin.Context) {

	var form dto.LoginForm
	if err := mustBindForm(c, &form); err != nil {
		err.JSON(c)
		return
	}

	user, err := ac.auth.Login(form)
	if err != nil {
		err.JSON(c)
		return
	}

	ac.user.GetOrMakeSessionKey(&user)

	if err := newSession(c, user); err != nil {
		err.JSON(c)
		return
	}

	isOK(c)
}

func (ac *AuthController) Logout(c *gin.Context) {
	if err := clearSession(c); err != nil {
		err.JSON(c)
		return
	}

	isOK(c)
}
