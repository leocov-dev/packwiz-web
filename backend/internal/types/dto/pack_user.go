package dto

import (
	"github.com/go-playground/validator/v10"
)

type AddPackUserRequest struct {
	UserID uint `json:"userId" validate:"required"`
	RoleID uint `json:"roleId" validate:"required"`
}

func (f *AddPackUserRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}

type EditUserAccessRequest struct {
	RoleID uint `json:"roleId" validate:"required"`
}

func (f *EditUserAccessRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}

type SearchPackUsersQuery struct {
	Query string `form:"q" validate:"required,min=2"`
}

func (f *SearchPackUsersQuery) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}
