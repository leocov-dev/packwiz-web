package dto

import (
	"github.com/go-playground/validator/v10"
)

type SetAcceptableVersionsRequest struct {
	Versions []string `json:"versions" validate:"required,min=1"`
}

// Validate
// set the acceptable minecraft versions for a modpack
func (r SetAcceptableVersionsRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(r)
}
