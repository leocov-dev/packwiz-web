package dto

import "time"

// PublicPackMod is the trimmed mod view shown on the public pack page.
type PublicPackMod struct {
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Side         string `json:"side"`
	Version      string `json:"version,omitempty"`
	Source       string `json:"source"`
	Optional     bool   `json:"optional"`
	IsDependency bool   `json:"isDependency"`
}

// PublicPackResponse is the unauthenticated view of a public, published pack.
// It lists only what the pack's own toml files already expose.
type PublicPackResponse struct {
	Slug                   string    `json:"slug"`
	Name                   string    `json:"name"`
	Description            string    `json:"description"`
	Author                 string    `json:"author"`
	Version                string    `json:"version"`
	MCVersion              string    `json:"mcVersion"`
	Loader                 string    `json:"loader"`
	LoaderVersion          string    `json:"loaderVersion"`
	AcceptableGameVersions []string  `json:"acceptableGameVersions"`
	PackFormat             string    `json:"packFormat"`
	UpdatedAt              time.Time `json:"updatedAt"`
	PackTomlURL            string    `json:"packTomlUrl"`
	// MultiMCURL is the importable MultiMC / Prism instance zip.
	MultiMCURL string          `json:"multimcUrl"`
	Mods       []PublicPackMod `json:"mods"`
}
