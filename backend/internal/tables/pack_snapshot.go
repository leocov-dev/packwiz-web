package tables

import (
	"gorm.io/datatypes"
	"time"
)

// SnapshotReason is why a pack snapshot was taken.
type SnapshotReason string

const (
	SnapshotBaseline    SnapshotReason = "baseline"
	SnapshotPublish     SnapshotReason = "publish"
	SnapshotPackEdit    SnapshotReason = "pack_edit"
	SnapshotModAdd      SnapshotReason = "mod_add"
	SnapshotModRemove   SnapshotReason = "mod_remove"
	SnapshotModUpdate   SnapshotReason = "mod_update"
	SnapshotModSide     SnapshotReason = "mod_side"
	SnapshotModOption   SnapshotReason = "mod_option"
	SnapshotModPin      SnapshotReason = "mod_pin"
	SnapshotRehash      SnapshotReason = "rehash"
	SnapshotUpdateAll   SnapshotReason = "update_all"
	SnapshotMigrate     SnapshotReason = "migrate"
	SnapshotMigrateMods SnapshotReason = "migrate_mods"
)

// PackSnapshot is a full copy of a pack's content at a point in time.
// Written only by packwiz_svc (recordPackHistory); rows are never deleted,
// a revert marks the snapshots after its target as abandoned.
type PackSnapshot struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	PackID              uint           `json:"packId"`
	Seq                 int            `json:"seq"`
	ParentID            *uint          `json:"parentId"`
	Reason              SnapshotReason `json:"reason"`
	Detail              datatypes.JSON `json:"detail"`
	Summary             datatypes.JSON `json:"summary"`
	CreatedBy           uint           `json:"createdBy"`
	CreatedAt           time.Time      `json:"createdAt"`
	AbandonedAt         *time.Time     `json:"abandonedAt"`
	AbandonedBy         *uint          `json:"abandonedBy"`
	AbandonedByRevertTo *uint          `json:"abandonedByRevertTo"`
	SchemaVersion       int            `json:"schemaVersion"`
	Payload             datatypes.JSON `json:"-"`
	PayloadHash         string         `json:"payloadHash"`
}

func (PackSnapshot) TableName() string {
	return "pack_snapshots"
}
