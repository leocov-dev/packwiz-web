package packwiz_svc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/leocov-dev/packwiz-nxt/core"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"packwiz-web/internal/jobs"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

// bulkAddStaleAfter is how long a queued/running bulk add may go without
// finishing before it is treated as dead (crashed worker). Must exceed
// jobs.BulkAddTimeout plus queueing time.
const bulkAddStaleAfter = 30 * time.Minute

// isActiveBulkRun reports whether run is a live queued/running bulk add at now.
func isActiveBulkRun(run *tables.PackBulkAddRun, now time.Time) bool {
	if run == nil {
		return false
	}
	if run.Status != tables.BulkAddQueued && run.Status != tables.BulkAddRunning {
		return false
	}
	return now.Sub(run.CreatedAt) <= bulkAddStaleAfter
}

// bulkRunStatus maps a stored run to the API status. A queued/running run that
// went stale is reported as failed.
func bulkRunStatus(run *tables.PackBulkAddRun, now time.Time) string {
	if (run.Status == tables.BulkAddQueued || run.Status == tables.BulkAddRunning) && !isActiveBulkRun(run, now) {
		return tables.BulkAddFailed
	}
	return run.Status
}

// isReviewableItem reports whether a line is waiting for the user to pick a
// mod or reject it.
func isReviewableItem(status string) bool {
	switch status {
	case tables.BulkItemPending, tables.BulkItemIncompatibleVersion, tables.BulkItemConflict:
		return true
	}
	return false
}

func bulkItemResponse(it tables.PackBulkAddItem) dto.BulkAddItem {
	candidates := make([]dto.BulkAddCandidate, 0, len(it.Candidates))
	for _, c := range it.Candidates {
		candidates = append(candidates, dto.BulkAddCandidate{
			Source: c.Source, Slug: c.Slug, Title: c.Title, ProjectId: c.ProjectId, IconUrl: c.IconUrl,
		})
	}
	return dto.BulkAddItem{
		Id:                    it.ID,
		Position:              it.Position,
		Line:                  it.Line,
		Status:                it.Status,
		Source:                it.Source,
		Slug:                  it.Slug,
		Name:                  it.Name,
		ProjectId:             it.ProjectID,
		Reason:                it.Reason,
		ConflictsWithPosition: it.ConflictsWithPosition,
		Candidates:            candidates,
	}
}

func buildBulkAddResponse(run tables.PackBulkAddRun, items []tables.PackBulkAddItem, now time.Time) dto.BulkAddRunResponse {
	out := dto.BulkAddRunResponse{
		Id:         run.ID,
		Status:     bulkRunStatus(&run, now),
		Error:      run.Error,
		CreatedAt:  run.CreatedAt,
		StartedAt:  run.StartedAt,
		FinishedAt: run.FinishedAt,
		Items:      make([]dto.BulkAddItem, 0, len(items)),
	}
	for _, it := range items {
		out.Items = append(out.Items, bulkItemResponse(it))
	}
	return out
}

// bulkItemUpdate is the column update that stores a line's final outcome.
// Candidates are only kept for lines left for review.
func bulkItemUpdate(m lineMatch, o itemOutcome) map[string]any {
	ref := m.Ref
	if o.Match != nil {
		ref = o.Match.Ref
	}
	// never nil: a nil slice is stored as JSON null instead of []
	candidates := []tables.BulkAddCandidate{}
	if isReviewableItem(o.Status) && m.Candidates != nil {
		candidates = m.Candidates
	}
	return map[string]any{
		"status":                  o.Status,
		"reason":                  o.Reason,
		"conflicts_with_position": o.ConflictsWith,
		"source":                  ref.Source,
		"slug":                    ref.Slug,
		"name":                    ref.Name,
		"project_id":              ref.ProjectID,
		"candidates":              datatypes.NewJSONSlice(candidates),
		"updated_at":              time.Now(),
	}
}

