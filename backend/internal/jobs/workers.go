package jobs

import (
	"github.com/riverqueue/river"
)

// NewWorkers builds the set of registered River workers. resolver implements
// the migrate job logic, checker the update-check job logic, adder the bulk
// add job logic, pruner the audit retention job logic, accessPruner the pack
// access retention job logic, compatScanner/compatPruner/compatCatchup the
// jar compatibility scan, its cache prune and its catch-up logic. Every process that runs jobs (web --worker,
// worker) builds its workers through here.
// Add new river.AddWorker(workers, &SomeWorker{...}) calls here as real jobs
// are added.
func NewWorkers(resolver MigrateModsResolver, checker UpdateChecker, adder BulkAdder, pruner AuditPruner, accessPruner PackAccessPruner, compatScanner CompatScanner, compatPruner CompatPruner, compatCatchup CompatCatchup) *river.Workers {
	workers := river.NewWorkers()

	river.AddWorker(workers, &MigrateModsWorker{resolver: resolver})
	river.AddWorker(workers, &CheckUpdatesWorker{checker: checker})
	river.AddWorker(workers, &BulkAddWorker{adder: adder})
	river.AddWorker(workers, &PruneAuditWorker{pruner: pruner})

	river.AddWorker(workers, &PrunePackAccessWorker{pruner: accessPruner})

	river.AddWorker(workers, &CompatScanWorker{scanner: compatScanner})
	river.AddWorker(workers, &CompatPruneWorker{pruner: compatPruner})
	river.AddWorker(workers, &CompatCatchupWorker{catchup: compatCatchup})

	return workers
}
