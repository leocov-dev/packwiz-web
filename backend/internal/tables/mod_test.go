package tables

import (
	"testing"

	"github.com/leocov-dev/packwiz-nxt/core"
)

func realisticMod(version string) Mod {
	return Mod{
		Slug:     "sodium",
		Name:     "Sodium",
		FileName: "sodium-fabric-0.5.11+mc1.21.jar",
		Version:  version,
		Side:     core.UniversalSide,
		Pinned:   true,
		Download: DownloadInfo{
			URL:        "https://cdn.modrinth.com/data/AANobbMI/versions/abc/sodium.jar",
			Mode:       "",
			Hash:       "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			HashFormat: "sha256",
		},
		HashFormat: "sha256",
		Type:       "mods",
		Source:     "modrinth",
		Update:     UpdateInfo{"mod-id": "AANobbMI", "version": "u1WlbJ4D"},
		Option:     OptionInfo{Optional: true, Description: "faster rendering", Default: true},
	}
}

func modToml(t *testing.T, m Mod) string {
	t.Helper()
	pack := Pack{Mods: []Mod{m}}.AsMeta()
	out, err := pack.AsModToml(m.Slug)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// The version is DB-only and must never change the generated .pw.toml.
func TestAsMetaVersionDoesNotAffectToml(t *testing.T) {
	without := modToml(t, realisticMod(""))
	with := modToml(t, realisticMod("0.5.11"))
	if without != with {
		t.Fatalf("toml differs with version set:\n--- empty\n%s\n--- set\n%s", without, with)
	}

	// even a core.Mod carrying a Version must serialize identically
	meta := realisticMod("").AsMeta()
	meta.Version = "0.5.11"
	got, _, err := meta.AsModToml()
	if err != nil {
		t.Fatal(err)
	}
	if got != without {
		t.Fatalf("core.Mod.Version leaked into toml:\n%s", got)
	}

	if realisticMod("0.5.11").AsMeta().Version != "" {
		t.Error("AsMeta must not set core.Mod.Version")
	}
}

func TestAsModTomlGolden(t *testing.T) {
	const golden = `name = 'Sodium'
filename = 'sodium-fabric-0.5.11+mc1.21.jar'
side = 'both'
pin = true

[download]
url = 'https://cdn.modrinth.com/data/AANobbMI/versions/abc/sodium.jar'
hash-format = 'sha256'
hash = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'

[update]
[update.modrinth]
mod-id = 'AANobbMI'
version = 'u1WlbJ4D'

[option]
optional = true
description = 'faster rendering'
default = true
`
	for _, v := range []string{"", "0.5.11"} {
		if got := modToml(t, realisticMod(v)); got != golden {
			t.Errorf("version %q: toml != golden:\n%s", v, got)
		}
	}
}
