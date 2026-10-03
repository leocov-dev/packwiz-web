package jobs

import (
	"context"
	"time"

	"github.com/riverqueue/river"

	"packwiz-web/internal/config"
	"packwiz-web/internal/log"
)

// PrunePackAccessArgs deletes pack access rows older than the configured audit
// retention (config.C.AuditRetentionDays).
type PrunePackAccessArgs struct{}

func (PrunePackAccessArgs) Kind() string { return "prune_pack_access" }

func (PrunePackAccessArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		// a prune is idempotent; retrying a failed one just waits for the next run.
		MaxAttempts: 1,
		UniqueOpts:  river.UniqueOpts{ByPeriod: pruneAuditInterval},
	}
}

// PackAccessPruner is implemented by pack_access_svc.PackAccessService.
type PackAccessPruner interface {
	PruneOlderThan(ctx context.Context, cutoff time.Time) (int64, error)
}

type PrunePackAccessWorker struct {
	river.WorkerDefaults[PrunePackAccessArgs]
	pruner PackAccessPruner
}

func (w *PrunePackAccessWorker) Work(ctx context.Context, _ *river.Job[PrunePackAccessArgs]) error {
	days := config.C.AuditRetentionDays
	if days <= 0 {
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	deleted, err := w.pruner.PruneOlderThan(ctx, cutoff)
	if err != nil {
		return err
	}

	log.Info("pruned pack access rows:", deleted, "older than", days, "days")
	return nil
}

// packAccessPrunePeriodicJobs schedules the retention job daily, and once at
// startup. Pruning is disabled when AUDIT_RETENTION_DAYS is 0.
func packAccessPrunePeriodicJobs() []*river.PeriodicJob {
	if config.C.AuditRetentionDays <= 0 {
		return nil
	}

	return []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(pruneAuditInterval),
			func() (river.JobArgs, *river.InsertOpts) { return PrunePackAccessArgs{}, nil },
			&river.PeriodicJobOpts{RunOnStart: true},
		),
	}
}
