package tables

import (
	"github.com/leocov-dev/packwiz-nxt/core"
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
	Author                 User                        `gorm:"foreignKey:CreatedBy" json:"author"`
	UpdatedBy              uint                        `json:"updatedBy"`
	IsPublic               bool                        `json:"isPublic"`
	Status                 types.PackStatus            `gorm:"default:draft" json:"status"`
	MCVersion              string                      `json:"mcVersion"`
	Loader                 string                      `json:"loader"`
	LoaderVersion          string                      `json:"loaderVersion"`
	AcceptableGameVersions datatypes.JSONSlice[string] `json:"acceptableGameVersions"`
	Version                string                      `json:"version"`
	PackFormat             string                      `json:"packFormat"`

	Mods []Mod `gorm:"foreignKey:PackSlug" json:"mods"`
}

func (p Pack) AsMeta() core.Pack {
	mods := map[string]*core.Mod{}

	for _, mod := range p.Mods {
		mods[mod.Slug] = mod.AsMeta()
	}

	return core.Pack{
		Name:        p.Name,
		Author:      p.Author.Username,
		Version:     p.Version,
		Description: p.Description,
		PackFormat:  p.PackFormat,
		Versions: map[string]string{
			"minecraft": p.MCVersion,
			p.Loader:    p.LoaderVersion,
		},
		Export: nil,
		Options: map[string]interface{}{
			"acceptable-game-versions": []string(p.AcceptableGameVersions),
		},
		Mods: mods,
	}
}
