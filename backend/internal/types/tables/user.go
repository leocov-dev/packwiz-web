package tables

import (
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID               uint `gorm:"primarykey"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Username         string `gorm:"unique"`
	Password         string `json:"-"`
	IsAdmin          bool
	IdentityProvider string
}

func (u User) CheckPassword(password string) bool {
	if u.Password == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil

}
