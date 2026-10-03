package jobs

import (
	"github.com/riverqueue/river"
)

// NewWorkers builds the set of registered River workers. resolver implements
// the migrate job logic, checker the update-check job logic, pruner the
// audit retention job logic, accessPruner the pack access retention job logic. Every process
// that runs jobs (web --worker, worker) builds its workers through here.
// Add new river.AddWorker(workers, &SomeWorker{...}) calls here as real jobs
// are added.
func NewWorkers(resolver MigrateModsResolver, checker UpdateChecker, pruner AuditPruner, accessPruner PackAccessPruner) *river.Workers {
	workers := river.NewWorkers()

	river.AddWorker(workers, &MigrateModsWorker{resolver: resolver})
	river.AddWorker(workers, &CheckUpdatesWorker{checker: checker})
	river.AddWorker(workers, &PruneAuditWorker{pruner: pruner})

	river.AddWorker(workers, &PrunePackAccessWorker{pruner: accessPruner})

	return workers
}
