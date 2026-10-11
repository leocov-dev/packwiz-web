package packwiz_svc

import (
	"testing"

	"github.com/leocov-dev/packwiz-nxt/core"

	"packwiz-web/internal/tables"
)

func mrRef(id, name string) projectRef {
	return projectRef{Source: sourceModrinth, ProjectID: id, Slug: id, Name: name}
}

func mrDep(id string) *core.Mod {
	return &core.Mod{Slug: id, Name: id, Update: core.ModUpdate{"modrinth": {"mod-id": id}}}
}

func matched(pos int, ref projectRef, incompatible []string, deps ...*core.Mod) lineMatch {
	return lineMatch{Position: pos, Status: tables.BulkItemAdded, Ref: ref,
		Match: &resolvedMod{Ref: ref, Mod: &core.Mod{Slug: ref.Slug, Name: ref.Name}, Deps: deps, Incompatible: incompatible}}
}

func statuses(out []itemOutcome) []string {
	s := make([]string, len(out))
	for i, o := range out {
		s[i] = o.Status
	}
	return s
}

func TestResolveInOrder(t *testing.T) {
	sodium := mrRef("sodium", "Sodium")
	embeddium := mrRef("embeddium", "Embeddium")
	lithium := mrRef("lithium", "Lithium")
	cfSodium := projectRef{Source: sourceCurseforge, ProjectID: "394468", Slug: "sodium", Name: "Sodium"}

	tests := []struct {
		name          string
		lines         []lineMatch
		installed     []projectRef
		slugs         map[string]bool
		want          []string
		conflictsWith []int
	}{
		{
			name:  "earlier line wins when it declares the later incompatible",
			lines: []lineMatch{matched(1, sodium, []string{"embeddium"}), matched(2, embeddium, nil)},
			want:  []string{tables.BulkItemAdded, tables.BulkItemConflict}, conflictsWith: []int{0, 1},
		},
		{
			name:  "earlier line wins when the later declares it incompatible",
			lines: []lineMatch{matched(1, embeddium, nil), matched(2, sodium, []string{"embeddium"})},
			want:  []string{tables.BulkItemAdded, tables.BulkItemConflict}, conflictsWith: []int{0, 1},
		},
		{
			name:      "installed mod wins",
			lines:     []lineMatch{matched(1, sodium, []string{"embeddium"})},
			installed: []projectRef{embeddium},
			want:      []string{tables.BulkItemConflict}, conflictsWith: []int{0},
		},
		{
			name:  "a dependency incompatible with an accepted mod blocks the line",
			lines: []lineMatch{matched(1, sodium, []string{"embeddium"}), matched(2, lithium, nil, mrDep("embeddium"))},
			want:  []string{tables.BulkItemAdded, tables.BulkItemConflict}, conflictsWith: []int{0, 1},
		},
		{
			name:  "a conflicting line accepts nothing, so later lines are unaffected",
			lines: []lineMatch{matched(1, embeddium, nil), matched(2, sodium, []string{"embeddium", "lithium"}), matched(3, lithium, nil)},
			want:  []string{tables.BulkItemAdded, tables.BulkItemConflict, tables.BulkItemAdded},
		},
		{
			name:  "same mod twice is skipped",
			lines: []lineMatch{matched(1, sodium, nil), matched(2, sodium, nil)},
			want:  []string{tables.BulkItemAdded, tables.BulkItemSkipped}, conflictsWith: []int{0, 1},
		},
		{
			name:  "same mod from another source is skipped by slug",
			lines: []lineMatch{matched(1, sodium, nil), matched(2, cfSodium, nil)},
			want:  []string{tables.BulkItemAdded, tables.BulkItemSkipped}, conflictsWith: []int{0, 1},
		},
		{
			name:      "installed mod is skipped",
			lines:     []lineMatch{matched(1, sodium, nil)},
			installed: []projectRef{sodium},
			want:      []string{tables.BulkItemSkipped},
		},
		{
			name:  "installed slug from another source is skipped",
			lines: []lineMatch{matched(1, cfSodium, nil)},
			slugs: map[string]bool{"sodium": true},
			want:  []string{tables.BulkItemSkipped},
		},
		{
			name:  "a mod pulled in as a dependency by an earlier line is skipped",
			lines: []lineMatch{matched(1, lithium, nil, mrDep("sodium")), matched(2, sodium, nil)},
			want:  []string{tables.BulkItemAdded, tables.BulkItemSkipped}, conflictsWith: []int{0, 1},
		},
		{
			name: "lines without a match keep their status",
			lines: []lineMatch{
				{Position: 1, Status: tables.BulkItemPending, Reason: "no exact match"},
				{Position: 2, Status: tables.BulkItemIncompatibleVersion},
				matched(3, sodium, nil),
			},
			want: []string{tables.BulkItemPending, tables.BulkItemIncompatibleVersion, tables.BulkItemAdded},
		},
		{
			name:  "declarations only apply within one source",
			lines: []lineMatch{matched(1, sodium, []string{"394468"}), matched(2, projectRef{Source: sourceCurseforge, ProjectID: "394468", Slug: "other", Name: "Other"}, nil)},
			want:  []string{tables.BulkItemAdded, tables.BulkItemAdded},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := resolveInOrder(tt.lines, tt.installed, tt.slugs)
			got := statuses(out)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got %v, want %v (reasons: %+v)", got, tt.want, out)
				}
				if out[i].Position != tt.lines[i].Position {
					t.Errorf("outcome %d position = %d", i, out[i].Position)
				}
				if tt.conflictsWith != nil && out[i].ConflictsWith != tt.conflictsWith[i] {
					t.Errorf("outcome %d conflictsWith = %d, want %d", i, out[i].ConflictsWith, tt.conflictsWith[i])
				}
				if (got[i] == tables.BulkItemConflict || got[i] == tables.BulkItemSkipped) && out[i].Reason == "" {
					t.Errorf("outcome %d has no reason", i)
				}
			}
		})
	}
}

