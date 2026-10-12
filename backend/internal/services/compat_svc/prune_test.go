package compat_svc

import (
	"slices"
	"testing"
	"time"

	"github.com/leocov-dev/packwiz-nxt/jarmeta"
)

func metaFor(loader jarmeta.Loader, mcRanges ...string) jarmeta.LoaderMeta {
	m := jarmeta.LoaderMeta{Loader: loader, ModID: "m"}
	if mcRanges != nil {
		m.Dependencies = []jarmeta.Dependency{{
			ID: jarmeta.IDMinecraft, Kind: jarmeta.KindDepends, VersionRanges: mcRanges,
		}}
	}
	return m
}

func TestPruneCandidates(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	oldAt := now.Add(-PruneAfter - time.Hour)
	recentAt := now.Add(-PruneAfter + time.Hour)

	key := JarKey{"sha1", "k"}
	fabric121 := PackTarget{"1.21.1", "fabric"}

	tests := []struct {
		name       string
		row        PruneRow
		referenced bool
		targets    []PackTarget
		want       bool
	}{
		{"old, unreferenced, no targets", PruneRow{key, oldAt, []string{"fabric"}, []jarmeta.LoaderMeta{metaFor("fabric", "~1.20")}}, false, nil, true},
		{"recently used", PruneRow{key, recentAt, []string{"fabric"}, []jarmeta.LoaderMeta{metaFor("fabric", "~1.20")}}, false, nil, false},
		{"referenced", PruneRow{key, oldAt, []string{"fabric"}, []jarmeta.LoaderMeta{metaFor("fabric", "~1.20")}}, true, nil, false},
		{"loader and mc match", PruneRow{key, oldAt, []string{"fabric"}, []jarmeta.LoaderMeta{metaFor("fabric", "~1.21")}}, false, []PackTarget{fabric121}, false},
		{"loader match, mc mismatch", PruneRow{key, oldAt, []string{"fabric"}, []jarmeta.LoaderMeta{metaFor("fabric", "~1.20")}}, false, []PackTarget{fabric121}, true},
		{"mc mismatch for first range, match second", PruneRow{key, oldAt, []string{"fabric"}, []jarmeta.LoaderMeta{metaFor("fabric", "~1.20", "~1.21")}}, false, []PackTarget{fabric121}, false},
		{"loader mismatch", PruneRow{key, oldAt, []string{"forge"}, []jarmeta.LoaderMeta{metaFor("forge", "[1.21.1]")}}, false, []PackTarget{fabric121}, true},
		{"no declared mc range keeps", PruneRow{key, oldAt, []string{"fabric"}, []jarmeta.LoaderMeta{metaFor("fabric")}}, false, []PackTarget{fabric121}, false},
		{"unevaluable range keeps", PruneRow{key, oldAt, []string{"fabric"}, []jarmeta.LoaderMeta{metaFor("fabric", "^^bogus^^")}}, false, []PackTarget{fabric121}, false},
		{"maven range match", PruneRow{key, oldAt, []string{"neoforge"}, []jarmeta.LoaderMeta{metaFor("neoforge", "[1.21,1.22)")}}, false, []PackTarget{{"1.21.1", "neoforge"}}, false},
		{"maven range mismatch", PruneRow{key, oldAt, []string{"neoforge"}, []jarmeta.LoaderMeta{metaFor("neoforge", "[1.20,1.21)")}}, false, []PackTarget{{"1.21.1", "neoforge"}}, true},
		{"meta without loaders keeps", PruneRow{key, oldAt, nil, []jarmeta.LoaderMeta{metaFor("fabric", "~1.20")}}, false, []PackTarget{fabric121}, false},
		{"quilt pack accepts fabric jar", PruneRow{key, oldAt, []string{"fabric"}, []jarmeta.LoaderMeta{metaFor("fabric", "~1.21")}}, false, []PackTarget{{"1.21.1", "quilt"}}, false},
		{"quilt pack, fabric jar wrong mc", PruneRow{key, oldAt, []string{"fabric"}, []jarmeta.LoaderMeta{metaFor("fabric", "~1.20")}}, false, []PackTarget{{"1.21.1", "quilt"}}, true},
		{"fabric pack does not accept quilt jar", PruneRow{key, oldAt, []string{"quilt"}, []jarmeta.LoaderMeta{metaFor("quilt", "~1.21")}}, false, []PackTarget{fabric121}, true},
		{"loader listed but meta missing keeps", PruneRow{key, oldAt, []string{"fabric"}, nil}, false, []PackTarget{fabric121}, false},
		{"parse error, old, unreferenced", PruneRow{key, oldAt, nil, nil}, false, []PackTarget{fabric121}, true},
		{"parse error, recent", PruneRow{key, recentAt, nil, nil}, false, nil, false},
		{"parse error, referenced", PruneRow{key, oldAt, nil, nil}, true, nil, false},
		{"multi-loader jar, one target matches", PruneRow{key, oldAt, []string{"fabric", "neoforge"}, []jarmeta.LoaderMeta{metaFor("fabric", "~1.20"), metaFor("neoforge", "[1.21,1.22)")}}, false, []PackTarget{{"1.21.1", "neoforge"}}, false},
		{"second target matches", PruneRow{key, oldAt, []string{"fabric"}, []jarmeta.LoaderMeta{metaFor("fabric", "~1.21")}}, false, []PackTarget{{"1.20.1", "forge"}, fabric121}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := map[JarKey]struct{}{}
			if tt.referenced {
				ref[key] = struct{}{}
			}
			got := pruneCandidates([]PruneRow{tt.row}, ref, tt.targets, now)
			if deleted := slices.Contains(got, key); deleted != tt.want {
				t.Errorf("deleted = %v, want %v", deleted, tt.want)
			}
		})
	}
}

func TestPruneCandidatesMixed(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	oldAt := now.Add(-PruneAfter - time.Hour)
	a, b := JarKey{"sha1", "a"}, JarKey{"sha1", "b"}
	rows := []PruneRow{{Key: a, LastUsedAt: oldAt}, {Key: b, LastUsedAt: oldAt}}
	got := pruneCandidates(rows, map[JarKey]struct{}{b: {}}, nil, now)
	if !slices.Equal(got, []JarKey{a}) {
		t.Errorf("got %v", got)
	}
}
