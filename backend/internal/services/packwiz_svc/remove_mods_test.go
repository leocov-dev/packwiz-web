package packwiz_svc

import (
	"slices"
	"testing"

	"packwiz-web/internal/tables"
)

func TestBlockedRemovals(t *testing.T) {
	mods := []tables.Mod{
		{ID: 1, Name: "Lithium", DependencyIds: []uint{3}},
		{ID: 2, Name: "Sodium", DependencyIds: []uint{3, 4}},
		{ID: 3, Name: "Fabric API", IsDependency: true},
		{ID: 4, Name: "Indium", IsDependency: true},
		{ID: 5, Name: "Self", DependencyIds: []uint{5}},
	}
	tests := []struct {
		name   string
		remove map[uint]bool
		want   []string
	}{
		{"regular mods", map[uint]bool{1: true, 2: true}, []string{}},
		{"dependency with all requirers", map[uint]bool{1: true, 2: true, 3: true}, []string{}},
		{"dependency still required", map[uint]bool{1: true, 3: true}, []string{"Fabric API (required by Sodium)"}},
		{"dependency required by several", map[uint]bool{3: true, 4: true}, []string{"Fabric API (required by Lithium, Sodium)", "Indium (required by Sodium)"}},
		{"self reference ignored", map[uint]bool{5: true}, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := blockedRemovals(mods, tt.remove)
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
