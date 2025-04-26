package tables

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"packwiz-web/internal/types"
	"time"
)

type Pack struct {
	Slug                   string                      `gorm:"primarykey" json:"slug"`
	Name                   string                      `json:"name"`
	CreatedAt              time.Time                   `json:"createdAt"`
	UpdatedAt              time.Time                   `json:"updatedAt"`
	DeletedAt              gorm.DeletedAt              `gorm:"index" json:"deletedAt"`
	Description            string                      `json:"description"`
	CreatedBy              uint                        `json:"createdBy"`
	UpdatedBy              uint                        `json:"updatedBy"`
	IsPublic               bool                        `json:"isPublic"`
	Status                 types.PackStatus            `gorm:"default:draft" json:"status"`
	MCVersion              string                      `json:"mcVersion"`
	Loader                 string                      `json:"loader"`
	LoaderVersion          string                      `json:"loaderVersion"`
	AcceptableGameVersions datatypes.JSONSlice[string] `json:"acceptableGameVersions"`

	Mods []Mod `json:"mods"`

	Version    string `json:"version"`
	PackFormat string `json:"packFormat"`
	Hash       string `json:"hash"`
	HashFormat string `json:"hashFormat"`

	// hydrated fields
	IsArchived bool                 `gorm:"-" json:"isArchived"` // hydrated based on DeletedAt
	Permission types.PackPermission `gorm:"-" json:"permission"` // hydrated with current PackUsers.Permission
}
