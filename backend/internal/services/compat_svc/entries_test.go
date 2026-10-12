package compat_svc

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/leocov-dev/packwiz-nxt/compat"
	"github.com/leocov-dev/packwiz-nxt/core"
	"github.com/leocov-dev/packwiz-nxt/jarmeta"

	"packwiz-web/internal/tables"
)

func mod(id uint, slug, typ, hash string) tables.Mod {
	return tables.Mod{
		ID: id, Slug: slug, Type: typ, Side: core.UniversalSide,
		Download: tables.DownloadInfo{HashFormat: "SHA256", Hash: hash},
	}
}

func TestJarModsAndTargetInput(t *testing.T) {
	mods := []tables.Mod{
		mod(1, "sodium", "mods", "AA"),
		mod(2, "legacy", "", "BB"),
		mod(3, "faithful", "resourcepacks", "CC"),
		mod(4, "complementary", "shaderpacks", "DD"),
		mod(5, "tags", "datapacks", "EE"),
	}
	jars := jarMods(mods)
	if len(jars) != 2 || jars[0].Slug != "sodium" || jars[1].Slug != "legacy" {
		t.Fatalf("jarMods = %v", jars)
	}

	pack := tables.Pack{MCVersion: "1.21.1", Loader: "fabric", LoaderVersion: "0.16.0"}
	in := targetInputOf(pack, mods)
	if len(in.Mods) != 2 {
		t.Fatalf("target mods = %d, want 2 (jars only)", len(in.Mods))
	}
	if in.Mods[0].HashFormat != "sha256" || in.Mods[0].Hash != "aa" {
		t.Errorf("hash not normalised: %+v", in.Mods[0])
	}

	// a resource pack change must not change the target hash
	base := TargetHash(targetInputOf(pack, mods))
	mods[2].Download.Hash = "CHANGED"
	if got := TargetHash(targetInputOf(pack, mods)); got != base {
		t.Error("non-jar mod changed the target hash")
	}
	mods[0].Download.Hash = "CHANGED"
	if got := TargetHash(targetInputOf(pack, mods)); got == base {
		t.Error("jar mod hash did not change the target hash")
	}
}

