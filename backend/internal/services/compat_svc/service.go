package compat_svc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"packwiz-web/internal/jobs"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
)

// CompatService runs and reports the jar compatibility check. It never changes
// pack content or history.
type CompatService struct {
	db *gorm.DB
	// riverClient enqueues scans. It is nil for instances built as job
	// resolvers; those enqueue through the client River puts in a worker
	// context (see client).
	riverClient *river.Client[*sql.Tx]
	fetcher     JarFetcher
	// flight makes concurrent resolutions of the same jar hash share one
	// download. The scan queue has one worker, so this only matters when
	// scans of two packs ever run at once (a raised concurrency, tests); it
	// is cheap and keeps that case from downloading a jar twice.
	flight singleflight.Group
	// failures remembers recent download failures (see negativeCache).
	failures *negativeCache
	now      func() time.Time
}

// NewCompatService builds the service. riverClient may be nil (job resolver
// instances). fetcher may be nil, which selects the production HTTPJarFetcher.
func NewCompatService(db *gorm.DB, riverClient *river.Client[*sql.Tx], fetcher JarFetcher) *CompatService {
	if fetcher == nil {
		fetcher = NewHTTPJarFetcher()
	}
	return &CompatService{db: db, riverClient: riverClient, fetcher: fetcher, failures: newNegativeCache(), now: time.Now}
}

// A nil *CompatService is a valid "no compat service": every method that needs
// the service returns jobs.ErrClientUnavailable instead of panicking. This
// matters because a nil *CompatService stored in an interface (a
// ScanRequester) is a non-nil interface, so callers cannot detect it.

// client returns the River client to enqueue with: the injected one, else the
// one River stores in a worker's context.
func (s *CompatService) client(ctx context.Context) (*river.Client[*sql.Tx], error) {
	if s == nil {
		return nil, jobs.ErrClientUnavailable
	}
	if s.riverClient != nil {
		return s.riverClient, nil
	}
	if c, err := river.ClientFromContextSafely[*sql.Tx](ctx); err == nil {
		return c, nil
	}
	return nil, jobs.ErrClientUnavailable
}

// RequestScan schedules an automatic scan of a pack, ScanDelay from now. It is
// the call for pack changes: repeated calls while a scan is pending collapse
// into the pending job (the job is unique per pack), so bursts of edits cost
// one scan. Archived and unknown packs are skipped. Call it after the change
// has committed, never inside its transaction.
func (s *CompatService) RequestScan(ctx context.Context, packID uint) error {
	if s == nil {
		return jobs.ErrClientUnavailable
	}
	_, err := s.enqueue(ctx, packID, scheduledAt(s.now()))
	return err
}

// ManualScanResult is the outcome of RequestManualScan.
type ManualScanResult struct {
	// Enqueued is false when a scan is already active or the last run is
	// younger than ManualScanMinInterval.
	Enqueued bool
	// Status is the pack's badge status after the request.
	Status Status
}

// ErrPackArchived is returned by RequestManualScan for an archived pack.
var ErrPackArchived = errors.New("pack is archived")

// RequestManualScan starts a scan on a user's request: scheduled right away,
// but at most one per ManualScanMinInterval per pack and never while a scan is
// queued or running. A missing pack is gorm.ErrRecordNotFound (wrapped), an
// archived one ErrPackArchived. When no scan is enqueued the result carries
// the status GetCompat would report.
func (s *CompatService) RequestManualScan(ctx context.Context, packID uint) (ManualScanResult, error) {
	if s == nil {
		return ManualScanResult{}, jobs.ErrClientUnavailable
	}
	now := s.now()
	status, run, err := s.liveStatus(ctx, packID, now)
	if err != nil {
		return ManualScanResult{}, err
	}
	if !shouldEnqueueManual(run, now) {
		return ManualScanResult{Status: status}, nil
	}
	enqueued, err := s.enqueue(ctx, packID, now)
	if err != nil {
		return ManualScanResult{}, err
	}
	if enqueued {
		return ManualScanResult{Enqueued: true, Status: StatusQueued}, nil
	}
	// Not enqueued: the pack was archived or deleted since the check above, or
	// a scan job is already pending. Look again to tell which.
	status, _, err = s.liveStatus(ctx, packID, s.now())
	if err != nil {
		return ManualScanResult{}, err
	}
	return ManualScanResult{Status: status}, nil
}

