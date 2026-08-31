package dto

import (
	"github.com/go-playground/validator/v10"
)

type MigratePackRequest struct {
	MinecraftDef       MinecraftDef `json:"minecraft" validate:"required"`
	LoaderDef          LoaderDef    `json:"loader" validate:"required"`
	UpdateMods         bool         `json:"updateMods"`
	UseRecommended     bool         `json:"useRecommended"`
	AcceptableVersions []string     `json:"acceptableVersions"`
}

func (r MigratePackRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(r)
}
