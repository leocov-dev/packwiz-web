package packwiz_svc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"

	"github.com/leocov-dev/packwiz-nxt/core"

	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
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
// map is JSON-encoded (map keys sorted) because the updaters may mutate it in
// place, and because JSON normalizes numeric types (uint32 vs float64 from a
// decoded TOML/JSON round trip) that fmt's %v would render differently.
func takeModSnapshot(mod *core.Mod) (modSnapshot, error) {
	update, err := json.Marshal(mod.Update)
	if err != nil {
		return modSnapshot{}, fmt.Errorf("snapshot mod %s: %w", mod.Slug, err)
	}
	return modSnapshot{
		fileName: mod.FileName,
		download: mod.Download,
		update:   string(update),
	}, nil
}

// changedSince reports whether mod differs from the snapshot.
func (s modSnapshot) changedSince(mod *core.Mod) (bool, error) {
	now, err := takeModSnapshot(mod)
	if err != nil {
		return false, err
	}
	return s != now, nil
}

var (
	packUpdateLocksMu sync.Mutex
	packUpdateLocks   = map[uint]*sync.Mutex{}
)

// errUpdateInProgress is returned when an update is already running for the pack.
func errUpdateInProgress() *response.HttpError {
	return response.New(http.StatusConflict, "an update is already in progress for this pack")
}

// lockPackUpdate takes the in-process per-pack update guard without blocking.
// It returns a 409 error if an update is already running for packId, else an
// unlock func. It only guards within one process (a single backend instance).
func lockPackUpdate(packId uint) (func(), response.ServerError) {
	packUpdateLocksMu.Lock()
	l, ok := packUpdateLocks[packId]
	if !ok {
		l = &sync.Mutex{}
		packUpdateLocks[packId] = l
	}
	packUpdateLocksMu.Unlock()

	if !l.TryLock() {
		return nil, errUpdateInProgress()
	}
	return l.Unlock, nil
}

// updateAllSummary accumulates per-mod outcomes of an update-all run.
type updateAllSummary struct {
	updated  []dto.UpdateAllItem
	skipped  []dto.UpdateAllItem
	failed   []dto.UpdateAllItem
	upToDate int
	// notChecked counts mods whose update status is unknown: no registered
	// updater (e.g. manual URL sources) or a failed check on a pinned mod.
	notChecked int
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

func (s *updateAllSummary) addNotChecked() { s.notChecked++ }

// response builds the DTO with deterministic ordering and non-nil slices.
func (s *updateAllSummary) response() dto.UpdateAllResponse {
	sorted := func(items []dto.UpdateAllItem) []dto.UpdateAllItem {
		out := make([]dto.UpdateAllItem, len(items))
		copy(out, items)
		sort.SliceStable(out, func(i, j int) bool { return out[i].Slug < out[j].Slug })
		return out
	}
	return dto.UpdateAllResponse{
		Updated:    sorted(s.updated),
		Skipped:    sorted(s.skipped),
		Failed:     sorted(s.failed),
		UpToDate:   s.upToDate,
		NotChecked: s.notChecked,
	}
}
