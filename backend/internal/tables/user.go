package tables

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

type User struct {
	Id               uint           `gorm:"primarykey" json:"id"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Username         string         `gorm:"unique" json:"username"`
	FullName         string         `json:"fullName"`
	Email            string         `gorm:"unique" json:"email"`
	Password         string         `json:"-"`
	IsAdmin          bool           `json:"isAdmin"`
	IdentityProvider string         `json:"identityProvider"`
	LinkToken        string         `gorm:"unique" json:"-"`
	SessionKey       string         `json:"-"`
}

func (u User) CheckPassword(password string) bool {
	if u.Password == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
