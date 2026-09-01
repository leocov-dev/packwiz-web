package tables

import "time"

// ModMigrationResult is the per-mod outcome of a MigrateModsArgs job,
// written once by PackwizService.ResolveMigratedMods and read back by the
// migrate job status endpoint. Mirrors dto.MigrateDryRunMod's shape.
type ModMigrationResult struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	JobId           int64     `json:"jobId"`
	PackID          uint      `json:"packId"`
	ModId           uint      `json:"modId"`
	Slug            string    `json:"slug"`
	Name            string    `json:"name"`
	Pinned          bool      `json:"pinned"`
	UpdateAvailable bool      `json:"updateAvailable"`
	UpdateString    string    `json:"updateString"`
	Incompatible    bool      `json:"incompatible"`
	Error           string    `json:"error"`
	CreatedAt       time.Time `json:"createdAt"`
}

func (ModMigrationResult) TableName() string {
	return "mod_migration_results"
}
