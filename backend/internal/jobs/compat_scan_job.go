package jobs

import (
	"context"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// QueueCompatScan is a dedicated queue capped at 1 worker (see client.go) for
// CompatScanArgs jobs: a scan downloads mod jars (up to 150 MiB each) and the
// cap bounds bandwidth, scratch disk and parse work across all packs.
const QueueCompatScan = "compat_scan"

// CompatScanArgs scans every mod jar of a pack and stores the compatibility
// findings (read-only for pack content; nothing is changed or snapshotted).
type CompatScanArgs struct {
	// PackID is the only uniqueness key: one pending scan per pack.
	PackID uint `json:"packId" river:"unique"`
}

func (CompatScanArgs) Kind() string { return "compat_scan" }

func (CompatScanArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: QueueCompatScan,
		// River requires running in ByState. So a trigger that arrives while a
		// scan runs is skipped as a duplicate; CompatService marks the run
		// queued again and the catch-up job re-enqueues it once this job has
		// finished. Completed jobs are not in ByState, so they never block.
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
			ByState: []rivertype.JobState{
				rivertype.JobStateAvailable,
				rivertype.JobStatePending,
				rivertype.JobStateRunning,
				rivertype.JobStateRetryable,
				rivertype.JobStateScheduled,
			},
		},
		// failures are recorded on the run row; a retry would only repeat
		// the downloads.
		MaxAttempts: 1,
	}
}

// CompatScanner is implemented by compat_svc.CompatService (interface here to
// avoid an import cycle, see UpdateChecker).
type CompatScanner interface {
	RunScan(ctx context.Context, args CompatScanArgs, jobId int64) error
}

// CompatScanTimeout bounds one scan. River cancels the job context after this;
// the client's RescueStuckJobsAfter (client.go) must exceed it.
const CompatScanTimeout = 10 * time.Minute

type CompatScanWorker struct {
	river.WorkerDefaults[CompatScanArgs]
	scanner CompatScanner
}

func (w *CompatScanWorker) Work(ctx context.Context, job *river.Job[CompatScanArgs]) error {
	return w.scanner.RunScan(ctx, job.Args, job.ID)
}

// Timeout overrides River's 1 minute default job timeout: a pack may hold
// hundreds of jars to download.
func (w *CompatScanWorker) Timeout(*river.Job[CompatScanArgs]) time.Duration {
	return CompatScanTimeout
}
