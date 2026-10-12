package compat_svc

import (
	"context"
	"fmt"
	"time"

	"github.com/leocov-dev/packwiz-nxt/jarmeta"
	"github.com/lib/pq"
	"gorm.io/datatypes"

	"packwiz-web/internal/jobs"
)

// pruneBatch is how many cache rows one DELETE removes.
const pruneBatch = 500

// pruneRowRecord is mod_jar_meta as the prune query reads it.
type pruneRowRecord struct {
	HashFormat string
	Hash       string
	LastUsedAt time.Time
	Loaders    pq.StringArray                           `gorm:"type:text[]"`
	Meta       *datatypes.JSONSlice[jarmeta.LoaderMeta] `gorm:"type:jsonb"`
}

// PruneJarMeta implements jobs.CompatPruner: it deletes mod_jar_meta rows that
// pruneCandidates allows. SQL narrows to rows unused for PruneAfter; the
// referenced-hash set (every mod of every pack, drafts and archived included)
// and the pack targets are loaded for the Go decision. It returns how many
// rows were deleted.
func (s *CompatService) PruneJarMeta(ctx context.Context) (int64, error) {
	if s == nil {
		return 0, jobs.ErrClientUnavailable
	}
	db := s.db.WithContext(ctx)
	now := s.now()
	cutoff := now.Add(-PruneAfter)

	var records []pruneRowRecord
	if err := db.Table("mod_jar_meta").
		Select("hash_format", "hash", "last_used_at", "loaders", "meta").
		Where("last_used_at < ?", cutoff).
		Find(&records).Error; err != nil {
		return 0, fmt.Errorf("load prune candidates: %w", err)
	}
	if len(records) == 0 {
		return 0, nil
	}

	var refs []struct{ HashFormat, Hash string }
	if err := db.Raw(`SELECT DISTINCT lower(download->>'hashFormat') AS hash_format, lower(download->>'hash') AS hash FROM mods`).
		Scan(&refs).Error; err != nil {
		return 0, fmt.Errorf("load referenced jar hashes: %w", err)
	}
	referenced := make(map[JarKey]struct{}, len(refs))
	for _, r := range refs {
		referenced[JarKey{HashFormat: r.HashFormat, Hash: r.Hash}] = struct{}{}
	}

	// raw SQL: archived packs count too, so no soft-delete scope
	var pairs []struct{ MCVersion, Loader string }
	if err := db.Raw(`SELECT DISTINCT mc_version, loader FROM packs`).Scan(&pairs).Error; err != nil {
		return 0, fmt.Errorf("load pack targets: %w", err)
	}
	targets := make([]PackTarget, 0, len(pairs))
	for _, p := range pairs {
		targets = append(targets, PackTarget{MinecraftVersion: p.MCVersion, Loader: p.Loader})
	}

	rows := make([]PruneRow, 0, len(records))
	for _, r := range records {
		row := PruneRow{Key: JarKey{HashFormat: r.HashFormat, Hash: r.Hash}, LastUsedAt: r.LastUsedAt, Loaders: r.Loaders}
		if r.Meta != nil {
			row.Meta = *r.Meta
		}
		rows = append(rows, row)
	}

	doomed := pruneCandidates(rows, referenced, targets, now)

	var deleted int64
	for start := 0; start < len(doomed); start += pruneBatch {
		batch := doomed[start:min(start+pruneBatch, len(doomed))]
		tuples := make([][]any, len(batch))
		for i, k := range batch {
			tuples[i] = []any{k.HashFormat, k.Hash}
		}
		// last_used_at is checked again: a scan may have touched a row since it was read
		res := db.Exec(`DELETE FROM mod_jar_meta WHERE (hash_format, hash) IN ? AND last_used_at < ?`, tuples, cutoff)
		if res.Error != nil {
			return deleted, fmt.Errorf("delete jar metadata: %w", res.Error)
		}
		deleted += res.RowsAffected
	}
	return deleted, nil
}
