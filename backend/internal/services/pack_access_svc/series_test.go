package pack_access_svc

import (
	"testing"
	"time"

	"packwiz-web/internal/types/dto"
)

func TestFillDays(t *testing.T) {
	now := time.Date(2026, 10, 3, 15, 30, 0, 0, time.UTC)
	sparse := []dto.AccessDay{
		{Date: "2026-10-03", Success: 4, Failure: 1},
		{Date: "2026-10-01", Success: 2},
		{Date: "2026-09-01", Success: 99}, // outside window, dropped
	}

	got := FillDays(sparse, now, 3)
	want := []dto.AccessDay{
		{Date: "2026-10-01", Success: 2},
		{Date: "2026-10-02"},
		{Date: "2026-10-03", Success: 4, Failure: 1},
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("day %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestFillDaysEmpty(t *testing.T) {
	got := FillDays(nil, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 7)
	if len(got) != 7 || got[0].Date != "2025-12-26" || got[6].Date != "2026-01-01" {
		t.Errorf("unexpected window: %+v", got)
	}
}

func TestRangeStartUsesUTC(t *testing.T) {
	loc := time.FixedZone("x", -8*3600)
	now := time.Date(2026, 10, 3, 20, 0, 0, 0, loc) // 04:00 UTC on the 4th
	if got := rangeStart(now, 1).Format(dateLayout); got != "2026-10-04" {
		t.Errorf("rangeStart = %s, want 2026-10-04", got)
	}
}
