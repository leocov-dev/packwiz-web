package dto

import (
	"errors"
	"packwiz-web/internal/interfaces"
)

type AddModrinth struct {
	Name            string `json:"name"`
	ProjectId       string `json:"projectId"`
	VersionFilename string `json:"versionFilename"`
	VersionId       string `json:"versionId"`
}

// Validate
// assert that AddModrinth is valid
func (r AddModrinth) Validate() error {
	if r.Name == "" && r.ProjectId == "" && r.VersionFilename == "" && r.VersionId == "" {
		return errors.New("invalid modrinth data")
	}
	return nil
}

func (r AddModrinth) IsSet() bool {
	return r.Validate() == nil
}

type AddCurseforge struct {
	Name     string `json:"name"`
	AddonId  string `json:"addonId"`
	Category string `json:"category"`
	FileId   string `json:"fileId"`
	Game     string `json:"game"`
}

// Validate
// assert that AddCurseforge is valid
func (r AddCurseforge) Validate() error {
	if r.Name == "" && r.AddonId == "" && r.Category == "" && r.FileId == "" && r.Game == "" {
		return errors.New("invalid curseforge data")
	}
	return nil
}

func (r AddCurseforge) IsSet() bool {
	return r.Validate() == nil
}

// AddModRequest
// only one (Modrinth or Curseforge) should be specified
type AddModRequest struct {
	Modrinth   AddModrinth   `json:"modrinth"`
	Curseforge AddCurseforge `json:"curseforge"`
}

// Validate
// assert that the AddModRequest is valid
func (r AddModRequest) Validate() error {
	errorGroup := interfaces.NewErrorGroup()

	modrinthErr := r.Modrinth.Validate()
	curseforgeErr := r.Curseforge.Validate()

	modrinthValid := modrinthErr == nil
	curseforgeValid := curseforgeErr == nil

	if !(modrinthValid || curseforgeValid) || (modrinthValid && curseforgeValid) {
		errorGroup.Add(errors.New("only one of modrinth or curseforge can be specified"))
	}

	if errorGroup.IsEmpty() {
		return nil
	}
	return errorGroup
}
