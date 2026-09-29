package jobs

import (
	"context"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// QueueUpdateChecks is a dedicated queue capped at 1 worker (see client.go)
// for CheckUpdatesArgs jobs: update checks hit third-party APIs (GitHub is
// rate limited to 60/h unauthenticated) and the packwiz-nxt updaters are not
// verified safe for concurrent use.
const QueueUpdateChecks = "update_checks"

// CheckUpdatesArgs checks every mod of a pack for an available update
// (read-only; nothing is applied) and stores the per-mod results.
type CheckUpdatesArgs struct {
	// PackID is the only uniqueness key: one queued/running check per pack.
	PackID uint `json:"packId" river:"unique"`
	UserID uint `json:"userId"`
}

func (CheckUpdatesArgs) Kind() string { return "check_updates" }

func (CheckUpdatesArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: QueueUpdateChecks,
		// completed jobs are intentionally not in ByState so a finished check
		// never blocks the next one.
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
		// check failures are recorded on the run row; retrying would only
		// burn API rate limit.
		MaxAttempts: 1,
	}
}

// UpdateChecker is implemented by PackwizService (interface here to avoid an
// import cycle, see MigrateModsResolver).
type UpdateChecker interface {
	RunUpdateCheck(ctx context.Context, args CheckUpdatesArgs, jobId int64) error
}

type CheckUpdatesWorker struct {
	river.WorkerDefaults[CheckUpdatesArgs]
	checker UpdateChecker
}

func (w *CheckUpdatesWorker) Work(ctx context.Context, job *river.Job[CheckUpdatesArgs]) error {
	return w.checker.RunUpdateCheck(ctx, job.Args, job.ID)
}
