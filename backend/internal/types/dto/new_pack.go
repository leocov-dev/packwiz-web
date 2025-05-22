package dto

import (
	"github.com/go-playground/validator/v10"
	"packwiz-web/internal/interfaces"
	"regexp"
)

type MinecraftDef struct {
	Version  string `json:"version" validate:"required_without=latest snapshot"`
	Latest   bool   `json:"latest" validate:"required_with=snapshot"`
	Snapshot bool   `json:"snapshot"`
}

func (m MinecraftDef) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(m)
}

type LoaderDef struct {
	Name    string `json:"name" validate:"required,oneof=fabric forge liteloader quilt neoforge"`
	Version string `json:"version"`
	Latest  bool   `json:"latest"`
}

func (l LoaderDef) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(l)
}

type NewPackRequest struct {
	Slug               string   `json:"slug" validate:"required,slug"`
	Name               string   `json:"name" validate:"required"`
	Version            string   `json:"version" validate:"required"`
	Description        string   `json:"description"`
	MinecraftVersion   string   `json:"minecraft" validate:"required"`
	LoaderName         string   `json:"loaderName" validate:"required"`
	LoaderVersion      string   `json:"loaderVersion" validate:"required"`
	AcceptableVersions []string `json:"acceptableVersions"`
}

var validSlugRegex = regexp.MustCompile(`^[a-zA-Z0-9\-._]+$`)

func (r NewPackRequest) Validate() error {
	errorGroup := interfaces.NewErrorGroup()
	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		return validSlugRegex.MatchString(fl.Field().String())
	}); err != nil {
		errorGroup.Add(err)
	}

	errorGroup.Add(validate.Struct(r))

	if errorGroup.IsEmpty() {
		return nil
	}
	return errorGroup
}
