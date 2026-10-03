package dto

import (
	"github.com/go-playground/validator/v10"

	"packwiz-web/internal/tables"
)

// RoleRef is a role's id and name.
type RoleRef struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// AdminUserResponse is a user as seen in the admin screens, with system roles.
type AdminUserResponse struct {
	tables.User
	Roles []RoleRef `json:"roles"`
}

// SetUserRolesRequest replaces a user's system roles. An empty list clears them.
type SetUserRolesRequest struct {
	RoleIDs []uint `json:"roleIds" validate:"dive,gt=0"`
}

func (f *SetUserRolesRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}
