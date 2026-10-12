package dto

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"gorm.io/datatypes"

	"packwiz-web/internal/tables"
)

func TestNewCompatResponseNeverScanned(t *testing.T) {
	got := NewCompatResponse("unknown", nil, nil)
	if got.Status != "unknown" || got.Findings == nil || len(got.Findings) != 0 || got.ScannedAt != nil {
		t.Fatalf("unexpected: %+v", got)
	}
	b, _ := json.Marshal(got)
	if !strings.Contains(string(b), `"findings":[]`) || !strings.Contains(string(b), `"scannedAt":null`) {
		t.Errorf("json shape: %s", b)
	}
}

func TestNewCompatResponseMapsRun(t *testing.T) {
	fin := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	run := &tables.PackCompatRun{ErrorCount: 2, WarningCount: 3, InfoCount: 4, UnscannedCount: 5, FinishedAt: &fin, Error: "x"}
	got := NewCompatResponse("errors", run, nil)
	if got.Counts != (CompatCounts{Errors: 2, Warnings: 3, Info: 4}) || got.Unscanned != 5 || got.RunError != "x" || got.ScannedAt != &fin {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestNewCompatResponseOrdersAndMaps(t *testing.T) {
	id := uint(7)
	findings := []tables.PackCompatFinding{
		{Slug: "b", Severity: "info", Code: "a"},
		{Slug: "z", Severity: "warning", Code: "a"},
		{Slug: "b", Severity: "error", Code: "z"},
		{Slug: "b", Severity: "error", Code: "a", ModID: &id, Message: "m", Detail: datatypes.JSONMap{"s": "v", "n": 3}},
		{Slug: "a", Severity: "warning", Code: "b"},
		{Slug: "a", Severity: "warning", Code: "a"},
	}
	got := NewCompatResponse("errors", nil, findings)

	var order []string
	for _, f := range got.Findings {
		order = append(order, f.Severity+"/"+f.Slug+"/"+f.Code)
	}
	want := []string{"error/b/a", "error/b/z", "warning/a/a", "warning/a/b", "warning/z/a", "info/b/a"}
	if !reflect.DeepEqual(order, want) {
		t.Errorf("order:\n got %v\nwant %v", order, want)
	}

	first := got.Findings[0]
	if first.ModId == nil || *first.ModId != 7 || first.Message != "m" {
		t.Errorf("mapping: %+v", first)
	}
	if !reflect.DeepEqual(first.Detail, map[string]string{"s": "v", "n": "3"}) {
		t.Errorf("detail: %v", first.Detail)
	}
	if got.Findings[1].Detail == nil {
		t.Error("detail must not be nil")
	}
}
