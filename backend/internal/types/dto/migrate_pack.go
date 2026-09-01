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

// MigrateDryRunResponse reports, per mod, whether Migrate would find a
// compatible update for the requested Minecraft version / loader target,
// without persisting anything.
type MigrateDryRunResponse struct {
	Mods []MigrateDryRunMod `json:"mods"`
}

type MigrateDryRunMod struct {
	ModId           uint   `json:"modId"`
	Slug            string `json:"slug"`
	Name            string `json:"name"`
	Pinned          bool   `json:"pinned"`
	UpdateAvailable bool   `json:"updateAvailable"`
	UpdateString    string `json:"updateString,omitempty"`
	// Incompatible is true if resolving this mod against the candidate target
	// failed (e.g. no matching version was found), rather than the mod simply
	// having no update available.
	Incompatible bool   `json:"incompatible"`
	Error        string `json:"error,omitempty"`
}
