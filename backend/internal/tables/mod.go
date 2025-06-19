package tables

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/leocov-dev/packwiz-nxt/core"
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

type UpdateInfo map[string]interface{}

func (u UpdateInfo) Value() (driver.Value, error) {
	return json.Marshal(u)
}

func (u *UpdateInfo) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed in UpdateInfo Scan")
	}
	return json.Unmarshal(bytes, u)
}

type Mod struct {
	ID         uint         `gorm:"primaryKey" json:"id"`
	Slug       string       `json:"slug"`
	PackID     uint         `json:"packId"`
	Name       string       `json:"name"`
	FileName   string       `json:"fileName"`
	Side       core.ModSide `json:"side"`
	Pinned     bool         `json:"pinned"`
	Download   DownloadInfo `gorm:"type:json"  json:"download"`
	HashFormat string       `gorm:"default:sha256" json:"hashFormat"`
	Alias      string       `json:"alias"`
	Type       string       `gorm:"default:mods" json:"type"`
	Source     string       `json:"source"`
	Update     UpdateInfo   `gorm:"type:json"  json:"update"`
	Preserve   bool         `gorm:"default:false" json:"preserve"`
	CreatedAt  time.Time    `json:"createdAt"`
	UpdatedAt  time.Time    `json:"updatedAt"`
	CreatedBy  uint         `json:"createdBy"`
	UpdatedBy  uint         `json:"updatedBy"`
}

func (m Mod) AsMeta() *core.Mod {
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
		Update:     core.ModUpdate{m.Source: core.ModSourceData(m.Update)},
		Option:     nil,
		Slug:       m.Slug,
		ModType:    m.Type,
		HashFormat: m.HashFormat,
		Alias:      m.Alias,
		Preserve:   m.Preserve,
	}
}

func ExtractModSource(mod *core.Mod) (string, map[string]interface{}) {
	if data, ok := mod.Update["modrinth"]; ok {
		return "modrinth", data
	} else if data, ok := mod.Update["curseforge"]; ok {
		return "curseforge", data
	} else if data, ok := mod.Update["github"]; ok {
		return "github", data
	} else {
		return "", map[string]interface{}{}
	}
}
