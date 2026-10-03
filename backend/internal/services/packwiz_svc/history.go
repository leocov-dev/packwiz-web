package packwiz_svc

import (
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"packwiz-web/internal/tables"
	"packwiz-web/internal/types"
)

// Pack history
//
// A pack that is not a draft keeps a snapshot of its full content after every
// change. Live snapshots form one chain ending at packs.head_snapshot_id; a
// revert abandons the snapshots after its target (see RevertToSnapshot).
//
// Every content change runs inside withPackHistory, which takes the pack row
// lock first (so lock order is always pack, then mods) and rolls the change
// back if its snapshot cannot be written: history is not optional.

// lockPackRow locks the pack row for the rest of tx. Archived packs are
// included; callers decide what that means.
func lockPackRow(tx *gorm.DB, packId uint) (tables.Pack, error) {
	var pack tables.Pack
	if err := tx.Unscoped().
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&pack, packId).Error; err != nil {
		return tables.Pack{}, fmt.Errorf("lock pack %d: %w", packId, err)
	}
	return pack, nil
}

// recordsHistory reports whether pack currently keeps history.
func recordsHistory(pack tables.Pack) bool {
	return pack.Status != types.PackStatusDraft && !pack.DeletedAt.Valid
}

// loadHistoryPayload builds the payload of the pack's live rows.
func loadHistoryPayload(tx *gorm.DB, pack tables.Pack) (historyPayload, error) {
	var mods []tables.Mod
	if err := tx.Where("pack_id = ?", pack.ID).Find(&mods).Error; err != nil {
		return historyPayload{}, fmt.Errorf("load mods of pack %d: %w", pack.ID, err)
	}
	return buildHistoryPayload(pack, mods), nil
}

// loadHeadSnapshot returns the head snapshot with its decoded payload, or nil
// when the pack has no history yet.
func loadHeadSnapshot(tx *gorm.DB, pack tables.Pack) (*tables.PackSnapshot, historyPayload, error) {
	if pack.HeadSnapshotID == nil {
		return nil, emptyHistoryPayload(), nil
	}
	var head tables.PackSnapshot
	if err := tx.Where("id = ? AND pack_id = ?", *pack.HeadSnapshotID, pack.ID).First(&head).Error; err != nil {
		return nil, historyPayload{}, fmt.Errorf("load head snapshot of pack %d: %w", pack.ID, err)
	}
	payload, err := decodeHistoryPayload(head.SchemaVersion, head.Payload)
	if err != nil {
		return nil, historyPayload{}, err
	}
	return &head, payload, nil
}

// recordPackHistory snapshots the pack's current content if it keeps history
// and the content differs from the head. It locks the pack row itself, so
// callers outside withPackHistory (update-all, the migrate job) can use it.
func recordPackHistory(tx *gorm.DB, packId, userID uint, reason tables.SnapshotReason, detail map[string]any) error {
	pack, err := lockPackRow(tx, packId)
	if err != nil {
		return err
	}
	if !recordsHistory(pack) {
		return nil
	}

	current, err := loadHistoryPayload(tx, pack)
	if err != nil {
		return err
	}
	raw, hash, err := encodeHistoryPayload(current)
	if err != nil {
		return err
	}

	head, headPayload, err := loadHeadSnapshot(tx, pack)
	if err != nil {
		return err
	}
	if head != nil && head.PayloadHash == hash {
		return nil
	}

	var maxSeq int
	if err := tx.Model(&tables.PackSnapshot{}).
		Where("pack_id = ?", packId).
		Select("COALESCE(MAX(seq), 0)").
		Scan(&maxSeq).Error; err != nil {
		return fmt.Errorf("next snapshot seq of pack %d: %w", packId, err)
	}

	if detail == nil {
		detail = map[string]any{}
	}
	detailJSON, err := json.Marshal(detail)
	if err != nil {
		return fmt.Errorf("encode snapshot detail: %w", err)
	}
	summaryJSON, err := json.Marshal(summarizeHistoryDiff(diffHistory(headPayload, current)))
	if err != nil {
		return fmt.Errorf("encode snapshot summary: %w", err)
	}

	snapshot := tables.PackSnapshot{
		PackID:        packId,
		Seq:           maxSeq + 1,
		Reason:        reason,
		Detail:        datatypes.JSON(detailJSON),
		Summary:       datatypes.JSON(summaryJSON),
		CreatedBy:     userID,
		SchemaVersion: historySchemaVersion,
		Payload:       datatypes.JSON(raw),
		PayloadHash:   hash,
	}
	if head != nil {
		snapshot.ParentID = &head.ID
	}
	if err := tx.Create(&snapshot).Error; err != nil {
		return fmt.Errorf("create snapshot of pack %d: %w", packId, err)
	}

	return setPackHead(tx, packId, snapshot.ID)
}

// setPackHead moves the head pointer. head_snapshot_id is read-only to gorm
// (see tables.Pack), so this is the only writer.
func setPackHead(tx *gorm.DB, packId, snapshotId uint) error {
	if err := tx.Exec("UPDATE packs SET head_snapshot_id = ? WHERE id = ?", snapshotId, packId).Error; err != nil {
		return fmt.Errorf("set head snapshot of pack %d: %w", packId, err)
	}
	return nil
}

// ensureHistoryBaseline snapshots a published pack that has no history yet, so
// the state before its first tracked change is not lost. Call it before the
// change. A no-op for drafts, archived packs and packs that already have a head.
func ensureHistoryBaseline(tx *gorm.DB, packId, userID uint) error {
	pack, err := lockPackRow(tx, packId)
	if err != nil {
		return err
	}
	if !recordsHistory(pack) || pack.HeadSnapshotID != nil {
		return nil
	}
	return recordPackHistory(tx, packId, userID, tables.SnapshotBaseline, nil)
}

// withPackHistory runs fn in a transaction that locks the pack, makes sure a
// baseline exists, applies the change, then records the snapshot. If any step
// fails the whole change is rolled back.
//
// detail is stored on the snapshot as it is after fn returns, so fn may add
// to it.
func (ps *PackwizService) withPackHistory(
	packId, userID uint,
	reason tables.SnapshotReason,
	detail map[string]any,
	fn func(tx *gorm.DB) error,
) error {
	return ps.db.Transaction(func(tx *gorm.DB) error {
		if _, err := lockPackRow(tx, packId); err != nil {
			return err
		}
		if err := ensureHistoryBaseline(tx, packId, userID); err != nil {
			return err
		}
		if err := fn(tx); err != nil {
			return err
		}
		return recordPackHistory(tx, packId, userID, reason, detail)
	})
}

// migrateJobActive reports whether a migrate_mods job for the pack is queued
// or running. River does not expose a per-args lookup, so this reads river_job.
func (ps *PackwizService) migrateJobActive(packId uint) (bool, error) {
	var active bool
	if err := ps.db.Raw(
		`SELECT EXISTS (
			SELECT 1 FROM river_job
			WHERE kind = 'migrate_mods'
			  AND (args->>'packId')::bigint = ?
			  AND state IN ('available', 'pending', 'scheduled', 'retryable', 'running')
		)`, packId,
	).Scan(&active).Error; err != nil {
		return false, fmt.Errorf("check active migrate job of pack %d: %w", packId, err)
	}
	return active, nil
}
