package packwiz_svc

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/leocov-dev/packwiz-nxt/core"

	"packwiz-web/internal/types/dto"
)

func TestModSnapshotChangedSince(t *testing.T) {
	newMod := func() *core.Mod {
		return &core.Mod{
			FileName: "a-1.0.jar",
			Download: core.ModDownload{URL: "u", Hash: "h1", HashFormat: "sha1"},
			Update:   core.ModUpdate{"modrinth": {"mod-id": "abc", "version": "v1"}},
		}
	}

	tests := []struct {
		name   string
		mutate func(m *core.Mod)
		want   bool
	}{
		{"untouched", func(m *core.Mod) {}, false},
		{"file name", func(m *core.Mod) { m.FileName = "a-2.0.jar" }, true},
		{"hash", func(m *core.Mod) { m.Download.Hash = "h2" }, true},
		{"url", func(m *core.Mod) { m.Download.URL = "u2" }, true},
		{"update map mutated in place", func(m *core.Mod) { m.Update["modrinth"]["version"] = "v2" }, true},
		{"same values rewritten", func(m *core.Mod) { m.FileName = "a-1.0.jar"; m.Update["modrinth"]["version"] = "v1" }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMod()
			snap := takeModSnapshot(m)
			tt.mutate(m)
			if got := snap.changedSince(m); got != tt.want {
				t.Fatalf("changedSince = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateAllSummaryResponse(t *testing.T) {
	t.Run("empty has non-nil slices", func(t *testing.T) {
		b, err := json.Marshal((&updateAllSummary{}).response())
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != `{"updated":[],"skipped":[],"failed":[],"upToDate":0}` {
			t.Fatalf("unexpected json: %s", b)
		}
	})

	t.Run("collects sections sorted by slug", func(t *testing.T) {
		s := &updateAllSummary{}
		s.addUpdated(dto.UpdateAllItem{ModId: 2, Slug: "zed", Name: "Zed", FileName: "zed-2.jar"})
		s.addUpdated(dto.UpdateAllItem{ModId: 1, Slug: "alpha", Name: "Alpha", FileName: "alpha-2.jar"})
		s.addSkipped(dto.UpdateAllItem{ModId: 3, Slug: "pin", Name: "Pin"}, SkipReasonPinned)
		s.addFailed(dto.UpdateAllItem{ModId: 4, Slug: "bad", Name: "Bad"}, errors.New("boom"))
		s.addUpToDate()
		s.addUpToDate()

		r := s.response()
		if len(r.Updated) != 2 || r.Updated[0].Slug != "alpha" || r.Updated[1].Slug != "zed" {
			t.Fatalf("updated not sorted: %+v", r.Updated)
		}
		if len(r.Skipped) != 1 || r.Skipped[0].Reason != "pinned" {
			t.Fatalf("skipped: %+v", r.Skipped)
		}
		if len(r.Failed) != 1 || r.Failed[0].Error != "boom" {
			t.Fatalf("failed: %+v", r.Failed)
		}
		if r.UpToDate != 2 {
			t.Fatalf("upToDate = %d", r.UpToDate)
		}
	})
}

func TestUpdateResponseJSONShape(t *testing.T) {
	b, _ := json.Marshal(dto.UpdateModResponse{Updated: true})
	if string(b) != `{"updated":true}` {
		t.Fatalf("unexpected json: %s", b)
	}

	b, _ = json.Marshal(dto.UpdateAllItem{ModId: 1, Slug: "s", Name: "N", FileName: "f.jar"})
	if string(b) != `{"modId":1,"slug":"s","name":"N","fileName":"f.jar"}` {
		t.Fatalf("unexpected json: %s", b)
	}
}