func (ps *PackwizService) latestBulkRun(db *gorm.DB, packId uint) (*tables.PackBulkAddRun, error) {
	var run tables.PackBulkAddRun
	err := db.Where("pack_id = ?", packId).Order("id DESC").First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func (ps *PackwizService) bulkAddResponse(db *gorm.DB, run tables.PackBulkAddRun) (dto.BulkAddRunResponse, error) {
	var items []tables.PackBulkAddItem
	if err := db.Where("run_id = ?", run.ID).Order("position").Find(&items).Error; err != nil {
		return dto.BulkAddRunResponse{}, err
	}
	return buildBulkAddResponse(run, items, time.Now()), nil
}

// GetBulkAdd returns a bulk add run of the pack with its lines.
func (ps *PackwizService) GetBulkAdd(ctx context.Context, packId, runId uint) (dto.BulkAddRunResponse, response.ServerError) {
	db := ps.db.WithContext(ctx)
	var run tables.PackBulkAddRun
	if err := db.Where("id = ? AND pack_id = ?", runId, packId).First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.BulkAddRunResponse{}, response.New(http.StatusNotFound, fmt.Sprintf("bulk add %d not found", runId))
		}
		return dto.BulkAddRunResponse{}, response.Wrap(err)
	}
	out, err := ps.bulkAddResponse(db, run)
	if err != nil {
		return dto.BulkAddRunResponse{}, response.Wrap(err)
	}
	return out, nil
}

// GetLatestBulkAdd returns the pack's most recent bulk add run, nil if none.
func (ps *PackwizService) GetLatestBulkAdd(ctx context.Context, packId uint) (*dto.BulkAddRunResponse, response.ServerError) {
	db := ps.db.WithContext(ctx)
	run, err := ps.latestBulkRun(db, packId)
	if err != nil {
		return nil, response.Wrap(err)
	}
	if run == nil {
		return nil, nil
	}
	out, err := ps.bulkAddResponse(db, *run)
	if err != nil {
		return nil, response.Wrap(err)
	}
	return &out, nil
}

// RequestBulkAdd stores the lines of a pasted list as a new run and enqueues
// the job that resolves them. Only one bulk add per pack may be active.
func (ps *PackwizService) RequestBulkAdd(ctx context.Context, packId uint, user tables.User, request dto.BulkAddRequest) (dto.BulkAddRunResponse, response.ServerError) {
	if ps.riverClient == nil {
		return dto.BulkAddRunResponse{}, response.New(http.StatusInternalServerError, "background jobs are not available")
	}

	lines := parseBulkLines(request.Text)
	if len(lines) == 0 {
		return dto.BulkAddRunResponse{}, response.New(http.StatusBadRequest, "the list has no mods")
	}
	if len(lines) > maxBulkLines {
		return dto.BulkAddRunResponse{}, response.New(http.StatusBadRequest, fmt.Sprintf("the list has %d mods, the limit is %d", len(lines), maxBulkLines))
	}

	db := ps.db.WithContext(ctx)
	latest, err := ps.latestBulkRun(db, packId)
	if err != nil {
		return dto.BulkAddRunResponse{}, response.Wrap(err)
	}
	if isActiveBulkRun(latest, time.Now()) {
		return dto.BulkAddRunResponse{}, response.New(http.StatusConflict, "a bulk add is already running for this pack")
	}

	run := tables.PackBulkAddRun{PackID: packId, Status: tables.BulkAddQueued, CreatedBy: user.ID}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&run).Error; err != nil {
			return err
		}
		items := make([]tables.PackBulkAddItem, 0, len(lines))
		for i, line := range lines {
			items = append(items, tables.PackBulkAddItem{
				RunID: run.ID, Position: i + 1, Line: line, Status: tables.BulkItemQueued,
				Candidates: datatypes.NewJSONSlice([]tables.BulkAddCandidate{}),
			})
		}
		return tx.Create(&items).Error
	}); err != nil {
		return dto.BulkAddRunResponse{}, response.Wrap(err)
	}

	result, insertErr := ps.riverClient.Insert(ctx, jobs.BulkAddArgs{PackID: packId, RunID: run.ID, UserID: user.ID}, nil)
	if insertErr != nil {
		ps.markBulkRunFailed(run.ID, fmt.Errorf("failed to enqueue: %w", insertErr))
		return dto.BulkAddRunResponse{}, response.Wrap(insertErr)
	}
	if result.UniqueSkippedAsDuplicate {
		// a concurrent request won the race for this pack
		ps.markBulkRunFailed(run.ID, errors.New("another bulk add is already running for this pack"))
		return dto.BulkAddRunResponse{}, response.New(http.StatusConflict, "a bulk add is already running for this pack")
	}

	// job_id = 0 guard: the worker may already have set its own job id.
	if err := db.Model(&tables.PackBulkAddRun{}).
		Where("id = ? AND job_id = 0", run.ID).
		Update("job_id", result.Job.ID).Error; err != nil {
		log.Error("failed to record bulk add job id:", err)
	}

	return ps.GetBulkAdd(ctx, packId, run.ID)
}

