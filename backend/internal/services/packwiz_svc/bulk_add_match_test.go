package packwiz_svc

import (
	"slices"
	"testing"

	"packwiz-web/internal/tables"
)

func TestParseBulkLines(t *testing.T) {
	text := "  Sodium \r\n\n# comment\nJust   Enough Items\nsodium\n  \nlithium"
	got := parseBulkLines(text)
	want := []string{"Sodium", "Just Enough Items", "lithium"}
	if !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
	if len(parseBulkLines("\n # only\n")) != 0 {
		t.Error("blank and comment lines must be dropped")
	}
}

func TestPickExact(t *testing.T) {
	hits := []tables.BulkAddCandidate{
		{Slug: "sodium-extra", Title: "Sodium Extra"},
		{Slug: "sodium", Title: "Sodium"},
		{Slug: "jei", Title: "Just Enough Items"},
		{Slug: "twin-a", Title: "Twin"},
		{Slug: "twin-b", Title: "twin"},
	}
	tests := []struct {
		line     string
		wantSlug string
		wantOK   bool
	}{
		{"sodium", "sodium", true},
		{"SODIUM", "sodium", true},
		{"just  enough items", "jei", true},
		{"Sodium Extra", "sodium-extra", true},
		{"twin", "", false},
		{"sod", "", false},
		{"twin-a", "twin-a", true},
	}
	for _, tt := range tests {
		got, ok := pickExact(tt.line, hits)
		if ok != tt.wantOK || (ok && got.Slug != tt.wantSlug) {
			t.Errorf("pickExact(%q) = %q, %v; want %q, %v", tt.line, got.Slug, ok, tt.wantSlug, tt.wantOK)
		}
	}
}

func TestLooksLikeSlug(t *testing.T) {
	for _, s := range []string{"sodium", "Fabric-API", "AANobbMI", "create_addon", "c2me.fabric"} {
		if !looksLikeSlug(s) {
			t.Errorf("%q should look like a slug", s)
		}
	}
	for _, s := range []string{"Just Enough Items", "a", "-dash", "mod!"} {
		if looksLikeSlug(s) {
			t.Errorf("%q should not look like a slug", s)
		}
	}
}

func TestParseLineURL(t *testing.T) {
	tests := []struct {
		line, source, slug string
		ok                 bool
	}{
		{"https://modrinth.com/mod/sodium", sourceModrinth, "sodium", true},
		{"https://www.curseforge.com/minecraft/mc-mods/jei", sourceCurseforge, "jei", true},
		{"sodium", "", "", false},
		{"https://example.com/mod/sodium", "", "", false},
	}
	for _, tt := range tests {
		source, slug, ok := parseLineURL(tt.line)
		if source != tt.source || slug != tt.slug || ok != tt.ok {
			t.Errorf("parseLineURL(%q) = %q, %q, %v", tt.line, source, slug, ok)
		}
	}
}
