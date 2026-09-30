package dto

import "github.com/go-playground/validator/v10"

// SetPasswordForm sets a first local password for an account that has none
// (e.g. one created through OIDC).
type SetPasswordForm struct {
	NewPassword string `form:"newPassword" validate:"required,min=12,max=64"`
}

func (f *SetPasswordForm) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}