// markBulkRunFailed records a failed run. Best effort: errors are only logged.
func (ps *PackwizService) markBulkRunFailed(runId uint, cause error) {
	if err := ps.db.Model(&tables.PackBulkAddRun{}).Where("id = ?", runId).Updates(map[string]any{
		"status":      tables.BulkAddFailed,
		"finished_at": time.Now(),
		"error":       cause.Error(),
	}).Error; err != nil {
		log.Error("failed to record failed bulk add:", err)
	}
}

// RunBulkAdd implements jobs.BulkAdder. It looks up every line, applies list
// order (resolveInOrder), adds the accepted mods in one history snapshot and
// stores each line's outcome. Failures are recorded on the run and not
// returned, so River does not retry.
func (ps *PackwizService) RunBulkAdd(ctx context.Context, args jobs.BulkAddArgs, jobId int64) error {
	started := time.Now()
	res := ps.db.WithContext(ctx).Model(&tables.PackBulkAddRun{}).
		Where("id = ? AND pack_id = ? AND status = ?", args.RunID, args.PackID, tables.BulkAddQueued).
		Updates(map[string]any{"status": tables.BulkAddRunning, "started_at": started, "job_id": jobId})
	if res.Error != nil || res.RowsAffected == 0 {
		markErr := fmt.Errorf("mark bulk add %d running: %v (rows %d)", args.RunID, res.Error, res.RowsAffected)
		log.Error(markErr)
		ps.markBulkRunFailed(args.RunID, markErr)
		return nil
	}

	if err := ps.runBulkAdd(ctx, args); err != nil {
		log.Error(fmt.Errorf("bulk add %d failed for pack %d: %w", args.RunID, args.PackID, err))
		ps.markBulkRunFailed(args.RunID, err)
	}
	return nil
}

