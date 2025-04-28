package tables

import (
	"packwiz-web/internal/types"
	"time"
)

type Mod struct {
	Id       uint          `gorm:"primarykey" json:"id"`
	PackSlug string        `gorm:"uniqueIndex:idx_pack_mod_slug,priority:1" json:"packSlug"`
	ModSlug  string        `gorm:"uniqueIndex:idx_pack_mod_slug,priority:2" json:"modSlug"`
	Name     string        `json:"name"`
	Type     string        `json:"type"`
	FileName string        `json:"fileName"`
	Side     types.ModSide `json:"side"`
	Pinned   bool          `json:"pinned"`

	DownloadUrl        string `json:"downloadUrl"`
	DownloadMode       string `json:"downloadMode"`
	DownloadHash       string `json:"downloadHash"`
	DownloadHashFormat string `json:"downloadHashFormat"`

	Source     string `json:"source"`
	ModKey     string `json:"modKey"`
	VersionKey string `json:"versionKey"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedBy uint      `json:"createdBy"`
	UpdatedBy uint      `json:"updatedBy"`

	Hash       string `json:"hash"` // index.toml hash
	HashFormat string `json:"hashFormat"`
	Metafile   bool   `json:"metafile"`

	SourceLink string `gorm:"-" json:"sourceLink"`
}
