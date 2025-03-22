package dto

import (
	"github.com/go-playground/validator/v10"
	"packwiz-web/internal/interfaces"
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
	errorGroup := interfaces.NewErrorGroup()
	validate := validator.New()

	errorGroup.Add(r.MinecraftDef.Validate())
	errorGroup.Add(r.LoaderDef.Validate())
	errorGroup.Add(validate.Struct(r))

	if errorGroup.IsEmpty() {
		return nil
	}
	return errorGroup
}
