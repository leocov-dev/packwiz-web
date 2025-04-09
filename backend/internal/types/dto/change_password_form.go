package dto

import "github.com/go-playground/validator/v10"

type ChangePasswordForm struct {
	OldPassword string `form:"oldPassword" validate:"required"`
	NewPassword string `form:"newPassword" validate:"required,min=12,max=64"`
}

func (f *ChangePasswordForm) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}
