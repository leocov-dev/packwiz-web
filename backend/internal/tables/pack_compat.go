package tables

import (
	"time"

	"gorm.io/datatypes"
)

// Compat run statuses stored in PackCompatRun.Status.
const (
	CompatQueued  = "queued"
	CompatRunning = "running"
	CompatDone    = "done"
	CompatFailed  = "failed"
)

// PackCompatRun is the state of the latest compatibility scan of a pack.
type PackCompatRun struct {
	PackID uint   `gorm:"primaryKey" json:"packId"`
	JobID  int64  `json:"jobId"`
	Status string `json:"status"`
	// TargetHash identifies the pack target and mods the scan ran against.
	TargetHash     string `json:"targetHash"`
	ErrorCount     int    `json:"errorCount"`
	WarningCount   int    `json:"warningCount"`
	InfoCount      int    `json:"infoCount"`
	UnscannedCount int    `json:"unscannedCount"`
	// StartedAt defaults to the current time on create (autoCreateTime).
	StartedAt  time.Time  `gorm:"autoCreateTime" json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt"`
	Error      string     `json:"error"`
}

func (PackCompatRun) TableName() string {
	return "pack_compat_runs"
}

// PackCompatFinding is one finding of the latest completed scan of a pack.
type PackCompatFinding struct {
	ID       uint              `gorm:"primaryKey" json:"id"`
	PackID   uint              `json:"packId"`
	ModID    *uint             `json:"modId"`
	Slug     string            `json:"slug"`
	Severity string            `json:"severity"`
	Code     string            `json:"code"`
	Message  string            `json:"message"`
	Detail   datatypes.JSONMap `gorm:"type:jsonb" json:"detail"`
}

func (PackCompatFinding) TableName() string {
	return "pack_compat_findings"
}
