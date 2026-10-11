package jobs

import (
	"context"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// QueueBulkAdd is a dedicated queue capped at 1 worker (see client.go) for
// BulkAddArgs jobs: they hit third-party search/version APIs once or more per
// line, and the packwiz-nxt sources are not verified safe for concurrent use.
const QueueBulkAdd = "bulk_add"

// BulkAddArgs resolves the lines of a bulk add run in list order: exact
// matches are added to the pack, the rest are left for review.
type BulkAddArgs struct {
	// PackID is the uniqueness key: one queued/running bulk add per pack, so
	// two runs never race over which line was first.
	PackID uint `json:"packId" river:"unique"`
	RunID  uint `json:"runId"`
	UserID uint `json:"userId"`
}

func (BulkAddArgs) Kind() string { return "bulk_add" }

func (BulkAddArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: QueueBulkAdd,
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
		// failures are recorded on the run and item rows; a retry would add
		// the same lines again.
		MaxAttempts: 1,
	}
}

// BulkAdder is implemented by PackwizService (interface here to avoid an
// import cycle, see MigrateModsResolver).
type BulkAdder interface {
	RunBulkAdd(ctx context.Context, args BulkAddArgs, jobId int64) error
}

// BulkAddTimeout bounds one bulk add run. The client's RescueStuckJobsAfter
// (client.go) must exceed it.
const BulkAddTimeout = 10 * time.Minute

type BulkAddWorker struct {
	river.WorkerDefaults[BulkAddArgs]
	adder BulkAdder
}

func (w *BulkAddWorker) Work(ctx context.Context, job *river.Job[BulkAddArgs]) error {
	return w.adder.RunBulkAdd(ctx, job.Args, job.ID)
}

// Timeout overrides River's 1 minute default job timeout: a long list does
// several API calls per line.
func (w *BulkAddWorker) Timeout(*river.Job[BulkAddArgs]) time.Duration {
	return BulkAddTimeout
}
