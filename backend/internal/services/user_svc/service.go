package user_svc

import (
	"gorm.io/gorm"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/utils"
	"regexp"
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

func (s *UserService) ChangePassword(user tables.User, newPassword string) error {
	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.db.
		Model(&tables.User{}).
		Where("id == ?", user.Id).
		Update("password", hashed).Error
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
		user.SessionKey = utils.GenerateRandomString(32)

		// ignore the error, prioritize letting the user log in, this session
		// will be invalidated on a fresh login
		if err := s.db.Model(&tables.User{}).
			Where("id == ?", user.Id).
			Update("session_key", user.SessionKey).Error; err != nil {
			log.Error("Failed to update session key, ", err)
		}
	}
	return user.SessionKey
}

func (s *UserService) InvalidateUserSessions(userId uint) error {
	return s.db.Model(&tables.User{}).
		Where("id == ?", userId).
		Update("session_key", "").Error
}

func (s *UserService) UpdateUser(userId uint, request dto.EditUserRequest) (tables.User, error) {
	user, err := s.FindById(userId)
	if err != nil {
		return tables.User{}, err
	}

	user.Username = request.Username
	user.FullName = request.FullName
	user.Email = request.Email

	if err := s.db.Save(&user).Error; err != nil {
		return tables.User{}, err
	}

	return user, nil
}