func TestSkipReasonNamesDependency(t *testing.T) {
	out := resolveInOrder([]lineMatch{
		matched(1, mrRef("lithium", "Lithium"), nil, mrDep("sodium")),
		matched(2, mrRef("sodium", "Sodium"), nil),
	}, nil, nil)
	if want := "already added as a dependency of line 1"; out[1].Reason != want {
		t.Errorf("reason = %q, want %q", out[1].Reason, want)
	}
}

func TestConflictReasonNamesWinner(t *testing.T) {
	out := resolveInOrder([]lineMatch{
		matched(1, mrRef("sodium", "Sodium"), []string{"embeddium"}),
		matched(2, mrRef("embeddium", "Embeddium"), nil),
	}, nil, nil)
	if want := "incompatible with Sodium (line 1)"; out[1].Reason != want {
		t.Errorf("reason = %q, want %q", out[1].Reason, want)
	}
}

func TestInstalledRefs(t *testing.T) {
	mods := []tables.Mod{
		{Slug: "Sodium", Name: "Sodium", Source: sourceModrinth, Update: tables.UpdateInfo{"mod-id": "AANobbMI"}},
		{Slug: "jei", Name: "JEI", Source: sourceCurseforge, Update: tables.UpdateInfo{"project-id": float64(238222)}},
		{Slug: "manual", Name: "Manual", Source: "github", Update: tables.UpdateInfo{}},
	}
	refs, slugs := installedRefs(mods)
	if len(refs) != 2 {
		t.Fatalf("want 2 refs, got %+v", refs)
	}
	if refs[0].key() != "modrinth:AANobbMI" || refs[1].key() != "curseforge:238222" {
		t.Errorf("refs = %+v", refs)
	}
	if !slugs["sodium"] || !slugs["manual"] {
		t.Errorf("slugs must hold every mod, lowercase: %v", slugs)
	}
}

func TestCoreModRef(t *testing.T) {
	ref, ok := coreModRef(&core.Mod{Slug: "jei", Name: "JEI", Update: core.ModUpdate{"curseforge": {"project-id": uint32(238222)}}})
	if !ok || ref.key() != "curseforge:238222" {
		t.Errorf("ref = %+v, ok = %v", ref, ok)
	}
	if _, ok := coreModRef(&core.Mod{Update: core.ModUpdate{"github": {}}}); ok {
		t.Error("a mod without a project id has no ref")
	}
}
