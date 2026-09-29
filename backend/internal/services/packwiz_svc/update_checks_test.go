package packwiz_svc

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/leocov-dev/packwiz-nxt/core"

	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
)

func TestBuildCheckRows(t *testing.T) {
	now := time.Now()
	mods := map[string]tables.Mod{"a": {ID: 1, Slug: "a"}, "b": {ID: 2, Slug: "b"}, "c": {ID: 3, Slug: "c"}}
	results := []core.UpdateCheckResult{
		{Mod: &core.Mod{Slug: "a"}, UpdateAvailable: true, UpdateString: "1 -> 2"},
		{Mod: &core.Mod{Slug: "b"}, UpdateAvailable: true, UpdateString: "x", Err: errors.New("boom")},
		{Mod: &core.Mod{Slug: "gone"}, UpdateAvailable: true},
		{Mod: nil},
	}
	rows := buildCheckRows(9, mods, results, now)
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	if rows[0].ModID != 1 || !rows[0].UpdateAvailable || rows[0].UpdateString != "1 -> 2" || rows[0].PackID != 9 {
		t.Errorf("row0 = %+v", rows[0])
	}
	if rows[1].ModID != 2 || rows[1].UpdateAvailable || rows[1].UpdateString != "" || rows[1].Error != "boom" {
		t.Errorf("failed check must store error only: %+v", rows[1])
	}
}

func TestRunStatusAndFlags(t *testing.T) {
	now := time.Now()
	fin := now.Add(-5 * time.Minute)
	oldFin := now.Add(-11 * time.Minute)

	if runStatus(nil, now) != dto.UpdateCheckIdle {
		t.Error("nil run must be idle")
	}
	running := &tables.PackUpdateCheckRun{Status: tables.UpdateCheckRunning, StartedAt: now.Add(-time.Minute)}
	if !isActiveRun(running, now) || runStatus(running, now) != "running" {
		t.Error("recent running run must be active")
	}
	stale := &tables.PackUpdateCheckRun{Status: tables.UpdateCheckQueued, StartedAt: now.Add(-time.Hour)}
	if isActiveRun(stale, now) || runStatus(stale, now) != "failed" {
		t.Error("stale queued run must be treated as failed")
	}
	if !isFreshDone(&tables.PackUpdateCheckRun{Status: tables.UpdateCheckDone, FinishedAt: &fin}, now) {
		t.Error("5 min old done run is fresh")
	}
	if isFreshDone(&tables.PackUpdateCheckRun{Status: tables.UpdateCheckDone, FinishedAt: &oldFin}, now) {
		t.Error("11 min old done run is not fresh")
	}
	if isFreshDone(&tables.PackUpdateCheckRun{Status: tables.UpdateCheckFailed, FinishedAt: &fin}, now) {
		t.Error("failed run is never fresh")
	}
}

