package packwiz_svc

import (
	"testing"

	"packwiz-web/internal/tables"
)

func TestBuildPublicPack(t *testing.T) {
	pack := tables.Pack{
		Slug:   "my-pack",
		Name:   "My Pack",
		Author: tables.User{Username: "alice"},
		Mods: []tables.Mod{
			{Slug: "zeta", Name: "zeta", Version: "1.0"},
			{Slug: "alpha", Name: "Alpha", Option: tables.OptionInfo{Optional: true}},
		},
	}

	got := buildPublicPack(pack, "https://x/packwiz/public/my-pack/pack.toml")

	if got.Author != "alice" {
		t.Errorf("author = %q", got.Author)
	}
	if got.AcceptableGameVersions == nil {
		t.Error("acceptable versions must be an empty slice, not nil")
	}
	if len(got.Mods) != 2 || got.Mods[0].Slug != "alpha" || !got.Mods[0].Optional {
		t.Errorf("mods not sorted by name or optional lost: %+v", got.Mods)
	}
}

func TestPackTomlURL(t *testing.T) {
	got := packTomlURL("https", "host.example:8080", "public", "my-pack")
	want := "https://host.example:8080/packwiz/public/my-pack/pack.toml"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
