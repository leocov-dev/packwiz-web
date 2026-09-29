package jobs

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"gorm.io/gorm"

	"packwiz-web/internal/config"
)

// NewClient builds a River client in poll-only mode, sharing the given
// GORM connection's underlying *sql.DB pool. No LISTEN/NOTIFY pool is used.
func NewClient(gdb *gorm.DB, workers *river.Workers) (*river.Client[*sql.Tx], error) {
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	driver := riverdatabasesql.New(sqlDB)

	client, err := river.NewClient(driver, &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: config.C.JobWorkerPoolSize},
			// migrate_mods is capped at 1 worker until the packwiz-nxt source
			// Updaters (sources/*.go) are verified safe for concurrent use
			// across simultaneous migrations.
			QueueMigrateMods: {MaxWorkers: 1},
			// update_checks: same 1-worker cap, plus third-party API rate limits.
			// NOTE: MaxWorkers is per process. The compose stack runs one worker
			// container (web without --worker), so effective concurrency is 1;
			// running `start --worker` as well would make it 2.
			QueueUpdateChecks: {MaxWorkers: 1},
		},
		// A crashed worker leaves its job "running" (which the update-check
		// uniqueness treats as active) until the rescuer picks it up; River's
		// default is 1h. Must be greater than the longest per-job timeout
		// (jobs.CheckUpdatesTimeout, 10m; other jobs use River's 1m default).
		RescueStuckJobsAfter: 15 * time.Minute,
		Workers:              workers,
		Logger:               newSlogLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create river client: %w", err)
	}

	return client, nil
}