func (ps *PackwizService) runBulkAdd(ctx context.Context, args jobs.BulkAddArgs) error {
	var user tables.User
	if err := ps.db.WithContext(ctx).First(&user, args.UserID).Error; err != nil {
		return fmt.Errorf("load user %d: %w", args.UserID, err)
	}

	dbPack, srvErr := ps.GetPackById(args.PackID)
	if srvErr != nil {
		return srvErr
	}
	if dbPack.DeletedAt.Valid {
		return errors.New("the pack is archived")
	}

	var items []tables.PackBulkAddItem
	if err := ps.db.WithContext(ctx).
		Where("run_id = ? AND status = ?", args.RunID, tables.BulkItemQueued).
		Order("position").Find(&items).Error; err != nil {
		return err
	}

	matcher := newBulkMatcher(dbPack)
	matches := make([]lineMatch, 0, len(items))
	for _, it := range items {
		if ctx.Err() != nil {
			// out of time: later lines lose to earlier ones anyway
			matches = append(matches, lineMatch{Position: it.Position, Status: tables.BulkItemFailed,
				Reason: "the run timed out before this line was looked up"})
			continue
		}
		matches = append(matches, matcher.match(it.Position, it.Line))
	}

	installed, installedSlugs := installedRefs(dbPack.Mods)
	outcomes := resolveInOrder(matches, installed, installedSlugs)

	// the writes below use ps.db without the job context: it may be spent,
	// and the results must still be stored
	write := func(tx *gorm.DB) error {
		skipped, err := insertBulkMods(tx, dbPack, user, outcomes)
		if err != nil {
			return err
		}
		for i, it := range items {
			o := outcomes[i]
			if skipped[o.Position] {
				o.Status, o.Reason, o.Match = tables.BulkItemSkipped, "already in the pack", nil
			}
			if err := tx.Model(&tables.PackBulkAddItem{}).Where("id = ?", it.ID).
				Updates(bulkItemUpdate(matches[i], o)).Error; err != nil {
				return err
			}
		}
		return tx.Model(&tables.PackBulkAddRun{}).Where("id = ?", args.RunID).
			Updates(map[string]any{"status": tables.BulkAddDone, "finished_at": time.Now(), "error": ""}).Error
	}

	var added []string
	for _, o := range outcomes {
		if o.Status == tables.BulkItemAdded {
			added = append(added, o.Match.Ref.Name)
		}
	}
	if len(added) == 0 {
		return ps.db.Transaction(write)
	}
	detail := map[string]any{"count": len(added), "names": added}
	return ps.withPackHistory(args.PackID, user.ID, tables.SnapshotModBulkAdd, detail, write)
}

// insertBulkMods creates the mods of the added lines, in list order, with
// their dependencies. A dependency shared by several lines is created once.
// Lines whose mod appeared in the pack since the lookup are returned as
// skipped.
func insertBulkMods(tx *gorm.DB, dbPack tables.Pack, user tables.User, outcomes []itemOutcome) (map[int]bool, error) {
	var existing []tables.Mod
	if err := tx.Select("id", "slug").Where("pack_id = ?", dbPack.ID).Find(&existing).Error; err != nil {
		return nil, err
	}
	ids := make(map[string]uint, len(existing))
	for _, m := range existing {
		ids[m.Slug] = m.ID
	}
	skipped := map[int]bool{}

	for _, o := range outcomes {
		if o.Status != tables.BulkItemAdded {
			continue
		}
		mod := o.Match.Mod
		if _, ok := ids[mod.Slug]; ok {
			skipped[o.Position] = true
			continue
		}

		var dependencyIds []uint
		for _, dep := range o.Match.Deps {
			if id, ok := ids[dep.Slug]; ok {
				dependencyIds = append(dependencyIds, id)
				continue
			}
			dbDep, err := createMod(dep, dbPack, user, tx, true, nil)
			if err != nil {
				return nil, err
			}
			ids[dep.Slug] = dbDep.ID
			dependencyIds = append(dependencyIds, dbDep.ID)
		}

		dbMod, err := createMod(mod, dbPack, user, tx, false, dependencyIds)
		if err != nil {
			return nil, err
		}
		ids[mod.Slug] = dbMod.ID
	}
	return skipped, nil
}

