package auth_svc

import (
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/log"
	"packwiz-web/internal/services/user_svc"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

type AuthService struct {
	db   *gorm.DB
	user *user_svc.UserService
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		db:   db,
		user: user_svc.NewUserService(db),
	}
}

// -----------------------------------------------------------------------------

func (as *AuthService) Login(form dto.LoginForm) (tables.User, response.ServerError) {

	user, err := as.user.FindByUsername(form.Username)
	if err != nil {
		log.Warn(
			"login failed",
			"username", form.Username,
			"error", err)
		return tables.User{}, response.New(http.StatusBadRequest, "Invalid username or password")
	}

	if !user.CheckPassword(form.Password) {
		log.Warn(
			"login failed",
			"username", form.Username,
			"bad password")
		return tables.User{}, response.New(http.StatusBadRequest, "Invalid username or password")
	}

	return user, nil
}
