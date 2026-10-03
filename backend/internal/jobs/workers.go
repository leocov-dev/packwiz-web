package jobs

import (
	"github.com/riverqueue/river"
)

// NewWorkers builds the set of registered River workers. resolver implements
// the migrate job logic, checker the update-check job logic, pruner the
// audit retention job logic. Every process
// that runs jobs (web --worker, worker) builds its workers through here.
// Add new river.AddWorker(workers, &SomeWorker{...}) calls here as real jobs
// are added.
func NewWorkers(resolver MigrateModsResolver, checker UpdateChecker, pruner AuditPruner) *river.Workers {
	workers := river.NewWorkers()

	river.AddWorker(workers, &MigrateModsWorker{resolver: resolver})
	river.AddWorker(workers, &CheckUpdatesWorker{checker: checker})
	river.AddWorker(workers, &PruneAuditWorker{pruner: pruner})

	return workers
}
