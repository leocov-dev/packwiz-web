package packwiz_svc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"packwiz-web/internal/tables"
	"packwiz-web/internal/types"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

var (
	errSnapshotNotFound  = errors.New("snapshot not found")
	errSnapshotAbandoned = errors.New("snapshot was abandoned by a revert and can no longer be restored")
	errPackArchived      = errors.New("pack is archived")
)

// snapshotListRow is a snapshot's metadata plus the author's username. It
// never carries the payload, which can be large.
type snapshotListRow struct {
	ID                  uint
	Seq                 int
	ParentID            *uint
	Reason              string
	Detail              datatypes.JSON
	Summary             datatypes.JSON
	CreatedAt           time.Time
	CreatedBy           uint
	CreatedByUsername   string
	AbandonedAt         *time.Time
	AbandonedByRevertTo *uint
}

const snapshotListColumns = "pack_snapshots.id, pack_snapshots.seq, pack_snapshots.parent_id, " +
	"pack_snapshots.reason, pack_snapshots.detail, pack_snapshots.summary, " +
	"pack_snapshots.created_at, pack_snapshots.created_by, " +
	"COALESCE(users.username, '') AS created_by_username, " +
	"pack_snapshots.abandoned_at, pack_snapshots.abandoned_by_revert_to"

// snapshotQuery selects snapshot metadata. The users join is not filtered by
// soft delete: history keeps the name of a user who has since been removed.
func (ps *PackwizService) snapshotQuery() *gorm.DB {
	return ps.db.Table("pack_snapshots").
		Select(snapshotListColumns).
		Joins("LEFT JOIN users ON users.id = pack_snapshots.created_by")
}

func (row snapshotListRow) item(headID *uint) dto.PackSnapshotItem {
	detail := map[string]any{}
	if len(row.Detail) > 0 {
		// a malformed detail is display-only data; fall back to empty
		_ = json.Unmarshal(row.Detail, &detail)
	}

	summary := dto.SnapshotSummary{PackFields: []string{}}
	if len(row.Summary) > 0 {
		_ = json.Unmarshal(row.Summary, &summary)
	}
	if summary.PackFields == nil {
		summary.PackFields = []string{}
	}

	return dto.PackSnapshotItem{
		ID:                  row.ID,
		Seq:                 row.Seq,
		ParentID:            row.ParentID,
		Reason:              row.Reason,
		Detail:              detail,
		Summary:             summary,
		CreatedAt:           row.CreatedAt,
		CreatedBy:           row.CreatedBy,
		CreatedByUsername:   row.CreatedByUsername,
		AbandonedAt:         row.AbandonedAt,
		AbandonedByRevertTo: row.AbandonedByRevertTo,
		IsHead:              headID != nil && *headID == row.ID,
	}
}

// revertBlockedError carries a guard refusal out of a revert transaction.
type revertBlockedError struct{ response.ServerError }

