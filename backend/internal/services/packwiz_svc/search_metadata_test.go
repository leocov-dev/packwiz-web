package packwiz_svc

import (
	"encoding/json"
	"reflect"
	"testing"
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
	item.DownloadCount = -5
	if cfDownloads(item) != 0 {
		t.Fatal("negative downloads should clamp to 0")
	}
}
