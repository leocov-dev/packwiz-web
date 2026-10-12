package compat_svc

import (
	"testing"
	"time"

	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
)

func TestSummarize(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		run  tables.PackCompatRun
		hash string
		want dto.CompatSummary
	}{
		{"errors", tables.PackCompatRun{Status: tables.CompatDone, TargetHash: "h", ErrorCount: 2, WarningCount: 1}, "h", dto.CompatSummary{Status: "errors", Errors: 2, Warnings: 1}},
		{"warnings", tables.PackCompatRun{Status: tables.CompatDone, TargetHash: "h", WarningCount: 4}, "h", dto.CompatSummary{Status: "warnings", Warnings: 4}},
		{"ok", tables.PackCompatRun{Status: tables.CompatDone, TargetHash: "h"}, "h", dto.CompatSummary{Status: "ok"}},
		{"stale keeps counts", tables.PackCompatRun{Status: tables.CompatDone, TargetHash: "old", ErrorCount: 1}, "new", dto.CompatSummary{Status: "stale", Errors: 1}},
		{"queued", tables.PackCompatRun{Status: tables.CompatQueued, StartedAt: now}, "", dto.CompatSummary{Status: "queued"}},
		{"failed", tables.PackCompatRun{Status: tables.CompatFailed}, "", dto.CompatSummary{Status: "failed"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := summarize(&tt.run, tt.hash, now); got != tt.want {
				t.Errorf("got %+v want %+v", got, tt.want)
			}
		})
	}
}

func TestSummariesForNilServiceAndEmpty(t *testing.T) {
	var nilSvc *CompatService
	if _, err := nilSvc.SummariesForPacks(t.Context(), []uint{1}); err == nil {
		t.Error("nil service must error, not panic")
	}
	s := &CompatService{}
	got, err := s.SummariesForPacks(t.Context(), nil)
	if err != nil || len(got) != 0 {
		t.Errorf("empty: %v %v", got, err)
	}
}