// snapshotErr maps a failed snapshot lookup or operation to a response.
func snapshotErr(err error) response.ServerError {
	var blocked revertBlockedError
	if errors.As(err, &blocked) {
		return blocked.ServerError
	}
	switch {
	case errors.Is(err, errSnapshotNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		return response.New(http.StatusNotFound, "snapshot not found")
	case errors.Is(err, errSnapshotAbandoned), errors.Is(err, errPackArchived):
		return response.New(http.StatusConflict, err.Error())
	default:
		return response.Wrap(err)
	}
}

// getPackHead reads the pack's status and head pointer, archived packs included.
func (ps *PackwizService) getPackHead(packId uint) (tables.Pack, error) {
	var pack tables.Pack
	if err := ps.db.Unscoped().
		Select("id", "status", "deleted_at", "head_snapshot_id").
		First(&pack, packId).Error; err != nil {
		return tables.Pack{}, err
	}
	return pack, nil
}

// ListSnapshots lists a pack's history, newest first. Abandoned snapshots are
// left out unless q.Abandoned is set.
func (ps *PackwizService) ListSnapshots(packId uint, q dto.PackSnapshotListQuery) (dto.PackSnapshotListResponse, response.ServerError) {
	pack, err := ps.getPackHead(packId)
	if err != nil {
		return dto.PackSnapshotListResponse{}, response.New(http.StatusNotFound, fmt.Sprintf("pack '%d' not found", packId))
	}

	filtered := func(db *gorm.DB) *gorm.DB {
		db = db.Where("pack_snapshots.pack_id = ?", packId)
		if !q.Abandoned {
			db = db.Where("pack_snapshots.abandoned_at IS NULL")
		}
		return db
	}

	var total int64
	if err := filtered(ps.db.Model(&tables.PackSnapshot{})).Count(&total).Error; err != nil {
		return dto.PackSnapshotListResponse{}, response.Wrap(err)
	}

	var rows []snapshotListRow
	if err := filtered(ps.snapshotQuery()).
		Order("pack_snapshots.seq DESC").
		Limit(q.PageSize).
		Offset((q.Page - 1) * q.PageSize).
		Scan(&rows).Error; err != nil {
		return dto.PackSnapshotListResponse{}, response.Wrap(err)
	}

	items := make([]dto.PackSnapshotItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.item(pack.HeadSnapshotID))
	}

	return dto.PackSnapshotListResponse{
		Snapshots:  items,
		Total:      total,
		HeadID:     pack.HeadSnapshotID,
		PackStatus: string(pack.Status),
	}, nil
}

// loadSnapshot reads one snapshot of a pack with its payload. A snapshot of a
// different pack is reported as not found.
func (ps *PackwizService) loadSnapshot(db *gorm.DB, packId, snapshotId uint) (tables.PackSnapshot, error) {
	var snapshot tables.PackSnapshot
	if err := db.Where("id = ? AND pack_id = ?", snapshotId, packId).First(&snapshot).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tables.PackSnapshot{}, errSnapshotNotFound
		}
		return tables.PackSnapshot{}, err
	}
	return snapshot, nil
}

// GetSnapshot returns one snapshot and a diff.
//
// against "parent" (the default) shows what the snapshot changed relative to
// the snapshot before it. against "current" shows what has changed in the pack
// since the snapshot, i.e. what a revert to it would undo.
func (ps *PackwizService) GetSnapshot(packId, snapshotId uint, against string) (dto.PackSnapshotDetailResponse, response.ServerError) {
	if against == "" {
		against = dto.SnapshotAgainstParent
	}

	pack, err := ps.getPackHead(packId)
	if err != nil {
		return dto.PackSnapshotDetailResponse{}, response.New(http.StatusNotFound, fmt.Sprintf("pack '%d' not found", packId))
	}

	snapshot, err := ps.loadSnapshot(ps.db, packId, snapshotId)
	if err != nil {
		return dto.PackSnapshotDetailResponse{}, snapshotErr(err)
	}
	payload, err := decodeHistoryPayload(snapshot.SchemaVersion, snapshot.Payload)
	if err != nil {
		return dto.PackSnapshotDetailResponse{}, response.Wrap(err)
	}

	var row snapshotListRow
	if err := ps.snapshotQuery().Where("pack_snapshots.id = ?", snapshotId).Scan(&row).Error; err != nil {
		return dto.PackSnapshotDetailResponse{}, response.Wrap(err)
	}

	var diff dto.SnapshotDiff
	switch against {
	case dto.SnapshotAgainstCurrent:
		var fullPack tables.Pack
		if err := ps.db.Unscoped().First(&fullPack, packId).Error; err != nil {
			return dto.PackSnapshotDetailResponse{}, response.Wrap(err)
		}
		current, err := loadHistoryPayload(ps.db, fullPack)
		if err != nil {
			return dto.PackSnapshotDetailResponse{}, response.Wrap(err)
		}
		diff = diffHistory(payload, current)
	default:
		parent := emptyHistoryPayload()
		if snapshot.ParentID != nil {
			parentSnapshot, err := ps.loadSnapshot(ps.db, packId, *snapshot.ParentID)
			if err != nil {
				return dto.PackSnapshotDetailResponse{}, snapshotErr(err)
			}
			parent, err = decodeHistoryPayload(parentSnapshot.SchemaVersion, parentSnapshot.Payload)
			if err != nil {
				return dto.PackSnapshotDetailResponse{}, response.Wrap(err)
			}
		}
		diff = diffHistory(parent, payload)
		against = dto.SnapshotAgainstParent
	}

	return dto.PackSnapshotDetailResponse{
		Snapshot: row.item(pack.HeadSnapshotID),
		Against:  against,
		Diff:     diff,
	}, nil
}

