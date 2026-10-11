package tables

import (
	"time"

	"github.com/leocov-dev/packwiz-nxt/jarmeta"
	"github.com/lib/pq"
	"gorm.io/datatypes"
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
	// Loaders are the loader names found in Meta, for prune filtering.
	Loaders    pq.StringArray `gorm:"type:text[]"`
	Error      string
	ParsedAt   time.Time
	LastUsedAt time.Time
}

func (ModJarMeta) TableName() string {
	return "mod_jar_meta"
}
