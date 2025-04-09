package dto

import "github.com/go-playground/validator/v10"

type EditUserRequest struct {
	Username string `json:"username" validate:"required"`
	FullName string `json:"fullName" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

func (f *EditUserRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}
