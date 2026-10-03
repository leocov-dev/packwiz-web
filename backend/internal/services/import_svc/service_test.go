package import_svc

import (
	"testing"
	"time"

	"github.com/leocov-dev/packwiz-nxt/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"packwiz-web/internal/tables"
)

func taken(values ...string) func(string) bool {
	return func(v string) bool {
		for _, t := range values {
			if t == v {
				return true
			}
		}
		return false
	}
}

func TestUniqueSlug(t *testing.T) {
	assert.Equal(t, "pack", uniqueSlug("pack", taken()))
	assert.Equal(t, "pack-imported", uniqueSlug("pack", taken("pack")))
	assert.Equal(t, "pack-imported-2", uniqueSlug("pack", taken("pack", "pack-imported")))
	assert.Equal(t, "pack-imported-3", uniqueSlug("pack", taken("pack", "pack-imported", "pack-imported-2")))
}

func TestUniqueName(t *testing.T) {
	assert.Equal(t, "My Pack", uniqueName("My Pack", taken()))
	assert.Equal(t, "My Pack (imported)", uniqueName("My Pack", taken("My Pack")))
	assert.Equal(t, "My Pack (imported 2)", uniqueName("My Pack", taken("My Pack", "My Pack (imported)")))
}

func TestSlugFromName(t *testing.T) {
	assert.Equal(t, "my-cool-pack", slugFromName("My Cool Pack!", ""))
	assert.Equal(t, "example.com", slugFromName("***", "https://example.com/p/pack.toml"))
	assert.Equal(t, "pack", slugFromName("", "%%"))
}

func TestImportedDescription(t *testing.T) {
	at := time.Date(2026, 10, 3, 14, 5, 59, 0, time.UTC)
	assert.Equal(t,
		"imported from https://x/pack.toml on 2026/10/03 14:05",
		importedDescription("https://x/pack.toml", at, " "))
	assert.Equal(t,
		"imported from https://x/pack.toml on 2026/10/03 14:05\n\nhello",
		importedDescription("https://x/pack.toml", at, "hello"))
}

func TestBuildRows(t *testing.T) {
	s := &ImportService{now: func() time.Time { return time.Date(2026, 10, 3, 14, 5, 0, 0, time.UTC) }}
	src := &core.Pack{
		Name:        "Remote",
		Description: "desc",
		Versions:    map[string]string{"minecraft": "1.20.1", "fabric": "0.15.0"},
		Options:     map[string]interface{}{"acceptable-game-versions": []interface{}{"1.20", "1.20.2"}},
		Mods: map[string]*core.Mod{
			"sodium": {
				Name: "Sodium", FileName: "sodium.jar", ModType: "mods", Preserve: true,
				Update: core.ModUpdate{"modrinth": core.ModSourceData{"mod-id": "x"}},
				Option: &core.ModOption{Optional: true},
			},
			"manual": {Name: "Manual", ModType: "mods", Update: core.ModUpdate{}},
		},
	}

	pack, mods, skipped, err := s.buildRows(src, "https://x/pack.toml", tables.User{ID: 7})
	require.Nil(t, err)

	assert.Equal(t, "remote", pack.Slug)
	assert.Equal(t, "fabric", pack.Loader)
	assert.Equal(t, "1.0.0", pack.Version)
	assert.Equal(t, []string{"1.20", "1.20.2"}, []string(pack.AcceptableGameVersions))
	assert.Equal(t, "imported from https://x/pack.toml on 2026/10/03 14:05\n\ndesc", pack.Description)

	require.Len(t, mods, 1)
	assert.Equal(t, "sodium", mods[0].Slug)
	assert.Equal(t, "modrinth", mods[0].Source)
	assert.True(t, mods[0].Preserve)
	assert.True(t, mods[0].Option.Optional)
	assert.Equal(t, "sha256", mods[0].HashFormat)
	assert.Equal(t, uint(7), mods[0].CreatedBy)

	require.Len(t, skipped, 1)
	assert.Contains(t, skipped[0], "manual")
}

func TestBuildRows_RequiresLoaderAndMinecraft(t *testing.T) {
	s := &ImportService{now: time.Now}
	_, _, _, err := s.buildRows(&core.Pack{Versions: map[string]string{"fabric": "1"}}, "u", tables.User{})
	require.NotNil(t, err)
	_, _, _, err = s.buildRows(&core.Pack{Versions: map[string]string{"minecraft": "1.20"}}, "u", tables.User{})
	require.NotNil(t, err)
}