// ResolveBulkItem reviews one line of a finished bulk add: reject dismisses
// it; pick adds the chosen mod (with its dependencies, one history snapshot)
// unless it is already in the pack or conflicts with an installed mod (409).
func (ps *PackwizService) ResolveBulkItem(
	ctx context.Context,
	packId, runId, itemId uint,
	user tables.User,
	request dto.ResolveBulkItemRequest,
) (dto.BulkAddItem, response.ServerError) {
	db := ps.db.WithContext(ctx)

	var run tables.PackBulkAddRun
	if err := db.Where("id = ? AND pack_id = ?", runId, packId).First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.BulkAddItem{}, response.New(http.StatusNotFound, fmt.Sprintf("bulk add %d not found", runId))
		}
		return dto.BulkAddItem{}, response.Wrap(err)
	}
	if isActiveBulkRun(&run, time.Now()) {
		return dto.BulkAddItem{}, response.New(http.StatusConflict, "the bulk add is still running")
	}

	var item tables.PackBulkAddItem
	if err := db.Where("id = ? AND run_id = ?", itemId, runId).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.BulkAddItem{}, response.New(http.StatusNotFound, fmt.Sprintf("bulk add line %d not found", itemId))
		}
		return dto.BulkAddItem{}, response.Wrap(err)
	}
	if !isReviewableItem(item.Status) {
		return dto.BulkAddItem{}, response.New(http.StatusConflict, "this line is already resolved")
	}

	if request.Action == dto.BulkItemActionReject {
		if err := db.Model(&item).Updates(map[string]any{
			"status": tables.BulkItemRejected, "updated_at": time.Now(),
		}).Error; err != nil {
			return dto.BulkAddItem{}, response.Wrap(err)
		}
		return ps.bulkItem(db, item.ID)
	}

	dbPack, srvErr := ps.GetPackById(packId)
	if srvErr != nil {
		return dto.BulkAddItem{}, srvErr
	}
	resolved, err := resolvePick(request, dbPack.AsMeta())
	if err != nil {
		if errors.Is(err, errNoCompatibleVersion) {
			name := resolved.Ref.Name
			if name == "" {
				name = "the mod"
			}
			return dto.BulkAddItem{}, response.New(http.StatusConflict, fmt.Sprintf("%s has no version for this pack's Minecraft version and loader", name))
		}
		return dto.BulkAddItem{}, response.Wrap(err)
	}

	installed, installedSlugs := installedRefs(dbPack.Mods)
	outcome := resolveInOrder([]lineMatch{{Position: item.Position, Match: &resolved}}, installed, installedSlugs)[0]
	if outcome.Status != tables.BulkItemAdded {
		return dto.BulkAddItem{}, response.New(http.StatusConflict, fmt.Sprintf("%s: %s", resolved.Ref.Name, outcome.Reason))
	}

	detail := map[string]any{"slug": resolved.Mod.Slug, "name": resolved.Mod.Name}
	if err := ps.withPackHistory(packId, user.ID, tables.SnapshotModAdd, detail, func(tx *gorm.DB) error {
		if err := insertModWithDeps(tx, dbPack, user, resolved.Mod, resolved.Deps); err != nil {
			return err
		}
		return tx.Model(&tables.PackBulkAddItem{}).Where("id = ?", item.ID).Updates(map[string]any{
			"status":                  tables.BulkItemPicked,
			"reason":                  "",
			"conflicts_with_position": 0,
			"source":                  resolved.Ref.Source,
			"slug":                    resolved.Ref.Slug,
			"name":                    resolved.Ref.Name,
			"project_id":              resolved.Ref.ProjectID,
			"updated_at":              time.Now(),
		}).Error
	}); err != nil {
		return dto.BulkAddItem{}, response.Wrap(err)
	}

	return ps.bulkItem(db, item.ID)
}

func resolvePick(request dto.ResolveBulkItemRequest, pack core.Pack) (resolvedMod, error) {
	if request.Source == sourceCurseforge {
		id, err := strconv.ParseUint(request.ProjectId, 10, 32)
		if err != nil {
			return resolvedMod{}, err
		}
		return resolveCurseforgeByID(uint32(id), pack)
	}
	key := request.Slug
	if key == "" {
		key = request.ProjectId
	}
	return resolveModrinthMod(key, pack)
}

func (ps *PackwizService) bulkItem(db *gorm.DB, itemId uint) (dto.BulkAddItem, response.ServerError) {
	var item tables.PackBulkAddItem
	if err := db.First(&item, itemId).Error; err != nil {
		return dto.BulkAddItem{}, response.Wrap(err)
	}
	return bulkItemResponse(item), nil
}
