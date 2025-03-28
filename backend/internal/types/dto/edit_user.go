package dto

import (
	"github.com/go-playground/validator/v10"
)

type EditUserRequest struct {
	Username string `json:"username" validate:"required"`
	FullName string `json:"fullName" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

func (r EditUserRequest) Validate() error {
	validate := validator.New()

	return validate.Struct(r)
}