// liveStatus loads the pack, its mods and its run, and derives the status
// against the pack's current content, as GetCompat does. Missing pack:
// gorm.ErrRecordNotFound (wrapped); archived pack: ErrPackArchived.
func (s *CompatService) liveStatus(ctx context.Context, packID uint, now time.Time) (Status, *tables.PackCompatRun, error) {
	pack, mods, err := loadScanInput(s.db.WithContext(ctx), packID, true)
	if err != nil {
		return "", nil, err
	}
	if pack.DeletedAt.Valid {
		return "", nil, fmt.Errorf("compat scan of pack %d: %w", packID, ErrPackArchived)
	}
	run, err := s.getRun(ctx, packID)
	if err != nil {
		return "", nil, err
	}
	return statusFor(run, pack, mods, now), run, nil
}

// statusFor derives the badge status of a run against the pack's current
// content. It is the one place that combines DeriveStatus with the target
// hash, shared by GetCompat, RequestManualScan and the list summaries, so they
// always agree. mods may hold all of the pack's mods or only its jars.
func statusFor(run *tables.PackCompatRun, pack tables.Pack, mods []tables.Mod, now time.Time) Status {
	return DeriveStatus(run, TargetHash(targetInputOf(pack, mods)), now)
}

// enqueue inserts the scan job for a pack and marks its run queued. It reports
// false when the pack is archived or missing, or a scan job is already pending.
//
// Run row protocol. pack_compat_runs.status moves queued -> running -> done or
// failed, written by enqueue and by RunScan (worker). A change that arrives
// while a scan job is active cannot get a job of its own (jobs are unique per
// pack, including running ones), so it must make the active scan's result
// count as "needs another scan" by flipping the run to queued; the catch-up job
// re-enqueues a queued run that stayed queued for orphanQueuedAfter. The rules
// that make this race free:
//
//   - The flip applies from running, done and failed, so it works in every
//     phase of the active job: before RunScan wrote its running row (status
//     still the previous result), while scanning, and after store/markFailed
//     but before River marked the job completed.
//   - The flip before RunScan's running upsert is overwritten by that upsert.
//     That is correct: the change committed before the flip, hence before the
//     scan reads the pack.
//   - store and markFailed never replace a queued status, so a flip during the
//     scan survives it.
//   - The queued upsert after a fresh Insert does not overwrite a row that
//     already belongs to this job (a fast worker may have written running or
//     even finished before the upsert runs), see below.
//   - If the job completed between Insert and the flip, the run is left queued
//     with no job coming; catch-up repairs that after orphanQueuedAfter.
func (s *CompatService) enqueue(ctx context.Context, packID uint, at time.Time) (bool, error) {
	client, err := s.client(ctx)
	if err != nil {
		return false, err
	}

	// the default scope excludes archived (soft-deleted) packs
	var live int64
	if err := s.db.WithContext(ctx).Model(&tables.Pack{}).Where("id = ?", packID).Count(&live).Error; err != nil {
		return false, fmt.Errorf("check pack %d: %w", packID, err)
	}
	if live == 0 {
		return false, nil
	}

	result, err := client.Insert(ctx, jobs.CompatScanArgs{PackID: packID}, &river.InsertOpts{ScheduledAt: at})
	if err != nil {
		return false, fmt.Errorf("enqueue compat scan of pack %d: %w", packID, err)
	}
	if result.UniqueSkippedAsDuplicate {
		if result.Job.State != rivertype.JobStateRunning {
			// the pending job already has its queued run row
			return false, nil
		}
		// A scan is running and may have read the pack before this change:
		// ask for a rescan (see the protocol above).
		if err := s.db.WithContext(ctx).Model(&tables.PackCompatRun{}).
			Where("pack_id = ? AND status IN ?", packID, []string{tables.CompatRunning, tables.CompatDone, tables.CompatFailed}).
			Updates(map[string]any{"status": tables.CompatQueued, "started_at": s.now()}).Error; err != nil {
			return false, fmt.Errorf("mark compat scan of pack %d for rescan: %w", packID, err)
		}
		return false, nil
	}

	if err := s.markQueued(ctx, packID, result.Job.ID); err != nil {
		return false, err
	}
	return true, nil
}

// markQueued records the run of a freshly inserted job as queued.
//
// The job may already be running, or even finished, before this upsert
// executes (the insert is committed and the worker is fast). A row that
// carries this job's id was written by the worker: leave it alone, or the
// queued status would hide a running scan and cause a second scan.
func (s *CompatService) markQueued(ctx context.Context, packID uint, jobID int64) error {
	queued := tables.PackCompatRun{PackID: packID, JobID: jobID, Status: tables.CompatQueued, StartedAt: s.now()}
	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "pack_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"job_id", "status", "started_at"}),
		Where:     clause.Where{Exprs: []clause.Expression{clause.Expr{SQL: "pack_compat_runs.job_id <> ?", Vars: []any{jobID}}}},
	}).Create(&queued).Error; err != nil {
		return fmt.Errorf("mark compat scan of pack %d queued: %w", packID, err)
	}
	return nil
}

