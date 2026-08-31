package dto

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"packwiz-web/internal/types"
)

func validPackPermission(permission types.PackPermission) bool {
	switch permission {
	case types.PackPermissionStatic, types.PackPermissionView, types.PackPermissionEdit:
		return true
	default:
		return false
	}
}

type AddPackUserRequest struct {
	UserID     uint                 `json:"userId" validate:"required"`
	Permission types.PackPermission `json:"permission" validate:"required"`
}

func (f *AddPackUserRequest) Validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(f); err != nil {
		return err
	}

	if !validPackPermission(f.Permission) {
		return fmt.Errorf("invalid permission: %d", f.Permission)
	}

	return nil
}

type EditUserAccessRequest struct {
	Permission types.PackPermission `json:"permission" validate:"required"`
}

func (f *EditUserAccessRequest) Validate() error {
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(f); err != nil {
		return err
	}

	if !validPackPermission(f.Permission) {
		return fmt.Errorf("invalid permission: %d", f.Permission)
	}

	return nil
}

type SearchPackUsersQuery struct {
	Query string `form:"q" validate:"required,min=2"`
}

func (f *SearchPackUsersQuery) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}
