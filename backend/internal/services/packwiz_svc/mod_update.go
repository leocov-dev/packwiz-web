package packwiz_svc

import (
	"fmt"
	"sort"

	"github.com/leocov-dev/packwiz-nxt/core"

	"packwiz-web/internal/types/dto"
)

// SkipReasonPinned is the reason reported for pinned mods with an available update.
const SkipReasonPinned = "pinned"

// modSnapshot captures the mod fields an update can change, so a mod can be
// compared before and after an update to tell whether anything really changed.
type modSnapshot struct {
	fileName string
	download core.ModDownload
	update   string
}

// takeModSnapshot deep-captures the update-relevant fields of mod. The update
// map is rendered to a string (fmt prints maps with sorted keys) because the
// updaters may mutate it in place.
func takeModSnapshot(mod *core.Mod) modSnapshot {
	return modSnapshot{
		fileName: mod.FileName,
		download: mod.Download,
		update:   fmt.Sprintf("%v", mod.Update),
	}
}

// changedSince reports whether mod differs from the snapshot.
func (s modSnapshot) changedSince(mod *core.Mod) bool {
	return s != takeModSnapshot(mod)
}

// updateAllSummary accumulates per-mod outcomes of an update-all run.
type updateAllSummary struct {
	updated  []dto.UpdateAllItem
	skipped  []dto.UpdateAllItem
	failed   []dto.UpdateAllItem
	upToDate int
}

func (s *updateAllSummary) addUpdated(item dto.UpdateAllItem) { s.updated = append(s.updated, item) }

func (s *updateAllSummary) addSkipped(item dto.UpdateAllItem, reason string) {
	item.Reason = reason
	s.skipped = append(s.skipped, item)
}

func (s *updateAllSummary) addFailed(item dto.UpdateAllItem, err error) {
	item.Error = err.Error()
	s.failed = append(s.failed, item)
}

func (s *updateAllSummary) addUpToDate() { s.upToDate++ }

// response builds the DTO with deterministic ordering and non-nil slices.
func (s *updateAllSummary) response() dto.UpdateAllResponse {
	sorted := func(items []dto.UpdateAllItem) []dto.UpdateAllItem {
		out := make([]dto.UpdateAllItem, len(items))
		copy(out, items)
		sort.SliceStable(out, func(i, j int) bool { return out[i].Slug < out[j].Slug })
		return out
	}
	return dto.UpdateAllResponse{
		Updated:  sorted(s.updated),
		Skipped:  sorted(s.skipped),
		Failed:   sorted(s.failed),
		UpToDate: s.upToDate,
	}
}
