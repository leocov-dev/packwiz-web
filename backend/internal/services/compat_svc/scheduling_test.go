package compat_svc

import (
	"testing"
	"time"

	"packwiz-web/internal/tables"
)

func TestScheduledAt(t *testing.T) {
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	if got := scheduledAt(now); got != now.Add(ScanDelay) {
		t.Errorf("scheduledAt = %v", got)
	}
	// burst of triggers within the delay: every scheduled time is later than
	// now, so the single pending job has not started when they arrive
	for i := 0; i < 3; i++ {
		if !scheduledAt(now.Add(time.Duration(i) * time.Second)).After(now.Add(time.Duration(i) * time.Second)) {
			t.Error("scheduled time must be in the future")
		}
	}
}

func TestCatchupSchedule(t *testing.T) {
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	prev := catchupSchedule(now, 0)
	if prev != now.Add(ScanDelay) {
		t.Errorf("first = %v", prev)
	}
	for i := 1; i < catchupBatch; i++ {
		next := catchupSchedule(now, i)
		if next.Sub(prev) != catchupStagger {
			t.Fatalf("gap at %d = %v, want %v", i, next.Sub(prev), catchupStagger)
		}
		prev = next
	}
	// the last queued scan still starts well inside the stale threshold
	if last := catchupSchedule(now, catchupBatch-1).Sub(now); last >= runStaleAfter {
		t.Errorf("last catch-up scan starts after %v, run would read as failed", last)
	}
}

func TestShouldEnqueueManual(t *testing.T) {
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	at := func(d time.Duration) *time.Time { t := now.Add(d); return &t }

	tests := []struct {
		name string
		run  *tables.PackCompatRun
		want bool
	}{
		{"never scanned", nil, true},
		{"queued", &tables.PackCompatRun{Status: tables.CompatQueued, StartedAt: now.Add(-time.Second)}, false},
		{"running", &tables.PackCompatRun{Status: tables.CompatRunning, StartedAt: now.Add(-5 * time.Minute)}, false},
		{"dead queued run", &tables.PackCompatRun{Status: tables.CompatQueued, StartedAt: now.Add(-time.Hour)}, true},
		{"finished just now", &tables.PackCompatRun{Status: tables.CompatDone, StartedAt: now.Add(-time.Minute), FinishedAt: at(-10 * time.Second)}, false},
		{"finished a while ago", &tables.PackCompatRun{Status: tables.CompatDone, StartedAt: now.Add(-time.Hour), FinishedAt: at(-time.Hour + time.Minute)}, true},
		{"failed just now", &tables.PackCompatRun{Status: tables.CompatFailed, StartedAt: now.Add(-time.Minute), FinishedAt: at(-5 * time.Second)}, false},
		{"exactly the interval", &tables.PackCompatRun{Status: tables.CompatDone, StartedAt: now.Add(-time.Hour), FinishedAt: at(-ManualScanMinInterval)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldEnqueueManual(tt.run, now); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
