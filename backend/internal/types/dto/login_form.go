package dto

import "github.com/go-playground/validator/v10"

type LoginForm struct {
	Username string `form:"username" validate:"required"`
	Password string `form:"password" validate:"required"`
}

func (f *LoginForm) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}
