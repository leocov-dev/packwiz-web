package packwiz_svc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/leocov-dev/packwiz-nxt/core"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"packwiz-web/internal/jobs"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

const (
	// updateCheckTTL is how long a finished check is served instead of
	// starting a new one (unless forced).
	updateCheckTTL = 10 * time.Minute
	// updateCheckStaleAfter is how long a queued/running check may go without
	// finishing before it is treated as dead (crashed worker) and replaceable.
	updateCheckStaleAfter = 30 * time.Minute
)

// isActiveRun reports whether run is a live queued/running check at now.
func isActiveRun(run *tables.PackUpdateCheckRun, now time.Time) bool {
	if run == nil {
		return false
	}
	if run.Status != tables.UpdateCheckQueued && run.Status != tables.UpdateCheckRunning {
		return false
	}
	return now.Sub(run.StartedAt) <= updateCheckStaleAfter
}

// runStatus maps a stored run to the API status. A queued/running run that
// went stale is reported as failed.
func runStatus(run *tables.PackUpdateCheckRun, now time.Time) string {
	switch {
	case run == nil:
		return dto.UpdateCheckIdle
	case run.Status == tables.UpdateCheckQueued || run.Status == tables.UpdateCheckRunning:
		if isActiveRun(run, now) {
			return run.Status
		}
		return dto.UpdateCheckFailed
	default:
		return run.Status
	}
}

// isFreshDone reports whether run is a finished check younger than the TTL.
func isFreshDone(run *tables.PackUpdateCheckRun, now time.Time) bool {
	return run != nil &&
		run.Status == tables.UpdateCheckDone &&
		run.FinishedAt != nil &&
		now.Sub(*run.FinishedAt) < updateCheckTTL
}

// checkRow is a stored check joined with the mod's current pinned flag.
type checkRow struct {
	tables.ModUpdateCheck
	Pinned bool
}

// buildUpdateChecksResponse assembles the API response from stored state. The
// available count excludes pinned mods and mods whose check failed, using the
// mods' current pinned flag.
func buildUpdateChecksResponse(run *tables.PackUpdateCheckRun, rows []checkRow, now time.Time) dto.UpdateChecksResponse {
	out := dto.UpdateChecksResponse{
		Status:  runStatus(run, now),
		Results: make([]dto.UpdateCheckItem, 0, len(rows)),
	}
	if run != nil {
		out.CheckedAt = run.FinishedAt
		out.Error = run.Error
	}
	for _, r := range rows {
		out.Results = append(out.Results, dto.UpdateCheckItem{
			ModId:           r.ModID,
			UpdateAvailable: r.UpdateAvailable,
			UpdateString:    r.UpdateString,
			Error:           r.Error,
		})
		if r.UpdateAvailable && !r.Pinned && r.Error == "" {
			out.AvailableCount++
		}
	}
	return out
}

// buildCheckRows maps nxt check results to rows for the pack's mods. Results
// for mods no longer in the pack are dropped; a failed check stores the error
// and never an update.
func buildCheckRows(packId uint, modsBySlug map[string]tables.Mod, results []core.UpdateCheckResult, now time.Time) []tables.ModUpdateCheck {
	rows := make([]tables.ModUpdateCheck, 0, len(results))
	for _, r := range results {
		if r.Mod == nil {
			continue
		}
		dbMod, ok := modsBySlug[r.Mod.Slug]
		if !ok {
			continue
		}
		row := tables.ModUpdateCheck{PackID: packId, ModID: dbMod.ID, CheckedAt: now}
		if r.Err != nil {
			row.Error = r.Err.Error()
		} else {
			row.UpdateAvailable = r.UpdateAvailable
			row.UpdateString = r.UpdateString
		}
		rows = append(rows, row)
	}
	return rows
}

func (ps *PackwizService) getCheckRun(db *gorm.DB, packId uint) (*tables.PackUpdateCheckRun, error) {
	var run tables.PackUpdateCheckRun
	err := db.Where("pack_id = ?", packId).First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// GetUpdateChecks returns the stored update-check state and results of a pack.
func (ps *PackwizService) GetUpdateChecks(ctx context.Context, packId uint) (dto.UpdateChecksResponse, response.ServerError) {
	db := ps.db.WithContext(ctx)

	run, err := ps.getCheckRun(db, packId)
	if err != nil {
		return dto.UpdateChecksResponse{}, response.Wrap(err)
	}

	var rows []checkRow
	if err := db.Table("mod_update_checks AS c").
		Select("c.*, m.pinned AS pinned").
		Joins("JOIN mods m ON m.id = c.mod_id").
		Where("c.pack_id = ?", packId).
		Order("c.mod_id").
		Scan(&rows).Error; err != nil {
		return dto.UpdateChecksResponse{}, response.Wrap(err)
	}

	return buildUpdateChecksResponse(run, rows, time.Now()), nil
}

// RequestUpdateCheck enqueues an update-check job for the pack. It returns
// enqueued=false (and the existing job) when a check is already queued or
// running, or when a check finished within updateCheckTTL and force is false.
func (ps *PackwizService) RequestUpdateCheck(ctx context.Context, packId uint, user tables.User, force bool) (dto.UpdateCheckJobResponse, bool, response.ServerError) {
	if ps.riverClient == nil {
		return dto.UpdateCheckJobResponse{}, false, response.New(http.StatusInternalServerError, "background jobs are not available")
	}

	now := time.Now()
	run, err := ps.getCheckRun(ps.db.WithContext(ctx), packId)
	if err != nil {
		return dto.UpdateCheckJobResponse{}, false, response.Wrap(err)
	}
	if isActiveRun(run, now) || (!force && isFreshDone(run, now)) {
		return dto.UpdateCheckJobResponse{JobId: run.JobID, Status: runStatus(run, now)}, false, nil
	}

	// Mark queued first (job id unknown yet) so the worker's own running/done
	// updates can never be overwritten by this write.
	queued := tables.PackUpdateCheckRun{PackID: packId, Status: tables.UpdateCheckQueued, StartedAt: now}
	if err := ps.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "pack_id"}},
		UpdateAll: true,
	}).Select("*").Create(&queued).Error; err != nil {
		return dto.UpdateCheckJobResponse{}, false, response.Wrap(err)
	}

	result, insertErr := ps.riverClient.Insert(ctx, jobs.CheckUpdatesArgs{PackID: packId, UserID: user.ID}, nil)
	if insertErr != nil {
		ps.markCheckFailed(packId, 0, fmt.Errorf("failed to enqueue: %w", insertErr))
		return dto.UpdateCheckJobResponse{}, false, response.Wrap(insertErr)
	}

	// job_id = 0 guard: the worker may already have set its own job id.
	if err := ps.db.WithContext(ctx).Model(&tables.PackUpdateCheckRun{}).
		Where("pack_id = ? AND job_id = 0", packId).
		Update("job_id", result.Job.ID).Error; err != nil {
		log.Error("failed to record update check job id:", err)
	}

	return dto.UpdateCheckJobResponse{JobId: result.Job.ID, Status: tables.UpdateCheckQueued}, !result.UniqueSkippedAsDuplicate, nil
}

