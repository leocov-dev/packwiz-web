package tables

import (
	"time"

	"gorm.io/datatypes"
)

// Bulk add run statuses stored in PackBulkAddRun.Status.
const (
	BulkAddQueued  = "queued"
	BulkAddRunning = "running"
	BulkAddDone    = "done"
	BulkAddFailed  = "failed"
)

// Bulk add item statuses stored in PackBulkAddItem.Status.
const (
	// BulkItemQueued is a line not yet processed by the job.
	BulkItemQueued = "queued"
	// BulkItemAdded is a line that matched exactly and was added by the job.
	BulkItemAdded = "added"
	// BulkItemPending is a line without an exact match, waiting for review.
	BulkItemPending = "pending"
	// BulkItemIncompatibleVersion is a line whose match has no version for the
	// pack's Minecraft version and loader, waiting for review.
	BulkItemIncompatibleVersion = "incompatible_version"
	// BulkItemConflict is a line whose match is incompatible with a mod accepted
	// before it (earlier line or installed), waiting for review.
	BulkItemConflict = "conflict"
	// BulkItemSkipped is a line whose match is already in the pack or was
	// added by an earlier line.
	BulkItemSkipped = "skipped"
	// BulkItemFailed is a line that could not be processed (lookup error).
	BulkItemFailed = "failed"
	// BulkItemPicked is a reviewed line where the user picked a mod to add.
	BulkItemPicked = "picked"
	// BulkItemRejected is a reviewed line the user dismissed.
	BulkItemRejected = "rejected"
)

// PackBulkAddRun is one "add mods from a list" request of a pack.
type PackBulkAddRun struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	PackID     uint       `json:"packId"`
	JobID      int64      `json:"jobId"`
	Status     string     `json:"status"`
	CreatedBy  uint       `json:"createdBy"`
	Error      string     `json:"error"`
	CreatedAt  time.Time  `json:"createdAt"`
	StartedAt  *time.Time `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt"`
}

func (PackBulkAddRun) TableName() string {
	return "pack_bulk_add_runs"
}

// BulkAddCandidate is a search hit offered for review of a line.
type BulkAddCandidate struct {
	Source    string `json:"source"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	ProjectId string `json:"projectId"`
	IconUrl   string `json:"iconUrl,omitempty"`
}

// PackBulkAddItem is one line of a bulk add run. Position is the 1-based list
// order, which is also the priority: earlier lines win conflicts.
type PackBulkAddItem struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	RunID    uint   `json:"runId"`
	Position int    `json:"position"`
	Line     string `json:"line"`
	Status   string `json:"status"`
	// Source, Slug, Name and ProjectID describe the matched (or picked) mod.
	Source    string `json:"source"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	ProjectID string `json:"projectId"`
	Reason    string `json:"reason"`
	// ConflictsWithPosition is the line that won a conflict, 0 for an
	// installed mod or no conflict.
	ConflictsWithPosition int                                   `json:"conflictsWithPosition"`
	Candidates            datatypes.JSONSlice[BulkAddCandidate] `json:"candidates"`
	UpdatedAt             time.Time                             `json:"updatedAt"`
}

func (PackBulkAddItem) TableName() string {
	return "pack_bulk_add_items"
}
