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
	// download. One CompatService runs scans per process, so this is
	// process-wide in practice.
	flight singleflight.Group
	now    func() time.Time
}

// NewCompatService builds the service. riverClient may be nil (job resolver
// instances). fetcher may be nil, which selects the production HTTPJarFetcher.
func NewCompatService(db *gorm.DB, riverClient *river.Client[*sql.Tx], fetcher JarFetcher) *CompatService {
	if fetcher == nil {
		fetcher = NewHTTPJarFetcher()
	}
	return &CompatService{db: db, riverClient: riverClient, fetcher: fetcher, now: time.Now}
}

// client returns the River client to enqueue with: the injected one, else the
// one River stores in a worker's context.
func (s *CompatService) client(ctx context.Context) (*river.Client[*sql.Tx], error) {
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

// RequestManualScan starts a scan on a user's request: scheduled right away,
// but at most one per ManualScanMinInterval per pack and never while a scan is
// queued or running.
func (s *CompatService) RequestManualScan(ctx context.Context, packID uint) (ManualScanResult, error) {
	now := s.now()
	run, err := s.getRun(ctx, packID)
	if err != nil {
		return ManualScanResult{}, err
	}
	if !shouldEnqueueManual(run, now) {
		return ManualScanResult{Status: statusOf(run, now)}, nil
	}
	enqueued, err := s.enqueue(ctx, packID, now)
	if err != nil {
		return ManualScanResult{}, err
	}
	return ManualScanResult{Enqueued: enqueued, Status: StatusQueued}, nil
}

// statusOf is DeriveStatus for callers that only care about run activity: a
// stored result is reported as is (the current target is not needed).
func statusOf(run *tables.PackCompatRun, now time.Time) Status {
	if run == nil {
		return StatusUnknown
	}
	return DeriveStatus(run, run.TargetHash, now)
}

// enqueue inserts the scan job for a pack and marks its run queued. It reports
// false when the pack is archived or missing, or a scan job is already pending.
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

	var result *rivertype.JobInsertResult
	for attempt := 0; ; attempt++ {
		result, err = client.Insert(ctx, jobs.CompatScanArgs{PackID: packID}, &river.InsertOpts{ScheduledAt: at})
		if err != nil {
			return false, fmt.Errorf("enqueue compat scan of pack %d: %w", packID, err)
		}
		if !result.UniqueSkippedAsDuplicate {
			break
		}
		if result.Job.State != rivertype.JobStateRunning {
			// the pending job already has its queued run row
			return false, nil
		}
		// A scan is running and may have read the pack before this change.
		// Unique jobs include running ones, so no job can be queued now: flip
		// the run back to queued instead. The running scan keeps that status
		// when it finishes (see store) and the catch-up job re-enqueues it.
		flipped := s.db.WithContext(ctx).Model(&tables.PackCompatRun{}).
			Where("pack_id = ? AND status = ?", packID, tables.CompatRunning).
			Updates(map[string]any{"status": tables.CompatQueued, "started_at": s.now()})
		if flipped.Error != nil {
			return false, fmt.Errorf("mark compat scan of pack %d for rescan: %w", packID, flipped.Error)
		}
		if flipped.RowsAffected > 0 || attempt > 0 {
			return false, nil
		}
		// the scan finished in between: insert again, it is no longer a duplicate
	}

	queued := tables.PackCompatRun{PackID: packID, JobID: result.Job.ID, Status: tables.CompatQueued, StartedAt: s.now()}
	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "pack_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"job_id", "status", "started_at"}),
	}).Create(&queued).Error; err != nil {
		return false, fmt.Errorf("mark compat scan of pack %d queued: %w", packID, err)
	}
	return true, nil
}

// EnqueueMissing schedules scans for packs that need one but have none coming:
// packs with no run yet, and packs whose run was marked queued for a rescan
// more than orphanQueuedAfter ago (a change arrived during a scan, see
// enqueue). At most catchupBatch per call, spaced catchupStagger apart so many
// packs do not hit the worker at once. Each queued pack gets a fresh run row,
// so the next call picks up the following batch. A pack whose job is in fact
// still pending is skipped by the job's uniqueness. It returns how many scans
// were queued.
func (s *CompatService) EnqueueMissing(ctx context.Context) (int, error) {
	var ids []uint
	if err := s.db.WithContext(ctx).Model(&tables.Pack{}).
		Where(`NOT EXISTS (SELECT 1 FROM pack_compat_runs r WHERE r.pack_id = packs.id)
			OR EXISTS (SELECT 1 FROM pack_compat_runs r WHERE r.pack_id = packs.id AND r.status = ? AND r.started_at < ?)`,
			tables.CompatQueued, s.now().Add(-orphanQueuedAfter)).
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
		Status:   DeriveStatus(run, TargetHash(targetInputOf(pack, mods)), s.now()),
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
