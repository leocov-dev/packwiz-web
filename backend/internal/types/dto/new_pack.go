package dto

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"packwiz-web/internal/interfaces"
	"packwiz-web/internal/services/packwiz_cli"
	"regexp"
)

type MinecraftDef struct {
	Version  string `json:"version"`
	Latest   bool   `json:"latest"`
	Snapshot bool   `json:"snapshot"`
}

func (m MinecraftDef) AsCliType() packwiz_cli.MinecraftDef {
	return packwiz_cli.MinecraftDef{
		Version:  m.Version,
		Latest:   m.Latest,
		Snapshot: m.Snapshot,
	}
}

func (m MinecraftDef) Validate() error {
	if m.Version == "" && !m.Latest {
		return errors.New("version or latest must be set")
	}
	if m.Snapshot && !m.Latest {
		return errors.New("snapshot must be set with latest")
	}
	return nil
}

type LoaderDef struct {
	Name    packwiz_cli.LoaderType `json:"name" validate:"required"`
	Version string                 `json:"version"`
	Latest  bool                   `json:"latest"`
}

func (l LoaderDef) AsCliType() packwiz_cli.LoaderDef {
	return packwiz_cli.LoaderDef{
		Name:    l.Name,
		Version: l.Version,
		Latest:  l.Latest,
	}
}

func (l LoaderDef) Validate() error {
	if l.Version == "" && !l.Latest {
		return errors.New("version or latest must be set")
	}
	return nil
}

type NewPackRequest struct {
	Slug               string       `json:"slug" validate:"required,slug"`
	Name               string       `json:"name" validate:"required"`
	Version            string       `json:"version" validate:"required"`
	Description        string       `json:"description"`
	MinecraftDef       MinecraftDef `json:"minecraft" validate:"required"`
	LoaderDef          LoaderDef    `json:"loader" validate:"required"`
	AcceptableVersions []string     `json:"acceptableVersions"`
}

var validSlugRegex = regexp.MustCompile(`^[a-zA-Z0-9\-._]+$`)

func (r NewPackRequest) Validate() error {
	errorGroup := interfaces.NewErrorGroup()
	validate := validator.New()

	if err := validate.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		return validSlugRegex.MatchString(fl.Field().String())
	}); err != nil {
		errorGroup.Add(err)
	}

	errorGroup.Add(r.MinecraftDef.Validate())
	errorGroup.Add(r.LoaderDef.Validate())
	errorGroup.Add(validate.Struct(r))

	if errorGroup.IsEmpty() {
		return nil
	}
	return errorGroup
}
