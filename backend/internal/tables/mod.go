package tables

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/leocov-dev/packwiz-nxt/core"
	"gorm.io/datatypes"
	"math"
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

type OptionInfo struct {
	Optional    bool   `json:"optional"`
	Description string `json:"description"`
	Default     bool   `json:"default"`
}

func (o OptionInfo) Value() (driver.Value, error) {
	return json.Marshal(o)
}

func (o *OptionInfo) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed in OptionInfo Scan")
	}
	return json.Unmarshal(bytes, o)
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
	// Version is the human-readable installed version. DB-only: it is never
	// written to the generated pack files (see AsMeta). NULL for legacy rows.
	Version       string                    `json:"version,omitempty"`
	Type          string                    `gorm:"default:mods" json:"type"`
	Source        string                    `json:"source"`
	Update        UpdateInfo                `gorm:"type:json"  json:"update"`
	Option        OptionInfo                `gorm:"type:json"  json:"option"`
	Preserve      bool                      `gorm:"default:false" json:"preserve"`
	CreatedAt     time.Time                 `json:"createdAt"`
	UpdatedAt     time.Time                 `json:"updatedAt"`
	CreatedBy     uint                      `json:"createdBy"`
	UpdatedBy     uint                      `json:"updatedBy"`
	IsDependency  bool                      `gorm:"default:false" json:"isDependency"`
	DependencyIds datatypes.JSONSlice[uint] `json:"dependencyIds"`
}

func (m Mod) AsMeta() *core.Mod {
	var option *core.ModOption
	if m.Option.Optional || m.Option.Description != "" || m.Option.Default {
		option = &core.ModOption{
			Optional:    m.Option.Optional,
			Description: m.Option.Description,
			Default:     m.Option.Default,
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
		Update:     core.ModUpdate{m.Source: core.ModSourceData(integralNumbersAsInts(m.Update))},
		Option:     option,
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

// maxSafeInteger is the largest integer a float64 (and so a JSON number read
// into interface{}) represents exactly.
const maxSafeInteger = 1 << 53

// integralNumbersAsInts returns update data with whole-number float64 values
// converted to int64, recursively. Update data is stored as JSON, so numbers
// such as CurseForge project and file ids come back from the database as
// float64, which TOML writes as "6928979.0". packwiz-installer decodes those
// ids as integers and rejects the float form, so they must be written as
// integers. The input is not modified.
func integralNumbersAsInts(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = integralValueAsInt(v)
	}
	return out
}

func integralValueAsInt(v interface{}) interface{} {
	switch t := v.(type) {
	case float64:
		if t == math.Trunc(t) && math.Abs(t) <= maxSafeInteger {
			return int64(t)
		}
		return t
	case map[string]interface{}:
		return integralNumbersAsInts(t)
	case UpdateInfo:
		return integralNumbersAsInts(t)
	case []interface{}:
		out := make([]interface{}, len(t))
		for i, e := range t {
			out[i] = integralValueAsInt(e)
		}
		return out
	default:
		return v
	}
}
