package tables

import "time"

// ModUpdateCheck is the latest update-check outcome for one mod, written by
// the CheckUpdatesArgs job. At most one row exists per (pack, mod).
type ModUpdateCheck struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	PackID          uint      `json:"packId"`
	ModID           uint      `json:"modId"`
	UpdateAvailable bool      `json:"updateAvailable"`
	UpdateString    string    `json:"updateString"`
	Error           string    `json:"error"`
	CheckedAt       time.Time `json:"checkedAt"`
}

func (ModUpdateCheck) TableName() string {
	return "mod_update_checks"
}

// Update check run statuses stored in PackUpdateCheckRun.Status.
const (
	UpdateCheckQueued  = "queued"
	UpdateCheckRunning = "running"
	UpdateCheckDone    = "done"
	UpdateCheckFailed  = "failed"
)

// PackUpdateCheckRun is the state of the latest update-check job of a pack.
type PackUpdateCheckRun struct {
	PackID     uint       `gorm:"primaryKey" json:"packId"`
	JobID      int64      `json:"jobId"`
	Status     string     `json:"status"`
	StartedAt  time.Time  `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt"`
	Error      string     `json:"error"`
}

func (PackUpdateCheckRun) TableName() string {
	return "pack_update_check_runs"
}
