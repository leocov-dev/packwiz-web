package dto

import (
	"github.com/go-playground/validator/v10"
)

type RehashRequest struct {
	Format string `json:"format" validate:"required,oneof=sha1 sha256 sha512"`
}

func (r *RehashRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(r)
}
