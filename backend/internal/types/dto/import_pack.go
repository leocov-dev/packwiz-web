package dto

import "github.com/go-playground/validator/v10"

// ImportPackRequest asks the server to import the live packwiz pack served at
// Url (the pack.toml address).
type ImportPackRequest struct {
	Url string `json:"url" validate:"required,http_url"`
}

func (r ImportPackRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(r)
}

// ImportPackResponse reports the new pack and everything that was left behind.
type ImportPackResponse struct {
	PackId       uint     `json:"packId"`
	Slug         string   `json:"slug"`
	Name         string   `json:"name"`
	ModsImported int      `json:"modsImported"`
	SkippedMods  []string `json:"skippedMods"`
	SkippedFiles []string `json:"skippedFiles"`
	Warnings     []string `json:"warnings"`
}
