package dto

import (
	"github.com/go-playground/validator/v10"
)

type EditPackRequest struct {
	Name               string       `json:"name" validate:"required"`
	Version            string       `json:"version" validate:"required"`
	Description        string       `json:"description"`
	MinecraftDef       MinecraftDef `json:"minecraft" validate:"required"`
	LoaderDef          LoaderDef    `json:"loader" validate:"required"`
	AcceptableVersions []string     `json:"acceptableVersions"`
}

func (r EditPackRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(r)
}
