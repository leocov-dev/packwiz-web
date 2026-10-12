package tables

import (
	"time"

	"github.com/leocov-dev/packwiz-nxt/jarmeta"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ModJarMeta is the cached, parsed loader metadata of one mod jar, keyed by
// the jar hash and shared across packs. Jars themselves are never retained.
// A parse failure is cached too: Meta is nil and Error is set.
type ModJarMeta struct {
	HashFormat    string `gorm:"primaryKey"`
	Hash          string `gorm:"primaryKey"`
	ParserVersion int
	// Meta is nil when the jar could not be parsed.
	Meta *datatypes.JSONSlice[jarmeta.LoaderMeta] `gorm:"type:jsonb"`
	// Loaders are the loader names found in Meta, for prune filtering. A nil
	// slice is stored as '{}' (see BeforeSave), never NULL.
	Loaders pq.StringArray `gorm:"type:text[]"`
	Error   string
	// ParsedAt is set to the current time on create when left zero
	// (autoCreateTime), so a zero value never reaches the column.
	ParsedAt time.Time `gorm:"autoCreateTime"`
	// LastUsedAt is set to the current time on create when left zero
	// (autoCreateTime). Later touches are explicit UPDATEs by the scanner;
	// GORM does not maintain it on update.
	LastUsedAt time.Time `gorm:"autoCreateTime"`
}

// BeforeSave turns a nil Loaders into an empty slice. A nil pq.StringArray is
// written as SQL NULL, which violates the NOT NULL column (parse-error rows
// have no loaders).
func (m *ModJarMeta) BeforeSave(*gorm.DB) error {
	if m.Loaders == nil {
		m.Loaders = pq.StringArray{}
	}
	return nil
}

func (ModJarMeta) TableName() string {
	return "mod_jar_meta"
}