// historyModColumns are the mods columns a snapshot owns, for an UPDATE. The
// version is written as NULL when empty, the convention applyModUpdate uses.
// Dependency ids are not here: they need every mod's id, see syncHistoryDependencies.
func historyModColumns(m historyMod, userID uint) map[string]any {
	var version any
	if m.Version != "" {
		version = m.Version
	}
	return map[string]any{
		"name":          m.Name,
		"file_name":     m.FileName,
		"side":          m.Side,
		"pinned":        m.Pinned,
		"hash_format":   m.HashFormat,
		"alias":         m.Alias,
		"type":          m.Type,
		"source":        m.Source,
		"preserve":      m.Preserve,
		"download":      m.Download,
		"update":        m.Update,
		"option":        m.Option,
		"version":       version,
		"is_dependency": m.IsDependency,
		"updated_by":    userID,
	}
}

// insertHistoryMods creates a row for each mod. Dependency ids are filled in
// afterwards by syncHistoryDependencies.
func insertHistoryMods(tx *gorm.DB, packId, userID uint, mods []historyMod) error {
	for _, m := range mods {
		row := tables.Mod{
			Slug:         m.Slug,
			PackID:       packId,
			Name:         m.Name,
			FileName:     m.FileName,
			Side:         m.Side,
			Pinned:       m.Pinned,
			Download:     m.Download,
			HashFormat:   m.HashFormat,
			Alias:        m.Alias,
			Version:      m.Version,
			Type:         m.Type,
			Source:       m.Source,
			Update:       m.Update,
			Option:       m.Option,
			Preserve:     m.Preserve,
			CreatedBy:    userID,
			UpdatedBy:    userID,
			IsDependency: m.IsDependency,
		}
		if err := tx.Create(&row).Error; err != nil {
			return fmt.Errorf("restore mod %s: %w", m.Slug, err)
		}
	}
	return nil
}

// syncHistoryDependencies points every mod's dependency_ids at the rows for the
// target's dependency slugs. Only rows whose valid ids differ are written, so
// untouched mods keep their updated_at.
func syncHistoryDependencies(tx *gorm.DB, packId, userID uint, target []historyMod) error {
	var rows []tables.Mod
	if err := tx.Where("pack_id = ?", packId).Find(&rows).Error; err != nil {
		return fmt.Errorf("load mods of pack %d: %w", packId, err)
	}

	slugToID := make(map[string]uint, len(rows))
	valid := make(map[uint]struct{}, len(rows))
	for _, r := range rows {
		slugToID[r.Slug] = r.ID
		valid[r.ID] = struct{}{}
	}
	byID := make(map[uint]tables.Mod, len(rows))
	for _, r := range rows {
		byID[r.ID] = r
	}

	for _, m := range target {
		id, ok := slugToID[m.Slug]
		if !ok {
			continue
		}
		want := resolveDependencyIds(slugToID, m.Dependencies)

		have := make([]uint, 0, len(byID[id].DependencyIds))
		for _, dep := range byID[id].DependencyIds {
			if _, ok := valid[dep]; ok {
				have = append(have, dep)
			}
		}

		slices.Sort(want)
		slices.Sort(have)
		if slices.Equal(want, have) {
			continue
		}

		if err := tx.Model(&tables.Mod{}).Where("id = ?", id).Updates(map[string]any{
			"dependency_ids": datatypes.JSONSlice[uint](want),
			"updated_by":     userID,
		}).Error; err != nil {
			return fmt.Errorf("restore dependencies of mod %s: %w", m.Slug, err)
		}
	}
	return nil
}

