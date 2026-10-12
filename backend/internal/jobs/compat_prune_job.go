package jobs

import (
	"context"
	"time"

	"github.com/riverqueue/river"

	"packwiz-web/internal/log"
)

// compatCatchupInterval is how often the catch-up job looks for packs that
// were never scanned.
const compatCatchupInterval = 15 * time.Minute

// CompatPruneArgs deletes unused mod_jar_meta cache rows.
type CompatPruneArgs struct{}

func (CompatPruneArgs) Kind() string { return "compat_prune" }

func (CompatPruneArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		// a prune is idempotent; retrying a failed one just waits for the next run.
		MaxAttempts: 1,
		UniqueOpts:  river.UniqueOpts{ByPeriod: pruneAuditInterval},
	}
}

// CompatPruner is implemented by compat_svc.CompatService.
type CompatPruner interface {
	PruneJarMeta(ctx context.Context) (int64, error)
}

type CompatPruneWorker struct {
	river.WorkerDefaults[CompatPruneArgs]
	pruner CompatPruner
}

func (w *CompatPruneWorker) Work(ctx context.Context, _ *river.Job[CompatPruneArgs]) error {
	deleted, err := w.pruner.PruneJarMeta(ctx)
	if err != nil {
		return err
	}

	log.Info("pruned jar metadata rows:", deleted)
	return nil
}

// CompatCatchupArgs enqueues scans for packs that have no compatibility run
// yet (existing packs when the feature is first deployed, packs created by
// paths that do not trigger a scan).
type CompatCatchupArgs struct{}

func (CompatCatchupArgs) Kind() string { return "compat_catchup" }

func (CompatCatchupArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		MaxAttempts: 1,
		UniqueOpts:  river.UniqueOpts{ByPeriod: compatCatchupInterval},
	}
}

// CompatCatchup is implemented by compat_svc.CompatService.
type CompatCatchup interface {
	EnqueueMissing(ctx context.Context) (int, error)
}

type CompatCatchupWorker struct {
	river.WorkerDefaults[CompatCatchupArgs]
	catchup CompatCatchup
}

func (w *CompatCatchupWorker) Work(ctx context.Context, _ *river.Job[CompatCatchupArgs]) error {
	queued, err := w.catchup.EnqueueMissing(ctx)
	if err != nil {
		return err
	}
	if queued > 0 {
		log.Info("queued compatibility scans for unscanned packs:", queued)
	}
	return nil
}

// compatPeriodicJobs schedules the daily jar metadata prune and the periodic
// catch-up, both also once at startup.
func compatPeriodicJobs() []*river.PeriodicJob {
	return []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(pruneAuditInterval),
			func() (river.JobArgs, *river.InsertOpts) { return CompatPruneArgs{}, nil },
			&river.PeriodicJobOpts{RunOnStart: true},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(compatCatchupInterval),
			func() (river.JobArgs, *river.InsertOpts) { return CompatCatchupArgs{}, nil },
			&river.PeriodicJobOpts{RunOnStart: true},
		),
	}
}
