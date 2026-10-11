package packwiz_svc

import (
	"testing"
	"time"

	"gorm.io/datatypes"

	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
)

func TestBulkRunActiveAndStatus(t *testing.T) {
	now := time.Now()
	running := &tables.PackBulkAddRun{Status: tables.BulkAddRunning, CreatedAt: now.Add(-time.Minute)}
	if !isActiveBulkRun(running, now) || bulkRunStatus(running, now) != tables.BulkAddRunning {
		t.Error("recent running run must be active")
	}
	stale := &tables.PackBulkAddRun{Status: tables.BulkAddQueued, CreatedAt: now.Add(-time.Hour)}
	if isActiveBulkRun(stale, now) || bulkRunStatus(stale, now) != tables.BulkAddFailed {
		t.Error("stale queued run must be reported failed")
	}
	done := &tables.PackBulkAddRun{Status: tables.BulkAddDone, CreatedAt: now}
	if isActiveBulkRun(done, now) || bulkRunStatus(done, now) != tables.BulkAddDone {
		t.Error("done run is not active")
	}
	if isActiveBulkRun(nil, now) {
		t.Error("nil run is not active")
	}
}

func TestIsReviewableItem(t *testing.T) {
	for _, s := range []string{tables.BulkItemPending, tables.BulkItemIncompatibleVersion, tables.BulkItemConflict} {
		if !isReviewableItem(s) {
			t.Errorf("%s must be reviewable", s)
		}
	}
	for _, s := range []string{tables.BulkItemQueued, tables.BulkItemAdded, tables.BulkItemSkipped, tables.BulkItemFailed, tables.BulkItemPicked, tables.BulkItemRejected} {
		if isReviewableItem(s) {
			t.Errorf("%s must not be reviewable", s)
		}
	}
}

func TestBulkItemUpdate(t *testing.T) {
	cands := []tables.BulkAddCandidate{{Source: sourceModrinth, Slug: "x"}}
	ref := mrRef("sodium", "Sodium")

	pending := bulkItemUpdate(lineMatch{Candidates: cands}, itemOutcome{Status: tables.BulkItemPending, Reason: "no exact match"})
	if got := pending["candidates"].(datatypes.JSONSlice[tables.BulkAddCandidate]); len(got) != 1 {
		t.Errorf("pending line must keep candidates, got %v", got)
	}

	m := matched(1, ref, nil)
	added := bulkItemUpdate(lineMatch{Candidates: cands, Ref: ref}, itemOutcome{Status: tables.BulkItemAdded, Match: m.Match})
	if got := added["candidates"].(datatypes.JSONSlice[tables.BulkAddCandidate]); got == nil || len(got) != 0 {
		t.Errorf("added line must store an empty candidate list, got %v", got)
	}
	if added["slug"] != "sodium" || added["source"] != sourceModrinth || added["name"] != "Sodium" {
		t.Errorf("added line must store the match: %v", added)
	}

	conflict := bulkItemUpdate(lineMatch{Ref: ref}, itemOutcome{Status: tables.BulkItemConflict, ConflictsWith: 3, Reason: "r"})
	if conflict["conflicts_with_position"] != 3 || conflict["reason"] != "r" {
		t.Errorf("conflict = %v", conflict)
	}
}

func TestBuildBulkAddResponse(t *testing.T) {
	now := time.Now()
	run := tables.PackBulkAddRun{ID: 4, Status: tables.BulkAddRunning, CreatedAt: now.Add(-time.Hour)}
	items := []tables.PackBulkAddItem{{ID: 1, Position: 1, Line: "sodium", Status: tables.BulkItemQueued,
		Candidates: datatypes.NewJSONSlice([]tables.BulkAddCandidate{{Source: sourceModrinth, Slug: "sodium"}})}}
	out := buildBulkAddResponse(run, items, now)
	if out.Id != 4 || out.Status != tables.BulkAddFailed {
		t.Errorf("stale run must report failed: %+v", out)
	}
	if len(out.Items) != 1 || len(out.Items[0].Candidates) != 1 || out.Items[0].Candidates[0].Slug != "sodium" {
		t.Errorf("items = %+v", out.Items)
	}
}

func TestResolveBulkItemRequestValidate(t *testing.T) {
	tests := []struct {
		req dto.ResolveBulkItemRequest
		ok  bool
	}{
		{dto.ResolveBulkItemRequest{Action: "reject"}, true},
		{dto.ResolveBulkItemRequest{Action: "pick", Source: "modrinth", Slug: "sodium"}, true},
		{dto.ResolveBulkItemRequest{Action: "pick", Source: "curseforge", ProjectId: "238222"}, true},
		{dto.ResolveBulkItemRequest{Action: "pick", Source: "curseforge", Slug: "jei"}, false},
		{dto.ResolveBulkItemRequest{Action: "pick", Source: "modrinth"}, false},
		{dto.ResolveBulkItemRequest{Action: "pick"}, false},
		{dto.ResolveBulkItemRequest{Action: "delete"}, false},
	}
	for _, tt := range tests {
		if err := tt.req.Validate(); (err == nil) != tt.ok {
			t.Errorf("%+v: err = %v, want ok %v", tt.req, err, tt.ok)
		}
	}
}
