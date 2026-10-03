package dto

import (
	"github.com/go-playground/validator/v10"
	"packwiz-web/internal/tables"
)

type CreateUserRequest struct {
	Username string `json:"username" validate:"required"`
	FullName string `json:"fullName" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"omitempty,min=12,max=64"`
}

// CreateUserResponse is returned once on creation; GeneratedPassword is never
// retrievable afterwards.
type CreateUserResponse struct {
	tables.User
	GeneratedPassword string `json:"generatedPassword,omitempty"`
}

type ResetPasswordResponse struct {
	Password string `json:"password"`
}

func (f *CreateUserRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}
