package compat_svc

import (
	"time"

	"packwiz-web/internal/tables"
)

// Status is the compatibility badge state of a pack, as used by the API and UI.
type Status string

const (
	StatusUnknown  Status = "unknown"  // no run recorded
	StatusQueued   Status = "queued"   // waiting for the worker
	StatusScanning Status = "scanning" // worker is running
	StatusStale    Status = "stale"    // done, but the pack changed since
	StatusOK       Status = "ok"       // no errors or warnings
	StatusWarnings Status = "warnings" // warnings, no errors
	StatusErrors   Status = "errors"   // at least one error
	StatusFailed   Status = "failed"   // scan failed or went stale mid-run
)

// runStaleAfter is how long a queued/running run may stay unfinished before it
// is reported as failed (job timeout 10 min, rescue after 15 min).
const runStaleAfter = 30 * time.Minute

// isActiveRun reports whether run is a live queued/running scan at now.
func isActiveRun(run *tables.PackCompatRun, now time.Time) bool {
	if run == nil {
		return false
	}
	if run.Status != tables.CompatQueued && run.Status != tables.CompatRunning {
		return false
	}
	return now.Sub(run.StartedAt) <= runStaleAfter
}

// DeriveStatus maps a stored run to the badge status. A queued/running run
// older than the stale threshold is failed. A done run whose target hash
// differs from currentTargetHash is stale. Errors outrank warnings outrank ok;
// info findings alone are ok.
func DeriveStatus(run *tables.PackCompatRun, currentTargetHash string, now time.Time) Status {
	if run == nil {
		return StatusUnknown
	}
	switch run.Status {
	case tables.CompatQueued, tables.CompatRunning:
		if !isActiveRun(run, now) {
			return StatusFailed
		}
		if run.Status == tables.CompatQueued {
			return StatusQueued
		}
		return StatusScanning
	case tables.CompatDone:
		switch {
		case run.TargetHash != currentTargetHash:
			return StatusStale
		case run.ErrorCount > 0:
			return StatusErrors
		case run.WarningCount > 0:
			return StatusWarnings
		}
		return StatusOK
	}
	return StatusFailed
}
