package tables

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/leocov-dev/packwiz-nxt/core"
	"strconv"
	"time"
)

type DownloadInfo struct {
	URL        string `json:"url"`
	Mode       string `json:"mode"`
	Hash       string `json:"hash"`
	HashFormat string `json:"hashFormat"`
}

func (d DownloadInfo) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DownloadInfo) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed in DownloadInfo Scan")
	}
	return json.Unmarshal(bytes, d)
}

type Mod struct {
	Id         uint         `gorm:"primarykey" json:"id"`
	Slug       string       `gorm:"uniqueIndex:idx_pack_mod_slug,priority:2" json:"slug"`
	PackSlug   string       `gorm:"uniqueIndex:idx_pack_mod_slug,priority:1" json:"packSlug"`
	Name       string       `json:"name"`
	FileName   string       `json:"fileName"`
	Side       core.ModSide `json:"side"`
	Pinned     bool         `json:"pinned"`
	Download   DownloadInfo `gorm:"type:json"  json:"download"`
	HashFormat string       `gorm:"default:sha256" json:"hashFormat"`
	Alias      string       `json:"alias"`
	Type       string       `gorm:"default:mods" json:"type"`
	Source     string       `json:"source"`
	ModKey     string       `json:"modKey"`
	VersionKey string       `json:"versionKey"`
	Preserve   bool         `gorm:"default:false" json:"preserve"`
	CreatedAt  time.Time    `json:"createdAt"`
	UpdatedAt  time.Time    `json:"updatedAt"`
	CreatedBy  uint         `json:"createdBy"`
	UpdatedBy  uint         `json:"updatedBy"`
}

func (m Mod) AsMeta() *core.Mod {
	update := core.ModUpdate{}
	if m.Source == "modrinth" {
		update["modrinth"] = map[string]interface{}{
			"modId":   m.ModKey,
			"version": m.VersionKey,
		}
	} else if m.Source == "curseforge" {
		projectId, _ := strconv.Atoi(m.ModKey)
		fileId, _ := strconv.Atoi(m.VersionKey)
		update["curseforge"] = map[string]interface{}{
			"projectId": projectId,
			"fileId":    fileId,
		}
	}

	return &core.Mod{
		Name:     m.Name,
		FileName: m.FileName,
		Side:     m.Side,
		Pin:      m.Pinned,
		Download: core.ModDownload{
			URL:        m.Download.URL,
			HashFormat: m.Download.HashFormat,
			Hash:       m.Download.Hash,
			Mode:       m.Download.Mode,
		},
		Update:     update,
		Option:     nil,
		Slug:       m.Slug,
		ModType:    m.Type,
		HashFormat: m.HashFormat,
		Alias:      m.Alias,
		Preserve:   m.Preserve,
	}
}
