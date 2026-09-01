package jobs

import (
	"github.com/riverqueue/river"
)

// NewWorkers builds the set of registered River workers. resolver implements
// the job-facing logic for jobs that need it (see MigrateModsResolver).
// Add new river.AddWorker(workers, &SomeWorker{...}) calls here as real jobs
// are added.
func NewWorkers(resolver MigrateModsResolver) *river.Workers {
	workers := river.NewWorkers()

	river.AddWorker(workers, &MigrateModsWorker{resolver: resolver})

	return workers
}
