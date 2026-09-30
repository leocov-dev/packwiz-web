package dto

import (
	"time"

	"packwiz-web/internal/tables"
)

// CurrentUserResponse is the signed-in user's own view of their account.
type CurrentUserResponse struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Username    string    `json:"username"`
	FullName    string    `json:"fullName"`
	Email       string    `json:"email"`
	IsAdmin     bool      `json:"isAdmin"`
	IsActive    bool      `json:"isActive"`
	HasPassword bool      `json:"hasPassword"`
}

func NewCurrentUserResponse(u tables.User) CurrentUserResponse {
	return CurrentUserResponse{
		ID:          u.ID,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		Username:    u.Username,
		FullName:    u.FullName,
		Email:       u.Email,
		IsAdmin:     u.IsAdmin,
		IsActive:    u.IsActive,
		HasPassword: u.Password != "",
	}
}