// applyHistoryRestore rewrites the pack's content and mods to match target.
func applyHistoryRestore(tx *gorm.DB, pack tables.Pack, target historyPayload, userID uint) error {
	var current []tables.Mod
	if err := tx.Where("pack_id = ?", pack.ID).Find(&current).Error; err != nil {
		return fmt.Errorf("load mods of pack %d: %w", pack.ID, err)
	}

	plan, err := planHistoryRestore(current, target)
	if err != nil {
		return err
	}

	if len(plan.Deletes) > 0 {
		if err := tx.Where("pack_id = ? AND id IN ?", pack.ID, plan.Deletes).Delete(&tables.Mod{}).Error; err != nil {
			return fmt.Errorf("remove mods: %w", err)
		}
	}

	for _, u := range plan.Updates {
		if err := tx.Model(&tables.Mod{}).
			Where("id = ? AND pack_id = ?", u.ID, pack.ID).
			Updates(historyModColumns(u.Mod, userID)).Error; err != nil {
			return fmt.Errorf("restore mod %s: %w", u.Mod.Slug, err)
		}
	}

	if err := insertHistoryMods(tx, pack.ID, userID, plan.Inserts); err != nil {
		return err
	}

	if err := syncHistoryDependencies(tx, pack.ID, userID, target.Mods); err != nil {
		return err
	}

	acceptable := datatypes.JSONSlice[string](target.Pack.AcceptableGameVersions)
	return tx.Model(&tables.Pack{}).
		Where("id = ?", pack.ID).
		Select(
			"Name", "Description", "Version", "PackFormat", "MCVersion", "Loader", "LoaderVersion",
			"AcceptableGameVersions", "UpdatedBy",
		).
		Updates(&tables.Pack{
			Name:                   target.Pack.Name,
			Description:            target.Pack.Description,
			Version:                target.Pack.Version,
			PackFormat:             target.Pack.PackFormat,
			MCVersion:              target.Pack.MCVersion,
			Loader:                 target.Pack.Loader,
			LoaderVersion:          target.Pack.LoaderVersion,
			AcceptableGameVersions: acceptable,
			UpdatedBy:              userID,
		}).Error
}

