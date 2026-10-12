package jobs

import (
	"testing"

	"github.com/riverqueue/river/rivertype"
)

func TestCompatScanInsertOpts(t *testing.T) {
	opts := CompatScanArgs{}.InsertOpts()

	if opts.Queue != QueueCompatScan {
		t.Errorf("queue = %q", opts.Queue)
	}
	if opts.MaxAttempts != 1 {
		t.Errorf("max attempts = %d, want 1", opts.MaxAttempts)
	}
	if !opts.UniqueOpts.ByArgs {
		t.Error("scan must be unique by args (pack id)")
	}

	states := map[rivertype.JobState]bool{}
	for _, s := range opts.UniqueOpts.ByState {
		states[s] = true
	}
	for _, s := range []rivertype.JobState{
		rivertype.JobStateAvailable, rivertype.JobStatePending, rivertype.JobStateRunning,
		rivertype.JobStateRetryable, rivertype.JobStateScheduled,
	} {
		if !states[s] {
			t.Errorf("unique states miss %s", s)
		}
	}
	// completed scans must never block the next one
	if states[rivertype.JobStateCompleted] {
		t.Error("unique states include completed")
	}
}

func TestCompatScanTimeoutBelowRescue(t *testing.T) {
	if CompatScanTimeout >= 15*60*1e9 {
		t.Errorf("scan timeout %v must stay below RescueStuckJobsAfter (15m)", CompatScanTimeout)
	}
}

func TestCompatArgsKinds(t *testing.T) {
	kinds := map[string]bool{}
	for _, k := range []string{CompatScanArgs{}.Kind(), CompatPruneArgs{}.Kind(), CompatCatchupArgs{}.Kind()} {
		if k == "" || kinds[k] {
			t.Errorf("kind %q empty or duplicated", k)
		}
		kinds[k] = true
	}
}
