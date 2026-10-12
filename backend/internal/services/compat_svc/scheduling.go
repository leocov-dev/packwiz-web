package compat_svc

import (
	"time"

	"packwiz-web/internal/tables"
)

const (
	// ScanDelay is how far in the future an automatic scan is scheduled, so a
	// burst of edits collapses into one scan (the job is unique per pack).
	ScanDelay = 30 * time.Second
	// ManualScanMinInterval is the minimum time between two scans of a pack,
	// enforced server side for manual requests.
	ManualScanMinInterval = 60 * time.Second
	// orphanQueuedAfter is how long a queued run may wait before the catch-up
	// job checks that a job is really coming for it. Longer than ScanDelay.
	orphanQueuedAfter = 2 * time.Minute
	// catchupBatch is how many never-scanned packs one catch-up run queues.
	catchupBatch = 25
	// catchupStagger spaces the scans queued by one catch-up run.
	catchupStagger = 20 * time.Second
)

// scheduledAt is when an automatic scan requested at now should run.
func scheduledAt(now time.Time) time.Time {
	return now.Add(ScanDelay)
}

// catchupSchedule is when the i-th scan of a catch-up run should start.
func catchupSchedule(now time.Time, i int) time.Time {
	return scheduledAt(now).Add(time.Duration(i) * catchupStagger)
}

// lastRunActivity is when run last started or finished.
func lastRunActivity(run *tables.PackCompatRun) time.Time {
	if run.FinishedAt != nil && run.FinishedAt.After(run.StartedAt) {
		return *run.FinishedAt
	}
	return run.StartedAt
}

// shouldEnqueueManual decides whether a manual scan request may start a job:
// not while a scan is already queued or running, and not within
// ManualScanMinInterval of the last run.
func shouldEnqueueManual(run *tables.PackCompatRun, now time.Time) bool {
	if run == nil {
		return true
	}
	if isActiveRun(run, now) {
		return false
	}
	return now.Sub(lastRunActivity(run)) >= ManualScanMinInterval
}

// catchupNeeded reports whether the catch-up job should (re)queue a scan for a
// pack with this run (nil: never scanned). It mirrors the query in
// EnqueueMissing: a queued run nobody started within orphanQueuedAfter, or a
// running run older than runStaleAfter whose worker must have died.
func catchupNeeded(run *tables.PackCompatRun, now time.Time) bool {
	if run == nil {
		return true
	}
	switch run.Status {
	case tables.CompatQueued:
		return run.StartedAt.Before(now.Add(-orphanQueuedAfter))
	case tables.CompatRunning:
		return run.StartedAt.Before(now.Add(-runStaleAfter))
	}
	return false
}
