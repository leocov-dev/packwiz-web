package compat_svc

import (
	"context"
	"fmt"
	"time"

	"packwiz-web/internal/jobs"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
)

// SummariesForPacks returns the badge summary of each given pack that has a
// scan run; packs never scanned are absent from the map. It issues a fixed
// number of queries however many packs are asked for: the runs, then (only
// for packs whose run is done, the only state that needs the current target
// to detect staleness) the pack targets and their jar mods.
func (s *CompatService) SummariesForPacks(ctx context.Context, packIDs []uint) (map[uint]dto.CompatSummary, error) {
	if s == nil {
		return nil, jobs.ErrClientUnavailable
	}
	out := make(map[uint]dto.CompatSummary, len(packIDs))
	if len(packIDs) == 0 {
		return out, nil
	}
	db := s.db.WithContext(ctx)

	var runs []tables.PackCompatRun
	if err := db.Where("pack_id IN ?", packIDs).Find(&runs).Error; err != nil {
		return nil, fmt.Errorf("load compat runs: %w", err)
	}

	var doneIDs []uint
	for _, r := range runs {
		if r.Status == tables.CompatDone {
			doneIDs = append(doneIDs, r.PackID)
		}
	}

	hashes := map[uint]string{}
	if len(doneIDs) > 0 {
		var packs []tables.Pack
		if err := db.Unscoped().Select("id", "mc_version", "loader", "loader_version").
			Where("id IN ?", doneIDs).Find(&packs).Error; err != nil {
			return nil, fmt.Errorf("load packs for compat summaries: %w", err)
		}
		// same rows as the scan reads (jar mods); only the columns the hash needs
		var mods []tables.Mod
		if err := db.Select("pack_id", "slug", "type", "side", "download").
			Where("pack_id IN ? AND COALESCE(type, '') IN ?", doneIDs, []string{"", modTypeJar}).
			Find(&mods).Error; err != nil {
			return nil, fmt.Errorf("load mods for compat summaries: %w", err)
		}
		byPack := make(map[uint][]tables.Mod, len(doneIDs))
		for _, m := range mods {
			byPack[m.PackID] = append(byPack[m.PackID], m)
		}
		for _, p := range packs {
			hashes[p.ID] = TargetHash(targetInputOf(p, byPack[p.ID]))
		}
	}

	now := s.now()
	for i := range runs {
		out[runs[i].PackID] = summarize(&runs[i], hashes[runs[i].PackID], now)
	}
	return out, nil
}

// summarize builds the badge summary of one run. currentTargetHash is only
// read for done runs.
func summarize(run *tables.PackCompatRun, currentTargetHash string, now time.Time) dto.CompatSummary {
	return dto.CompatSummary{
		Status:   string(DeriveStatus(run, currentTargetHash, now)),
		Errors:   run.ErrorCount,
		Warnings: run.WarningCount,
	}
}