// RevertToSnapshot restores the pack's content to a snapshot.
//
// The snapshot becomes the head and every live snapshot after it is marked
// abandoned (kept, but no longer restorable); later changes chain from it. No
// new snapshot is written. Allowed in any pack status except archived, and not
// for an abandoned snapshot. A revert that finds the pack already identical to
// the head is a no-op.
func (ps *PackwizService) RevertToSnapshot(packId, snapshotId uint, user tables.User) (dto.RevertSnapshotResponse, response.ServerError) {
	unlock, lockErr := lockPackUpdate(packId)
	if lockErr != nil {
		return dto.RevertSnapshotResponse{}, lockErr
	}
	defer unlock()

	if active, err := ps.migrateJobActive(packId); err != nil {
		return dto.RevertSnapshotResponse{}, response.Wrap(err)
	} else if active {
		return dto.RevertSnapshotResponse{}, response.New(http.StatusConflict, "a migration is running for this pack")
	}

	changed := false
	if err := ps.db.Transaction(func(tx *gorm.DB) error {
		pack, err := lockPackRow(tx, packId)
		if err != nil {
			return err
		}
		if pack.DeletedAt.Valid {
			return errPackArchived
		}

		target, err := ps.loadSnapshot(tx, packId, snapshotId)
		if err != nil {
			return err
		}
		if target.AbandonedAt != nil {
			return errSnapshotAbandoned
		}

		payload, err := decodeHistoryPayload(target.SchemaVersion, target.Payload)
		if err != nil {
			return err
		}

		if guardErr := checkPublishedTargetChange(pack.Status, pack.MCVersion, pack.Loader, payload.Pack.MCVersion, payload.Pack.Loader); guardErr != nil {
			return revertBlockedError{guardErr}
		}

		current, err := loadHistoryPayload(tx, pack)
		if err != nil {
			return err
		}
		_, currentHash, err := encodeHistoryPayload(current)
		if err != nil {
			return err
		}
		if pack.HeadSnapshotID != nil && *pack.HeadSnapshotID == target.ID && currentHash == target.PayloadHash {
			return nil
		}
		changed = true

		if err := applyHistoryRestore(tx, pack, payload, user.ID); err != nil {
			return err
		}

		if err := tx.Model(&tables.PackSnapshot{}).
			Where("pack_id = ? AND abandoned_at IS NULL AND seq > ?", packId, target.Seq).
			Updates(map[string]any{
				"abandoned_at":           time.Now(),
				"abandoned_by":           user.ID,
				"abandoned_by_revert_to": target.ID,
			}).Error; err != nil {
			return fmt.Errorf("abandon snapshots after %d: %w", target.ID, err)
		}

		return setPackHead(tx, packId, target.ID)
	}); err != nil {
		return dto.RevertSnapshotResponse{}, snapshotErr(err)
	}

	if changed {
		// stored update checks described the state that was just replaced
		invalidatePackChecks(ps.db, packId)
	}

	return dto.RevertSnapshotResponse{Changed: changed, HeadID: snapshotId}, nil
}

// cloneDescription prepends a "cloned from" line to the description a clone
// inherits from its snapshot, e.g. "cloned from My Pack on 2026/10/03 14:05".
func cloneDescription(sourceName string, at time.Time, original string) string {
	note := fmt.Sprintf("cloned from %s on %s", sourceName, at.Format("2006/01/02 15:04"))
	if original == "" {
		return note
	}
	return note + "\n\n" + original
}

// CloneFromSnapshot creates a new draft pack from a snapshot's content. The
// caller becomes its owner. Abandoned snapshots can be cloned.
func (ps *PackwizService) CloneFromSnapshot(packId, snapshotId uint, request dto.CloneSnapshotRequest, user tables.User) (uint, response.ServerError) {
	snapshot, err := ps.loadSnapshot(ps.db, packId, snapshotId)
	if err != nil {
		return 0, snapshotErr(err)
	}
	payload, err := decodeHistoryPayload(snapshot.SchemaVersion, snapshot.Payload)
	if err != nil {
		return 0, response.Wrap(err)
	}

	if ps.PackExistsBySlug(request.Slug, true) {
		return 0, response.New(http.StatusConflict, "pack already exists")
	}

	var source tables.Pack
	if err := ps.db.Unscoped().Select("id", "name", "slug").First(&source, packId).Error; err != nil {
		return 0, response.Wrap(err)
	}
	sourceName := source.Name
	if sourceName == "" {
		sourceName = source.Slug
	}

	newPack := &tables.Pack{
		Slug:                   request.Slug,
		Name:                   request.Name,
		Description:            cloneDescription(sourceName, time.Now(), payload.Pack.Description),
		CreatedBy:              user.ID,
		UpdatedBy:              user.ID,
		IsPublic:               false,
		Status:                 types.PackStatusDraft,
		MCVersion:              payload.Pack.MCVersion,
		Loader:                 payload.Pack.Loader,
		LoaderVersion:          payload.Pack.LoaderVersion,
		AcceptableGameVersions: datatypes.JSONSlice[string](payload.Pack.AcceptableGameVersions),
		Version:                payload.Pack.Version,
		PackFormat:             payload.Pack.PackFormat,
	}

	if err := ps.db.Transaction(func(tx *gorm.DB) error {
		if err := CreatePackWithOwner(tx, newPack, user); err != nil {
			return err
		}
		if err := insertHistoryMods(tx, newPack.ID, user.ID, payload.Mods); err != nil {
			return err
		}
		return syncHistoryDependencies(tx, newPack.ID, user.ID, payload.Mods)
	}); err != nil {
		return 0, response.Wrap(err)
	}

	return newPack.ID, nil
}