// markCheckFailed records a failed run. Best effort: errors are only logged.
func (ps *PackwizService) markCheckFailed(packId uint, jobId int64, cause error) {
	finished := time.Now()
	run := tables.PackUpdateCheckRun{
		PackID:     packId,
		JobID:      jobId,
		Status:     tables.UpdateCheckFailed,
		StartedAt:  finished,
		FinishedAt: &finished,
		Error:      cause.Error(),
	}
	if err := ps.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "pack_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "finished_at", "error"}),
	}).Create(&run).Error; err != nil {
		log.Error("failed to record failed update check:", err)
	}
}

// RunUpdateCheck implements jobs.UpdateChecker. It runs core.CheckAllMods
// (read-only; per-mod failures are stored, not fatal) and replaces the pack's
// stored results in one transaction. Failures are recorded on the run row and
// not returned, so River does not retry and burn API rate limit.
//
// It deliberately does not take the per-pack update lock: checks only read.
// A check overlapping Update All may report already-applied updates until the
// next check; UpdateAll clears the stored results when it finishes.
func (ps *PackwizService) RunUpdateCheck(ctx context.Context, args jobs.CheckUpdatesArgs, jobId int64) error {
	started := time.Now()
	if err := ps.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "pack_id"}},
		UpdateAll: true,
	}).Select("*").Create(&tables.PackUpdateCheckRun{
		PackID: args.PackID, JobID: jobId, Status: tables.UpdateCheckRunning, StartedAt: started,
	}).Error; err != nil {
		return fmt.Errorf("mark update check running: %w", err)
	}

	if err := ps.runUpdateCheck(ctx, args.PackID, jobId); err != nil {
		log.Error(fmt.Errorf("update check failed for pack %d: %w", args.PackID, err))
		ps.markCheckFailed(args.PackID, jobId, err)
	}
	return nil
}

func (ps *PackwizService) runUpdateCheck(ctx context.Context, packId uint, jobId int64) error {
	dbPack, srvErr := ps.GetPackById(packId)
	if srvErr != nil {
		return srvErr
	}

	results, err := core.CheckAllMods(nil, dbPack.AsMeta())
	if err != nil {
		return err
	}

	modsBySlug := make(map[string]tables.Mod, len(dbPack.Mods))
	for _, m := range dbPack.Mods {
		modsBySlug[m.Slug] = m
	}

	now := time.Now()
	rows := buildCheckRows(packId, modsBySlug, results, now)

	return ps.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("pack_id = ?", packId).Delete(&tables.ModUpdateCheck{}).Error; err != nil {
			return err
		}
		if len(rows) > 0 {
			if err := tx.Create(&rows).Error; err != nil {
				return err
			}
		}
		return tx.Model(&tables.PackUpdateCheckRun{}).
			Where("pack_id = ? AND job_id = ?", packId, jobId).
			Updates(map[string]any{"status": tables.UpdateCheckDone, "finished_at": now, "error": ""}).Error
	})
}

// invalidateModChecks drops the stored check rows of the given mods. Best
// effort: failure only leaves a stale badge, so it is logged, not returned.
func invalidateModChecks(db *gorm.DB, modIDs ...uint) {
	if len(modIDs) == 0 {
		return
	}
	if err := db.Where("mod_id IN ?", modIDs).Delete(&tables.ModUpdateCheck{}).Error; err != nil {
		log.Error("failed to invalidate update checks:", err)
	}
}

// invalidatePackChecks drops all stored check results of a pack and resets its
// run to idle, unless a check is currently queued/running (its results will
// replace these anyway).
func invalidatePackChecks(db *gorm.DB, packId uint) {
	if err := db.Where("pack_id = ?", packId).Delete(&tables.ModUpdateCheck{}).Error; err != nil {
		log.Error("failed to invalidate pack update checks:", err)
		return
	}
	if err := db.Where("pack_id = ? AND status NOT IN ?", packId,
		[]string{tables.UpdateCheckQueued, tables.UpdateCheckRunning}).
		Delete(&tables.PackUpdateCheckRun{}).Error; err != nil {
		log.Error("failed to reset update check run:", err)
	}
}
