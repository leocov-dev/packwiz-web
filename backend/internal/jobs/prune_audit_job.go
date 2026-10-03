package jobs

import (
	"context"
	"time"

	"github.com/riverqueue/river"

	"packwiz-web/internal/config"
	"packwiz-web/internal/log"
)

// pruneAuditInterval is how often the retention job runs.
const pruneAuditInterval = 24 * time.Hour

// PruneAuditArgs deletes audit rows older than the configured retention
// (config.C.AuditRetentionDays).
type PruneAuditArgs struct{}

func (PruneAuditArgs) Kind() string { return "prune_audit" }

func (PruneAuditArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		// a prune is idempotent; retrying a failed one just waits for the next run.
		MaxAttempts: 1,
		UniqueOpts:  river.UniqueOpts{ByPeriod: pruneAuditInterval},
	}
}

// AuditPruner is implemented by audit_svc.AuditService.
type AuditPruner interface {
	PruneOlderThan(ctx context.Context, cutoff time.Time) (int64, error)
}

type PruneAuditWorker struct {
	river.WorkerDefaults[PruneAuditArgs]
	pruner AuditPruner
}

func (w *PruneAuditWorker) Work(ctx context.Context, _ *river.Job[PruneAuditArgs]) error {
	days := config.C.AuditRetentionDays
	if days <= 0 {
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	deleted, err := w.pruner.PruneOlderThan(ctx, cutoff)
	if err != nil {
		return err
	}

	log.Info("pruned audit rows:", deleted, "older than", days, "days")
	return nil
}

// auditPrunePeriodicJobs schedules the retention job daily, and once at
// startup. Pruning is disabled when AUDIT_RETENTION_DAYS is 0.
func auditPrunePeriodicJobs() []*river.PeriodicJob {
	if config.C.AuditRetentionDays <= 0 {
		return nil
	}

	return []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(pruneAuditInterval),
			func() (river.JobArgs, *river.InsertOpts) { return PruneAuditArgs{}, nil },
			&river.PeriodicJobOpts{RunOnStart: true},
		),
	}
}
