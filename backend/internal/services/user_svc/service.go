package user_svc

import (
	"gorm.io/gorm"
	"packwiz-web/internal/tables"
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
