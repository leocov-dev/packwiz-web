package packwiz_svc

import (
	"encoding/json"
	"reflect"
	"testing"

	"codeberg.org/jmansfield/go-modrinth/modrinth"

	"packwiz-web/internal/types/dto"
)

func TestSplitModrinthCategories(t *testing.T) {
	tests := []struct {
		name           string
		in             []string
		wantLoaders    []string
		wantCategories []string
	}{
		{"nil", nil, nil, nil},
		{"mixed", []string{"fabric", "adventure", "Forge", "magic"}, []string{"fabric", "forge"}, []string{"adventure", "magic"}},
		{"dupes and blanks", []string{"fabric", "fabric", " ", "utility"}, []string{"fabric"}, []string{"utility"}},
		{"only categories", []string{"technology"}, nil, []string{"technology"}},
		{"newer loaders", []string{"babric", "bta-babric", "legacy-fabric", "ornithe", "nilloader", "java-agent", "geyser", "leaf", "magic"}, []string{"babric", "bta-babric", "legacy-fabric", "ornithe", "nilloader", "java-agent", "geyser", "leaf"}, []string{"magic"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, c := splitModrinthCategories(tt.in)
			if !reflect.DeepEqual(l, tt.wantLoaders) || !reflect.DeepEqual(c, tt.wantCategories) {
				t.Fatalf("got %v %v, want %v %v", l, c, tt.wantLoaders, tt.wantCategories)
			}
		})
	}
}

func TestDownloadsPtr(t *testing.T) {
	if downloadsPtr(nil) != 0 {
		t.Fatal("nil should be 0")
	}
	v := uint32(42)
	if downloadsPtr(&v) != 42 {
		t.Fatal("want 42")
	}
}

func TestCurseforgeSearchItemMapping(t *testing.T) {
	const body = `{"data":[{"id":1,"name":"X","slug":"x","summary":"s",
		"authors":[{"name":""},{"name":"alice"},{"name":"bob"}],
		"downloadCount":1234567.0,
		"latestFilesIndexes":[{"modLoader":4},{"modLoader":1},{"modLoader":4},{"modLoader":2},{"modLoader":6},{"modLoader":99},{}],
		"categories":[{"name":"Magic"},{"name":"Magic"},{"name":"Utility"}]}]}`
	var res cfSearchResponse
	if err := json.Unmarshal([]byte(body), &res); err != nil {
		t.Fatal(err)
	}
	item := res.Data[0]
	if got := cfAuthor(item); got != "alice" {
		t.Errorf("author %q", got)
	}
	if got := cfDownloads(item); got != 1234567 {
		t.Errorf("downloads %d", got)
	}
	if got, want := cfLoaders(item), []string{"fabric", "forge", "neoforge"}; !reflect.DeepEqual(got, want) {
		t.Errorf("loaders %v want %v", got, want)
	}
	if got, want := cfCategories(item), []string{"Magic", "Utility"}; !reflect.DeepEqual(got, want) {
		t.Errorf("categories %v want %v", got, want)
	}
}

func TestCurseforgeSearchItemEmpty(t *testing.T) {
	var item cfSearchItem
	if cfAuthor(item) != "" || cfDownloads(item) != 0 || cfLoaders(item) != nil || cfCategories(item) != nil {
		t.Fatal("empty item should yield zero values")
	}
	item.DownloadCount = "-5"
	if cfDownloads(item) != 0 {
		t.Fatal("negative downloads should clamp to 0")
	}
}

func TestCfDownloadsTolerantNumbers(t *testing.T) {
	for in, want := range map[string]uint64{"1234567": 1234567, "1.0e6": 1000000, "2500.0": 2500, "": 0, "abc": 0} {
		if got := cfDownloads(cfSearchItem{DownloadCount: json.Number(in)}); got != want {
			t.Errorf("%q: got %d want %d", in, got, want)
		}
	}
}

func TestModSearchResultJSON(t *testing.T) {
	full, err := json.Marshal(dto.ModSearchResult{
		Slug: "s", Title: "t", Description: "d", IconUrl: "i", ProjectId: "p", Installed: true,
		Author: "a", Downloads: 5, Loaders: []string{"fabric"}, Categories: []string{"magic"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(full, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"slug", "title", "description", "iconUrl", "projectId", "installed", "author", "downloads", "loaders", "categories"} {
		if _, ok := m[k]; !ok {
			t.Errorf("missing key %q in %s", k, full)
		}
	}

	empty, _ := json.Marshal(dto.ModSearchResult{})
	m = nil
	_ = json.Unmarshal(empty, &m)
	for _, k := range []string{"author", "downloads", "loaders", "categories"} {
		if _, ok := m[k]; ok {
			t.Errorf("key %q should be omitted when empty", k)
		}
	}
	for _, k := range []string{"slug", "title", "description", "iconUrl", "projectId", "installed"} {
		if _, ok := m[k]; !ok {
			t.Errorf("key %q should always be present", k)
		}
	}
}

func TestModrinthResultFromHit(t *testing.T) {
	s := func(v string) *string { return &v }
	d := uint32(99)
	hit := &modrinth.SearchResult{
		Slug: s("sl"), ProjectID: s("pid"), Title: s("T"), Description: s("D"), IconURL: s("ic"),
		Author: s("alice"), Downloads: &d, Categories: []string{"fabric", "game-mechanics"},
	}
	got := modrinthResultFromHit(hit, true)
	want := dto.ModSearchResult{
		Slug: "sl", ProjectId: "pid", Title: "T", Description: "D", IconUrl: "ic", Installed: true,
		Author: "alice", Downloads: 99, Loaders: []string{"fabric"}, Categories: []string{"game-mechanics"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestCurseforgeResultFromItem(t *testing.T) {
	const body = `{"data":[{"id":1,"name":"X","slug":"x","summary":"s","authors":[{"name":"bob"}],
		"downloadCount":10,"latestFilesIndexes":[{"modLoader":5}],"categories":[{"name":"Magic"}]}]}`
	var res cfSearchResponse
	if err := json.Unmarshal([]byte(body), &res); err != nil {
		t.Fatal(err)
	}
	got := curseforgeResultFromItem(res.Data[0], "1", "icon", false)
	want := dto.ModSearchResult{
		Slug: "x", ProjectId: "1", Title: "X", Description: "s", IconUrl: "icon",
		Author: "bob", Downloads: 10, Loaders: []string{"quilt"}, Categories: []string{"Magic"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v want %+v", got, want)
	}
}
