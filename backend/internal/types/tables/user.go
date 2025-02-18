package tables

import (
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID               uint      `gorm:"primarykey" json:"id"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	Username         string    `gorm:"unique" json:"username"`
	Password         string    `json:"-"`
	IsAdmin          bool      `json:"isAdmin"`
	IdentityProvider string    `json:"identityProvider"`
}

func (u User) CheckPassword(password string) bool {
	if u.Password == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil

}
