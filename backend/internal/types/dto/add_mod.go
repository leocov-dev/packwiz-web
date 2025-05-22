package dto

import (
	"github.com/go-playground/validator/v10"
	"github.com/leocov-dev/packwiz-nxt/core"
)

type AddModrinth struct {
	Url string `json:"url" validate:"required,url"`
	//Name            string `json:"name"`
	//ProjectId       string `json:"projectId"`
	//VersionFilename string `json:"versionFilename"`
	//VersionId       string `json:"versionId"`
}

// Validate
// assert that AddModrinth is valid
func (r AddModrinth) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(r)
}

func (r AddModrinth) IsSet() bool {
	return r.Validate() == nil
}

type AddCurseforge struct {
	Url string `json:"url" validate:"required,url"`
	//Name     string `json:"name"`
	//AddonId  string `json:"addonId"`
	//Category string `json:"category"`
	//FileId   string `json:"fileId"`
	//Game     string `json:"game"`
}

// Validate
// assert that AddCurseforge is valid
func (r AddCurseforge) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(r)
}

func (r AddCurseforge) IsSet() bool {
	return r.Validate() == nil
}

// AddModRequest
// only one (Modrinth or Curseforge) should be specified
type AddModRequest struct {
	Modrinth   *AddModrinth   `json:"modrinth"`
	Curseforge *AddCurseforge `json:"curseforge"`
}

// Validate
// assert that the AddModRequest is valid
func (r AddModRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(r)
	//errorGroup := interfaces.NewErrorGroup()
	//
	//modrinthErr := r.Modrinth.Validate()
	//curseforgeErr := r.Curseforge.Validate()
	//
	//modrinthValid := modrinthErr == nil
	//curseforgeValid := curseforgeErr == nil
	//
	//if !(modrinthValid || curseforgeValid) || (modrinthValid && curseforgeValid) {
	//	errorGroup.Add(errors.New("only one of modrinth or curseforge can be specified"))
	//}
	//
	//if errorGroup.IsEmpty() {
	//	return nil
	//}
	//return errorGroup
}

type ModDependency struct {
	Slug     string       `json:"slug"`
	Name     string       `json:"name"`
	FileName string       `json:"fileName"`
	ModType  string       `json:"modType"`
	Side     core.ModSide `json:"side"`
	Url      string       `json:"url"`
}
