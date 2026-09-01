package dto

import (
	"github.com/go-playground/validator/v10"
)

type ChangeModOptionRequest struct {
	Optional    bool   `json:"optional"`
	Description string `json:"description" validate:"max=500"`
	Default     bool   `json:"default"`
}

func (f *ChangeModOptionRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}