// EnqueueMissing schedules scans for packs that need one but have none coming:
// packs with no run yet, packs whose run was marked queued for a rescan more
// than orphanQueuedAfter ago (a change arrived during a scan, see enqueue), and
// packs whose run has been running for more than runStaleAfter (a crashed
// worker). catchupNeeded is the same rule for one run. At most catchupBatch per call, spaced catchupStagger apart so many
// packs do not hit the worker at once. Each queued pack gets a fresh run row,
// so the next call picks up the following batch. A pack whose job is in fact
// still pending is skipped by the job's uniqueness. It returns how many scans
// were queued.
func (s *CompatService) EnqueueMissing(ctx context.Context) (int, error) {
	if s == nil {
		return 0, jobs.ErrClientUnavailable
	}
	var ids []uint
	if err := s.db.WithContext(ctx).Model(&tables.Pack{}).
		Where(`NOT EXISTS (SELECT 1 FROM pack_compat_runs r WHERE r.pack_id = packs.id)
			OR EXISTS (SELECT 1 FROM pack_compat_runs r WHERE r.pack_id = packs.id
				AND ((r.status = ? AND r.started_at < ?) OR (r.status = ? AND r.started_at < ?)))`,
			tables.CompatQueued, s.now().Add(-orphanQueuedAfter),
			tables.CompatRunning, s.now().Add(-runStaleAfter)).
		Order("packs.id").
		Limit(catchupBatch).
		Pluck("packs.id", &ids).Error; err != nil {
		return 0, fmt.Errorf("find unscanned packs: %w", err)
	}

	now := s.now()
	queued := 0
	for i, id := range ids {
		ok, err := s.enqueue(ctx, id, catchupSchedule(now, i))
		if err != nil {
			if errors.Is(err, jobs.ErrClientUnavailable) {
				return queued, err
			}
			log.Error(fmt.Errorf("catch-up compat scan of pack %d: %w", id, err))
			continue
		}
		if ok {
			queued++
		}
	}
	return queued, nil
}

func (s *CompatService) getRun(ctx context.Context, packID uint) (*tables.PackCompatRun, error) {
	var runs []tables.PackCompatRun
	if err := s.db.WithContext(ctx).Where("pack_id = ?", packID).Limit(1).Find(&runs).Error; err != nil {
		return nil, fmt.Errorf("load compat run of pack %d: %w", packID, err)
	}
	if len(runs) == 0 {
		return nil, nil
	}
	return &runs[0], nil
}

// Report is the stored compatibility state of a pack.
type Report struct {
	// Run is nil when the pack was never scanned.
	Run      *tables.PackCompatRun
	Findings []tables.PackCompatFinding
	// Status is the badge status, derived against the pack's current content.
	Status Status
}

// GetCompat returns the stored run and findings of a pack with the derived
// status. Archived packs are included. gorm.ErrRecordNotFound (wrapped) means
// the pack does not exist.
func (s *CompatService) GetCompat(ctx context.Context, packID uint) (Report, error) {
	if s == nil {
		return Report{}, jobs.ErrClientUnavailable
	}
	db := s.db.WithContext(ctx)

	pack, mods, err := loadScanInput(db, packID, true)
	if err != nil {
		return Report{}, err
	}
	run, err := s.getRun(ctx, packID)
	if err != nil {
		return Report{}, err
	}

	report := Report{
		Run:      run,
		Findings: []tables.PackCompatFinding{},
		Status:   statusFor(run, pack, mods, s.now()),
	}
	if run == nil {
		return report, nil
	}
	if err := db.Where("pack_id = ?", packID).Order("id").Find(&report.Findings).Error; err != nil {
		return Report{}, fmt.Errorf("load compat findings of pack %d: %w", packID, err)
	}
	return report, nil
}

// loadScanInput loads the pack and its mods. With includeArchived false an
// archived pack is gorm.ErrRecordNotFound.
func loadScanInput(db *gorm.DB, packID uint, includeArchived bool) (tables.Pack, []tables.Mod, error) {
	q := db
	if includeArchived {
		q = q.Unscoped()
	}
	var pack tables.Pack
	if err := q.Select("id", "mc_version", "loader", "loader_version", "deleted_at").First(&pack, packID).Error; err != nil {
		return tables.Pack{}, nil, fmt.Errorf("load pack %d: %w", packID, err)
	}
	var mods []tables.Mod
	if err := db.Where("pack_id = ?", packID).Order("id").Find(&mods).Error; err != nil {
		return tables.Pack{}, nil, fmt.Errorf("load mods of pack %d: %w", packID, err)
	}
	return pack, mods, nil
}