// PruneSnapshots permanently deletes the pack's abandoned snapshots, the
// branches left behind by reverts. The live chain is untouched. Not allowed on
// an archived pack.
func (ps *PackwizService) PruneSnapshots(packId uint) (dto.PruneSnapshotsResponse, response.ServerError) {
	var deleted int64
	if err := ps.db.Transaction(func(tx *gorm.DB) error {
		pack, err := lockPackRow(tx, packId)
		if err != nil {
			return err
		}
		if pack.DeletedAt.Valid {
			return errPackArchived
		}

		// one statement: abandoned snapshots reference each other through parent_id
		result := tx.Where("pack_id = ? AND abandoned_at IS NOT NULL", packId).Delete(&tables.PackSnapshot{})
		if result.Error != nil {
			return fmt.Errorf("prune snapshots of pack %d: %w", packId, result.Error)
		}
		deleted = result.RowsAffected
		return nil
	}); err != nil {
		return dto.PruneSnapshotsResponse{}, snapshotErr(err)
	}

	return dto.PruneSnapshotsResponse{Deleted: deleted}, nil
}

// RebaseOnSnapshot makes a live snapshot the root of the pack's history and
// permanently deletes every snapshot before it, abandoned ones included. The
// root loses its parent and its summary is recomputed against an empty pack,
// as for a first snapshot. Snapshots after the root, and the pack itself, are
// untouched. Not allowed on an archived pack or an abandoned snapshot.
func (ps *PackwizService) RebaseOnSnapshot(packId, snapshotId uint) (dto.RebaseSnapshotResponse, response.ServerError) {
	var deleted int64
	if err := ps.db.Transaction(func(tx *gorm.DB) error {
		pack, err := lockPackRow(tx, packId)
		if err != nil {
			return err
		}
		if pack.DeletedAt.Valid {
			return errPackArchived
		}

		root, err := ps.loadSnapshot(tx, packId, snapshotId)
		if err != nil {
			return err
		}
		if root.AbandonedAt != nil {
			return errSnapshotAbandoned
		}

		payload, err := decodeHistoryPayload(root.SchemaVersion, root.Payload)
		if err != nil {
			return err
		}
		summaryJSON, err := json.Marshal(summarizeHistoryDiff(diffHistory(emptyHistoryPayload(), payload)))
		if err != nil {
			return fmt.Errorf("encode snapshot summary: %w", err)
		}

		if err := tx.Model(&tables.PackSnapshot{}).
			Where("id = ?", root.ID).
			Updates(map[string]any{"parent_id": nil, "summary": datatypes.JSON(summaryJSON)}).Error; err != nil {
			return fmt.Errorf("detach snapshot %d: %w", root.ID, err)
		}

		// one statement: the older snapshots reference each other through parent_id
		result := tx.Where("pack_id = ? AND seq < ?", packId, root.Seq).Delete(&tables.PackSnapshot{})
		if result.Error != nil {
			return fmt.Errorf("delete snapshots before %d: %w", root.ID, result.Error)
		}
		deleted = result.RowsAffected
		return nil
	}); err != nil {
		return dto.RebaseSnapshotResponse{}, snapshotErr(err)
	}

	return dto.RebaseSnapshotResponse{Deleted: deleted}, nil
}
