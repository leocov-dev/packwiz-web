package tables

import (
	"reflect"
	"strings"
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

// Update data is stored as JSON, so ids come back from the database as float64.
// They must be served as TOML integers: packwiz-installer rejects "6928979.0".
func TestAsModTomlWritesCurseforgeIdsAsIntegers(t *testing.T) {
	var update UpdateInfo
	if err := update.Scan([]byte(`{"file-id": 6928979, "project-id": 419699}`)); err != nil {
		t.Fatal(err)
	}

	m := realisticMod("1.0")
	m.Source = "curseforge"
	m.Update = update

	got := modToml(t, m)

	for _, want := range []string{"file-id = 6928979\n", "project-id = 419699\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, ".0\n") {
		t.Errorf("found a float-formatted number in:\n%s", got)
	}
}

func TestIntegralNumbersAsInts(t *testing.T) {
	in := map[string]interface{}{
		"whole":  float64(42),
		"frac":   1.5,
		"huge":   float64(1 << 60),
		"str":    "x",
		"nested": map[string]interface{}{"id": float64(7)},
		"list":   []interface{}{float64(1), 2.5, "y"},
	}

	got := integralNumbersAsInts(in)

	want := map[string]interface{}{
		"whole":  int64(42),
		"frac":   1.5,
		"huge":   float64(1 << 60),
		"str":    "x",
		"nested": map[string]interface{}{"id": int64(7)},
		"list":   []interface{}{int64(1), 2.5, "y"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
	if _, ok := in["whole"].(float64); !ok {
		t.Error("input was modified")
	}
	if integralNumbersAsInts(nil) != nil {
		t.Error("nil input should stay nil")
	}
}
