package dto

import (
	"github.com/go-playground/validator/v10"
)

type ListUsersQuery struct {
	NameSearch  string `form:"nameSearch"`
	EmailSearch string `form:"emailSearch"`
	UserType    string `form:"userType" validate:"omitempty,oneofci=admin user"`
	Page        int    `form:"page" validate:"gte=1"`
	PageSize    int    `form:"pageSize" validate:"gte=1,lte=100"`
}

func (f *ListUsersQuery) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}
