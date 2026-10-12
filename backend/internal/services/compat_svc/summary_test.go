package compat_svc

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
)

func TestSummarize(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	pack := tables.Pack{MCVersion: "1.21.1", Loader: "fabric", LoaderVersion: "0.16.0"}
	mods := []tables.Mod{mod(1, "sodium", "mods", "AA")}
	h := TargetHash(targetInputOf(pack, mods))
	tests := []struct {
		name string
		run  tables.PackCompatRun
		want dto.CompatSummary
	}{
		{"errors", tables.PackCompatRun{Status: tables.CompatDone, TargetHash: h, ErrorCount: 2, WarningCount: 1}, dto.CompatSummary{Status: "errors", Errors: 2, Warnings: 1}},
		{"warnings", tables.PackCompatRun{Status: tables.CompatDone, TargetHash: h, WarningCount: 4}, dto.CompatSummary{Status: "warnings", Warnings: 4}},
		{"ok", tables.PackCompatRun{Status: tables.CompatDone, TargetHash: h}, dto.CompatSummary{Status: "ok"}},
		{"stale keeps counts", tables.PackCompatRun{Status: tables.CompatDone, TargetHash: "old", ErrorCount: 1}, dto.CompatSummary{Status: "stale", Errors: 1}},
		{"queued", tables.PackCompatRun{Status: tables.CompatQueued, StartedAt: now}, dto.CompatSummary{Status: "queued"}},
		{"failed", tables.PackCompatRun{Status: tables.CompatFailed}, dto.CompatSummary{Status: "failed"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := summarize(&tt.run, pack, mods, now); got != tt.want {
				t.Errorf("got %+v want %+v", got, tt.want)
			}
		})
	}
}

// The list summary (which loads only jar mods) and GetCompat / the manual scan
// response (which load all mods) must derive the same status.
func TestStatusAgreesBetweenListAndDetail(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	pack := tables.Pack{MCVersion: "1.21.1", Loader: "fabric", LoaderVersion: "0.16.0"}
	all := []tables.Mod{
		mod(1, "sodium", "mods", "AA"),
		mod(2, "faithful", "resourcepacks", "BB"),
		mod(3, "complementary", "shaderpacks", "CC"),
	}
	run := tables.PackCompatRun{Status: tables.CompatDone, TargetHash: TargetHash(targetInputOf(pack, all))}

	check := func(t *testing.T, all []tables.Mod, want Status) {
		t.Helper()
		detail := statusFor(&run, pack, all, now)
		list := Status(summarize(&run, pack, jarMods(all), now).Status)
		if detail != want || list != want {
			t.Errorf("detail=%s list=%s want %s", detail, list, want)
		}
	}
	t.Run("unchanged is ok", func(t *testing.T) { check(t, all, StatusOK) })

	t.Run("non-jar change is not stale", func(t *testing.T) {
		changed := append([]tables.Mod(nil), all...)
		changed[1].Download.Hash = "NEW"
		changed = append(changed, mod(4, "extra-shader", "shaderpacks", "DD"))
		check(t, changed, StatusOK)
	})

	t.Run("jar change is stale", func(t *testing.T) {
		changed := append([]tables.Mod(nil), all...)
		changed[0].Download.Hash = "NEW"
		check(t, changed, StatusStale)
	})

	t.Run("added jar is stale", func(t *testing.T) {
		check(t, append(append([]tables.Mod(nil), all...), mod(5, "lithium", "mods", "EE")), StatusStale)
	})
}

func TestFailureReason(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"deadline", fmt.Errorf("scan interrupted: %w", context.DeadlineExceeded), "scan timed out"},
		{"cancelled", context.Canceled, "scan timed out"},
		{"database", errors.New(`ERROR: duplicate key value violates unique constraint "pack_compat_findings_pkey" (SQLSTATE 23505)`), "scan failed"},
		{"archived", errors.New("pack is archived or deleted"), "scan failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := failureReason(tt.err); got != tt.want {
				t.Errorf("got %q want %q", got, tt.want)
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
