package jobs

import (
	"context"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// PackWebhookArgs calls the webhook of a pack after it changed.
type PackWebhookArgs struct {
	// PackID is the only uniqueness key: one pending call per pack.
	PackID uint `json:"packId" river:"unique"`
}

func (PackWebhookArgs) Kind() string { return "pack_webhook" }

func (PackWebhookArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		// A change that arrives while the call is running enqueues a new job,
		// so the receiver always hears about the latest state. Running is
		// therefore left out of ByState, unlike the compat scan.
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
			ByState: []rivertype.JobState{
				rivertype.JobStateAvailable,
				rivertype.JobStatePending,
				rivertype.JobStateRetryable,
				rivertype.JobStateScheduled,
			},
		},
		// network errors and 5xx answers are retried with River's backoff
		MaxAttempts: 5,
	}
}

// PackWebhookTimeout bounds one call. Longer than the HTTP client's own timeout.
const PackWebhookTimeout = 30 * time.Second

// PackWebhookCaller is implemented by webhook_svc.Service (interface here to
// avoid an import cycle, see UpdateChecker). A returned error is retried.
type PackWebhookCaller interface {
	RunWebhook(ctx context.Context, args PackWebhookArgs) error
}

type PackWebhookWorker struct {
	river.WorkerDefaults[PackWebhookArgs]
	caller PackWebhookCaller
}

func (w *PackWebhookWorker) Work(ctx context.Context, job *river.Job[PackWebhookArgs]) error {
	return w.caller.RunWebhook(ctx, job.Args)
}

func (w *PackWebhookWorker) Timeout(*river.Job[PackWebhookArgs]) time.Duration {
	return PackWebhookTimeout
}