func TestBuildUpdateChecksResponse(t *testing.T) {
	now := time.Now()
	fin := now.Add(-time.Minute)
	run := &tables.PackUpdateCheckRun{Status: tables.UpdateCheckDone, FinishedAt: &fin}
	rows := []checkRow{
		{ModUpdateCheck: tables.ModUpdateCheck{ModID: 1, UpdateAvailable: true, UpdateString: "u"}},
		{ModUpdateCheck: tables.ModUpdateCheck{ModID: 2, UpdateAvailable: true}, Pinned: true},
		{ModUpdateCheck: tables.ModUpdateCheck{ModID: 3, Error: "e"}},
		{ModUpdateCheck: tables.ModUpdateCheck{ModID: 4, CheckedAt: now.Add(-time.Hour)}},
	}
	rows[0].CheckedAt = now.Add(-2 * time.Hour)
	rows[2].CheckedAt = now.Add(-3 * time.Hour)
	got := buildUpdateChecksResponse(run, rows, now)
	if got.Status != "done" || got.AvailableCount != 1 || len(got.Results) != 4 || got.CheckedAt == nil {
		t.Fatalf("unexpected response: %+v", got)
	}

	if !got.CheckedAt.Equal(rows[3].CheckedAt) {
		t.Errorf("checkedAt should be newest row time, got %v", got.CheckedAt)
	}
	if got.RunFinishedAt == nil || !got.RunFinishedAt.Equal(fin) {
		t.Errorf("runFinishedAt = %v", got.RunFinishedAt)
	}

	// a failed run keeps the last successful results' time and reports its own error
	failedFin := now
	failed := buildUpdateChecksResponse(&tables.PackUpdateCheckRun{
		Status: tables.UpdateCheckFailed, FinishedAt: &failedFin, Error: "boom",
	}, rows, now)
	if failed.Status != "failed" || failed.Error != "boom" || !failed.CheckedAt.Equal(rows[3].CheckedAt) ||
		!failed.RunFinishedAt.Equal(failedFin) {
		t.Errorf("failed run response: %+v", failed)
	}

	idle := buildUpdateChecksResponse(nil, nil, now)
	b, _ := json.Marshal(idle)
	if string(b) != `{"status":"idle","checkedAt":null,"runFinishedAt":null,"results":[],"availableCount":0}` {
		t.Errorf("idle JSON shape: %s", b)
	}

	b, _ = json.Marshal(got.Results[2])
	if string(b) != `{"modId":3,"updateAvailable":false,"error":"e"}` {
		t.Errorf("item JSON shape: %s", b)
	}
}

func TestShouldEnqueueCheck(t *testing.T) {
	now := time.Now()
	ago := func(d time.Duration) *time.Time { t := now.Add(-d); return &t }
	done := func(d time.Duration) *tables.PackUpdateCheckRun {
		return &tables.PackUpdateCheckRun{Status: tables.UpdateCheckDone, StartedAt: now.Add(-d - time.Second), FinishedAt: ago(d)}
	}
	cases := []struct {
		name  string
		run   *tables.PackUpdateCheckRun
		force bool
		want  bool
	}{
		{"no run", nil, false, true},
		{"active", &tables.PackUpdateCheckRun{Status: tables.UpdateCheckRunning, StartedAt: now.Add(-time.Minute)}, true, false},
		{"stale queued", &tables.PackUpdateCheckRun{Status: tables.UpdateCheckQueued, StartedAt: now.Add(-time.Hour)}, false, true},
		{"fresh done", done(5 * time.Minute), false, false},
		{"fresh done forced", done(5 * time.Minute), true, true},
		{"force within min interval", done(10 * time.Second), true, false},
		{"force at min interval", done(updateCheckMinInterval), true, true},
		{"old done", done(20 * time.Minute), false, true},
		{"failed within min interval", &tables.PackUpdateCheckRun{Status: tables.UpdateCheckFailed, StartedAt: now.Add(-20 * time.Second), FinishedAt: ago(5 * time.Second)}, true, false},
		{"failed long ago", &tables.PackUpdateCheckRun{Status: tables.UpdateCheckFailed, StartedAt: now.Add(-time.Hour), FinishedAt: ago(time.Hour)}, false, true},
	}
	for _, c := range cases {
		if got := shouldEnqueueCheck(c.run, now, c.force); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

func TestFilterFreshCheckRows(t *testing.T) {
	started := time.Now()
	rows := []tables.ModUpdateCheck{{ModID: 1}, {ModID: 2}, {ModID: 3}}
	mods := map[uint]time.Time{
		1: started.Add(-time.Hour),  // untouched: kept
		2: started.Add(time.Second), // updated during the check: dropped
		// 3 removed: dropped
	}
	old := started.Add(-time.Hour)
	got := filterFreshCheckRows(rows, mods, old, started)
	if len(got) != 1 || got[0].ModID != 1 {
		t.Fatalf("got %+v", got)
	}
	if got := filterFreshCheckRows(rows, mods, started.Add(time.Second), started); len(got) != 0 {
		t.Errorf("pack changed during check should drop all rows, got %+v", got)
	}
}
