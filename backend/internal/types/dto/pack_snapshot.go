package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// SnapshotFieldChange is one changed field, with display-ready values.
type SnapshotFieldChange struct {
	Field string `json:"field"`
	From  any    `json:"from"`
	To    any    `json:"to"`
}

// SnapshotModRef identifies a mod inside a diff.
type SnapshotModRef struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	FileName string `json:"fileName"`
}

// SnapshotModChange is a mod present on both sides with differing fields.
type SnapshotModChange struct {
	SnapshotModRef
	Changes []SnapshotFieldChange `json:"changes"`
}

// SnapshotDiff is the difference between two pack states.
type SnapshotDiff struct {
	Pack    []SnapshotFieldChange `json:"pack"`
	Added   []SnapshotModRef      `json:"added"`
	Removed []SnapshotModRef      `json:"removed"`
	Changed []SnapshotModChange   `json:"changed"`
}

// SnapshotSummary is the stored per-snapshot change count against its parent.
type SnapshotSummary struct {
	Added      int      `json:"added"`
	Removed    int      `json:"removed"`
	Changed    int      `json:"changed"`
	PackFields []string `json:"packFields"`
}

// PackSnapshotItem is one row of the snapshot list.
type PackSnapshotItem struct {
	ID                  uint            `json:"id"`
	Seq                 int             `json:"seq"`
	ParentID            *uint           `json:"parentId"`
	Reason              string          `json:"reason"`
	Detail              map[string]any  `json:"detail"`
	Summary             SnapshotSummary `json:"summary"`
	CreatedAt           time.Time       `json:"createdAt"`
	CreatedBy           uint            `json:"createdBy"`
	CreatedByUsername   string          `json:"createdByUsername"`
	AbandonedAt         *time.Time      `json:"abandonedAt"`
	AbandonedByRevertTo *uint           `json:"abandonedByRevertTo"`
	IsHead              bool            `json:"isHead"`
}

// PackSnapshotListQuery filters the snapshot list.
type PackSnapshotListQuery struct {
	Abandoned bool `form:"abandoned"`
	Page      int  `form:"page" validate:"gte=1"`
	PageSize  int  `form:"pageSize" validate:"gte=1,lte=100"`
}

func (q *PackSnapshotListQuery) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(q)
}

// PackSnapshotListResponse is a page of snapshots.
type PackSnapshotListResponse struct {
	Snapshots  []PackSnapshotItem `json:"snapshots"`
	Total      int64              `json:"total"`
	HeadID     *uint              `json:"headId"`
	PackStatus string             `json:"packStatus"`
}

// Snapshot diff bases.
const (
	SnapshotAgainstParent  = "parent"
	SnapshotAgainstCurrent = "current"
)

// PackSnapshotDetailQuery picks what a snapshot is diffed against.
type PackSnapshotDetailQuery struct {
	Against string `form:"against" validate:"omitempty,oneof=parent current"`
}

func (q *PackSnapshotDetailQuery) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(q)
}

// PackSnapshotDetailResponse is one snapshot with its diff.
type PackSnapshotDetailResponse struct {
	Snapshot PackSnapshotItem `json:"snapshot"`
	Against  string           `json:"against"`
	Diff     SnapshotDiff     `json:"diff"`
}

// CloneSnapshotRequest names the new pack created from a snapshot.
type CloneSnapshotRequest struct {
	Slug string `json:"slug" validate:"required,slug"`
	Name string `json:"name" validate:"required"`
}

func (r CloneSnapshotRequest) Validate() error {
	validate, err := newSlugValidator()
	if err != nil {
		return err
	}
	return validate.Struct(r)
}

// RevertSnapshotResponse reports the outcome of a revert.
type RevertSnapshotResponse struct {
	Changed bool `json:"changed"`
	HeadID  uint `json:"headId"`
}

// PruneSnapshotsResponse reports how many abandoned snapshots were deleted.
type PruneSnapshotsResponse struct {
	Deleted int64 `json:"deleted"`
}

// RebaseSnapshotResponse reports how many snapshots before the new root were deleted.
type RebaseSnapshotResponse struct {
	Deleted int64 `json:"deleted"`
}
