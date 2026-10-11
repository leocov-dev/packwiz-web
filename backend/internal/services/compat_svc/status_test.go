package compat_svc

import (
	"testing"
	"time"

	"packwiz-web/internal/tables"
)

func TestDeriveStatus(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-time.Minute)
	old := now.Add(-2 * runStaleAfter)

	done := func(target string, e, w, i int) *tables.PackCompatRun {
		return &tables.PackCompatRun{Status: tables.CompatDone, TargetHash: target,
			ErrorCount: e, WarningCount: w, InfoCount: i, StartedAt: fresh}
	}

	tests := []struct {
		name string
		run  *tables.PackCompatRun
		want Status
	}{
		{"no run", nil, StatusUnknown},
		{"queued", &tables.PackCompatRun{Status: tables.CompatQueued, StartedAt: fresh}, StatusQueued},
		{"running", &tables.PackCompatRun{Status: tables.CompatRunning, StartedAt: fresh}, StatusScanning},
		{"queued stale", &tables.PackCompatRun{Status: tables.CompatQueued, StartedAt: old}, StatusFailed},
		{"running stale", &tables.PackCompatRun{Status: tables.CompatRunning, StartedAt: old}, StatusFailed},
		{"failed", &tables.PackCompatRun{Status: tables.CompatFailed, TargetHash: "t"}, StatusFailed},
		{"unknown status", &tables.PackCompatRun{Status: "weird"}, StatusFailed},
		{"done clean", done("t", 0, 0, 0), StatusOK},
		{"done info only", done("t", 0, 0, 5), StatusOK},
		{"done warnings", done("t", 0, 2, 5), StatusWarnings},
		{"done errors", done("t", 1, 2, 5), StatusErrors},
		{"done stale clean", done("other", 0, 0, 0), StatusStale},
		{"done stale errors", done("other", 3, 0, 0), StatusStale},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeriveStatus(tt.run, "t", now); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
