package dto

import "github.com/go-playground/validator/v10"

type CreateUserRequest struct {
	Username string `json:"username" validate:"required"`
	FullName string `json:"fullName" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=12,max=64"`
	IsAdmin  bool   `json:"isAdmin"`
}

func (f *CreateUserRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}
