package user_svc

import (
	"fmt"
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
	"packwiz-web/internal/utils"
	"regexp"
	"strings"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) FindById(id uint) (tables.User, error) {
	var user tables.User
	if err := s.db.Where("id == ?", id).First(&user).Error; err != nil {
		return user, err
	}

	return user, nil
}

func (s *UserService) FindByUsername(username string) (tables.User, error) {
	var user tables.User
	if err := s.db.Where("username == ?", username).First(&user).Error; err != nil {
		return user, err
	}
	return user, nil
}

func (s *UserService) ChangePassword(user tables.User, form dto.ChangePasswordForm) response.ServerError {
	if user.Username == "admin" {
		return response.New(
			http.StatusBadRequest,
			"admin password can only be set via PWW_ADMIN_PASSWORD env var",
		)
	}

	oldPassword := strings.TrimSpace(form.OldPassword)
	newPassword := strings.TrimSpace(form.NewPassword)

	if !user.CheckPassword(oldPassword) {
		return response.New(
			http.StatusBadRequest,
			"Invalid current password",
		)
	}

	if !s.CheckPasswordComplexity(newPassword) {
		return response.New(
			http.StatusBadRequest,
			"Password must contain at least one letter and one number",
		)
	}

	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return response.New(
			http.StatusInternalServerError,
			"Failed to hash password",
		)
	}

	if err := s.db.
		Model(&tables.User{}).
		Where("id == ?", user.Id).
		Update("password", hashed).Error; err != nil {
		return response.New(
			http.StatusInternalServerError,
			"Failed to update db password",
		)
	}

	return nil
}

func (s *UserService) CheckPasswordLength(password string, minLen, maxLen uint) bool {
	length := uint(len(password))
	return length >= minLen && length <= maxLen
}

var passwordCheck = regexp.MustCompile(`^[A-Za-z\d\s!@#$%^&*()-+={}\[\]|:;"'<>,.?/~` + "`" + `]{8,}$`)
var letterCheck = regexp.MustCompile(`[A-Za-z]`)
var numberCheck = regexp.MustCompile(`\d`)

func (s *UserService) CheckPasswordComplexity(password string) bool {
	hasLetter := letterCheck.MatchString(password)
	hasNumber := numberCheck.MatchString(password)
	formatCheck := passwordCheck.MatchString(password)
	return hasLetter && hasNumber && formatCheck
}

func (s *UserService) GetOrMakeSessionKey(user *tables.User) string {
	if user.SessionKey == "" {
		user.SessionKey = s.NewSessionKey(user.Id)
	}
	return user.SessionKey
}

func (s *UserService) NewSessionKey(userId uint) string {
	newKey := utils.GenerateRandomString(32)

	// ignore the error, prioritize letting the user log in, this session
	// will be invalidated on a fresh login
	if err := s.db.Model(&tables.User{}).
		Where("id == ?", userId).
		Update("session_key", newKey).Error; err != nil {
		log.Error("Failed to update session key, ", err)
	}

	return newKey
}

func (s *UserService) InvalidateUserSessions(userId uint) response.ServerError {
	if err := s.db.Model(&tables.User{}).
		Where("id == ?", userId).
		Update("session_key", "").Error; err != nil {
		return response.New(
			http.StatusInternalServerError,
			"Failed to invalidate db sessions")
	}
	return nil
}

func (s *UserService) UpdateUser(userId uint, request dto.EditUserRequest) response.ServerError {
	user, err := s.FindById(userId)
	if err != nil {
		return response.New(http.StatusNotFound, fmt.Sprintf("user %d not found", userId))
	}

	if user.Username == "admin" {
		request.Username = "admin"
	}

	// there can be only one admin
	if strings.ToLower(request.Username) == "admin" && user.Username != "admin" {
		return response.New(http.StatusBadRequest, "all forms of 'admin' username are reserved")
	}

	user.Username = request.Username
	user.FullName = request.FullName
	user.Email = request.Email

	if err := s.db.Save(&user).Error; err != nil {
		return response.New(http.StatusInternalServerError, "failed to update db user")
	}

	return nil
}

func (s *UserService) ListUsers(request dto.ListUsersQuery) ([]tables.User, int64, response.ServerError) {
	var users []tables.User
	var total int64

	offset := (request.Page - 1) * request.PageSize

	if err := s.db.Transaction(func(tx *gorm.DB) error {

		query := tx.Model(&tables.User{})

		switch strings.ToLower(request.UserType) {
		case "admin":
			query.Where("is_admin = ?", true)
		case "user":
			query.Where("is_admin = ?", false)
		}

		if err := query.Count(&total).Error; err != nil {
			return err
		}

		if err := query.Offset(offset).Limit(request.PageSize).Scan(&users).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, 0, response.New(http.StatusInternalServerError, "failed to list users")
	}

	if users == nil {
		users = []tables.User{}
	}

	return users, total, nil
}
