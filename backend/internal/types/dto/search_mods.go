package dto

import (
	"github.com/go-playground/validator/v10"
)

type SearchModsQuery struct {
	Query    string   `form:"q" validate:"required,min=2"`
	Versions []string `form:"versions"`
}

func (f *SearchModsQuery) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}

// ModSearchResult
// a minimal Modrinth project shape for the "search for a mod" picker
type ModSearchResult struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IconUrl     string `json:"iconUrl"`
	ProjectId   string `json:"projectId"`
}
