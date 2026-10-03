package dto

import (
	"time"

	"packwiz-web/internal/tables"
)

// CurrentUserResponse is the signed-in user's own view of their account.
type CurrentUserResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Username  string    `json:"username"`
	FullName  string    `json:"fullName"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"isActive"`
	// IsSuperuser is display-only; the UI gates on Permissions.
	IsSuperuser bool `json:"isSuperuser"`
	HasPassword bool `json:"hasPassword"`
	// Permissions are the system-level permission names the user holds.
	Permissions []string `json:"permissions"`
}

func NewCurrentUserResponse(u tables.User, permissions []string) CurrentUserResponse {
	if permissions == nil {
		permissions = []string{}
	}
	return CurrentUserResponse{
		ID:          u.ID,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		Username:    u.Username,
		FullName:    u.FullName,
		Email:       u.Email,
		IsActive:    u.IsActive,
		IsSuperuser: u.IsSuperuser,
		HasPassword: u.Password != "",
		Permissions: permissions,
	}
}
