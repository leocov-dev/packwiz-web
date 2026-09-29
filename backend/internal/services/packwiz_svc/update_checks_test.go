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
		{ModUpdateCheck: tables.ModUpdateCheck{ModID: 4}},
	}
	got := buildUpdateChecksResponse(run, rows, now)
	if got.Status != "done" || got.AvailableCount != 1 || len(got.Results) != 4 || got.CheckedAt == nil {
		t.Fatalf("unexpected response: %+v", got)
	}

	idle := buildUpdateChecksResponse(nil, nil, now)
	b, _ := json.Marshal(idle)
	if string(b) != `{"status":"idle","checkedAt":null,"results":[],"availableCount":0}` {
		t.Errorf("idle JSON shape: %s", b)
	}

	b, _ = json.Marshal(got.Results[2])
	if string(b) != `{"modId":3,"updateAvailable":false,"error":"e"}` {
		t.Errorf("item JSON shape: %s", b)
	}
}