func TestCacheUsable(t *testing.T) {
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		row  tables.ModJarMeta
		want bool
	}{
		{"ok row", tables.ModJarMeta{ParserVersion: ParserVersion, ParsedAt: now.Add(-1000 * time.Hour)}, true},
		{"outdated parser", tables.ModJarMeta{ParserVersion: ParserVersion - 1}, false},
		{"newer parser", tables.ModJarMeta{ParserVersion: ParserVersion + 1}, false},
		{"fresh error", tables.ModJarMeta{ParserVersion: ParserVersion, Error: "boom", ParsedAt: now.Add(-time.Hour)}, true},
		{"error just inside window", tables.ModJarMeta{ParserVersion: ParserVersion, Error: "boom", ParsedAt: now.Add(-errorRetryAfter + time.Second)}, true},
		{"old error", tables.ModJarMeta{ParserVersion: ParserVersion, Error: "boom", ParsedAt: now.Add(-errorRetryAfter)}, false},
		{"old no-metadata is permanent", tables.ModJarMeta{ParserVersion: ParserVersion, Error: jarmeta.ErrNoMetadata.Error(), ParsedAt: now.Add(-1000 * time.Hour)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cacheUsable(tt.row, now); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRowResultRoundTrip(t *testing.T) {
	key := JarKey{HashFormat: "sha256", Hash: "ab"}
	metas := []jarmeta.LoaderMeta{{Loader: jarmeta.LoaderFabric, ModID: "a"}, {Loader: jarmeta.LoaderForge, ModID: "a"}}

	tests := []struct {
		name        string
		res         jarResult
		wantLoaders []string
		wantMeta    bool
		wantErrText string
	}{
		{"parsed", jarResult{Metas: metas}, []string{"fabric", "forge"}, true, ""},
		{"no metadata", jarResult{ParseErr: jarmeta.ErrNoMetadata}, nil, false, jarmeta.ErrNoMetadata.Error()},
		{"partial with error", jarResult{Metas: metas[:1], ParseErr: errors.New("bad forge")}, []string{"fabric"}, true, "bad forge"},
		{"malformed", jarResult{ParseErr: errors.New("malformed jar")}, nil, false, "malformed jar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := rowFromResult(key, tt.res)
			if row.ParserVersion != ParserVersion || row.HashFormat != "sha256" || row.Hash != "ab" {
				t.Errorf("row identity = %+v", row)
			}
			if (row.Meta != nil) != tt.wantMeta {
				t.Errorf("meta present = %v", row.Meta != nil)
			}
			if len(row.Loaders) != len(tt.wantLoaders) {
				t.Errorf("loaders = %v, want %v", row.Loaders, tt.wantLoaders)
			}
			if row.Error != tt.wantErrText {
				t.Errorf("error = %q, want %q", row.Error, tt.wantErrText)
			}

			back := resultFromRow(row)
			if len(back.Metas) != len(tt.res.Metas) {
				t.Errorf("metas = %d, want %d", len(back.Metas), len(tt.res.Metas))
			}
			if (back.ParseErr == nil) != (tt.res.ParseErr == nil) {
				t.Errorf("parse err = %v", back.ParseErr)
			}
			if errors.Is(tt.res.ParseErr, jarmeta.ErrNoMetadata) && !errors.Is(back.ParseErr, jarmeta.ErrNoMetadata) {
				t.Error("no-metadata sentinel lost in the round trip")
			}
		})
	}
}

func TestRowFromResultTruncatesError(t *testing.T) {
	row := rowFromResult(JarKey{}, jarResult{ParseErr: errors.New(strings.Repeat("a", maxErrorText*2))})
	if len(row.Error) != maxErrorText {
		t.Errorf("error length = %d, want %d", len(row.Error), maxErrorText)
	}
}

func TestEntryFor(t *testing.T) {
	m := mod(1, "sodium", "mods", "aa")
	m.Side = core.ClientSide
	metas := []jarmeta.LoaderMeta{{Loader: jarmeta.LoaderFabric, ModID: "sodium"}}
	e := entryFor(m, jarResult{Metas: metas, ParseErr: jarmeta.ErrNoMetadata})
	if e.Slug != "sodium" || e.Side != core.ClientSide || len(e.Metas) != 1 || e.ParseErr == nil {
		t.Errorf("entry = %+v", e)
	}
}

func TestCountFindings(t *testing.T) {
	fs := []compat.Finding{
		{Severity: compat.SeverityError}, {Severity: compat.SeverityError},
		{Severity: compat.SeverityWarning},
		{Severity: compat.SeverityInfo}, {Severity: compat.SeverityInfo}, {Severity: compat.SeverityInfo},
	}
	if got := countFindings(fs); got != (FindingCounts{Errors: 2, Warnings: 1, Infos: 3}) {
		t.Errorf("counts = %+v", got)
	}
	if got := countFindings(nil); got != (FindingCounts{}) {
		t.Errorf("empty counts = %+v", got)
	}
}

func TestSoftenForUnscanned(t *testing.T) {
	mk := func() []compat.Finding {
		return []compat.Finding{
			{Severity: compat.SeverityError, Code: compat.CodeMissingDependency, Slug: "a", Detail: map[string]string{"dep": "lib"}},
			{Severity: compat.SeverityError, Code: compat.CodeWrongLoader, Slug: "b"},
			{Severity: compat.SeverityWarning, Code: compat.CodeMissingDependency, Slug: "c"},
		}
	}

	if got := softenForUnscanned(mk(), 0); got[0].Severity != compat.SeverityError {
		t.Error("softened with nothing unscanned")
	}

	got := softenForUnscanned(mk(), 2)
	if got[0].Severity != compat.SeverityWarning || got[0].Detail["note"] == "" || got[0].Detail["dep"] != "lib" {
		t.Errorf("missing dependency not softened: %+v", got[0])
	}
	if got[1].Severity != compat.SeverityError {
		t.Error("unrelated error was softened")
	}
	if got[2].Severity != compat.SeverityWarning || got[2].Detail != nil {
		t.Errorf("existing warning changed: %+v", got[2])
	}
}

func TestFindingRows(t *testing.T) {
	fs := []compat.Finding{
		{Severity: compat.SeverityError, Code: compat.CodeBreaks, Slug: "a", Message: "m", Detail: map[string]string{"k": "v"}},
		{Severity: compat.SeverityInfo, Code: compat.CodeNoMetadata, Slug: "gone"},
	}
	rows := findingRows(7, fs, map[string]uint{"a": 42})
	if len(rows) != 2 {
		t.Fatalf("rows = %d", len(rows))
	}
	if rows[0].PackID != 7 || rows[0].ModID == nil || *rows[0].ModID != 42 || rows[0].Severity != "error" ||
		rows[0].Code != "breaks" || rows[0].Detail["k"] != "v" {
		t.Errorf("row0 = %+v", rows[0])
	}
	if rows[1].ModID != nil || rows[1].Detail != nil {
		t.Errorf("row1 = %+v", rows[1])
	}
}
