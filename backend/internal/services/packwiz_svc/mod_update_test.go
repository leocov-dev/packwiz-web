package packwiz_svc

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/leocov-dev/packwiz-nxt/core"

	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
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
			snap, err := takeModSnapshot(m)
			if err != nil {
				t.Fatal(err)
			}
			tt.mutate(m)
			got, err := snap.changedSince(m)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("changedSince = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModSnapshotNumericTypeNormalized(t *testing.T) {
	// same logical value, different Go numeric types (e.g. TOML vs JSON decode)
	m := &core.Mod{Update: core.ModUpdate{"curseforge": {"file-id": uint32(42), "project-id": uint32(7)}}}
	snap, err := takeModSnapshot(m)
	if err != nil {
		t.Fatal(err)
	}
	m.Update = core.ModUpdate{"curseforge": {"file-id": float64(42), "project-id": float64(7)}}
	changed, err := snap.changedSince(m)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("uint32 vs float64 with equal value must not count as changed")
	}
}

func TestLockPackUpdate(t *testing.T) {
	unlock, err := lockPackUpdate(9001)
	if err != nil {
		t.Fatal(err)
	}

	_, err2 := lockPackUpdate(9001)
	var he *response.HttpError
	if !errors.As(err2, &he) || he.Code != http.StatusConflict {
		t.Fatalf("second lock: want 409 HttpError, got %v", err2)
	}

	other, err := lockPackUpdate(9002)
	if err != nil {
		t.Fatalf("different pack must not be blocked: %v", err)
	}
	other()

	unlock()
	again, err := lockPackUpdate(9001)
	if err != nil {
		t.Fatalf("lock after unlock: %v", err)
	}
	again()
}

func TestUpdateAllSummaryResponse(t *testing.T) {
	t.Run("empty has non-nil slices", func(t *testing.T) {
		b, err := json.Marshal((&updateAllSummary{}).response())
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != `{"updated":[],"skipped":[],"failed":[],"upToDate":0,"notChecked":0}` {
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
		s.addNotChecked()

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
		if r.UpToDate != 2 || r.NotChecked != 1 {
			t.Fatalf("upToDate = %d notChecked = %d", r.UpToDate, r.NotChecked)
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

// Version is display-only: a refreshed version alone must not count as an update.
func TestModSnapshotIgnoresVersion(t *testing.T) {
	m := &core.Mod{FileName: "a.jar", Update: core.ModUpdate{"modrinth": {"version": "v1"}}}
	snap, err := takeModSnapshot(m)
	if err != nil {
		t.Fatal(err)
	}
	m.Version = "1.0.0"
	changed, err := snap.changedSince(m)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("version-only change must not count as changed")
	}
}

func TestNextStoredVersion(t *testing.T) {
	tests := []struct {
		name        string
		old, new    string
		fileChanged bool
		wantVersion string
		wantWrite   bool
	}{
		{"new wins, file changed", "1", "2", true, "2", true},
		{"new wins, file same", "1", "2", false, "2", true},
		{"no new, file changed clears", "1", "", true, "", true},
		{"no new, file same keeps", "1", "", false, "1", false},
		{"nothing stored, file changed", "", "", true, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, w := nextStoredVersion(tt.old, tt.new, tt.fileChanged)
			if v != tt.wantVersion || w != tt.wantWrite {
				t.Errorf("got (%q,%v) want (%q,%v)", v, w, tt.wantVersion, tt.wantWrite)
			}
		})
	}
}
