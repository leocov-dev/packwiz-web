package packwiz_svc

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"gorm.io/datatypes"

	"packwiz-web/internal/tables"
)

func testPack() tables.Pack {
	return tables.Pack{
		Name:                   "Pack",
		Description:            "desc",
		Version:                "1.0.0",
		PackFormat:             "packwiz:1.1.0",
		MCVersion:              "1.21.1",
		Loader:                 "fabric",
		LoaderVersion:          "0.16.0",
		AcceptableGameVersions: datatypes.JSONSlice[string]{"1.21"},
	}
}

func testMod(id uint, slug string, deps ...uint) tables.Mod {
	return tables.Mod{
		ID:         id,
		Slug:       slug,
		Name:       slug,
		FileName:   slug + ".jar",
		Side:       "both",
		HashFormat: "sha256",
		Type:       "mods",
		Source:     "modrinth",
		Download:   tables.DownloadInfo{URL: "https://x/" + slug, Mode: "", Hash: "h-" + slug, HashFormat: "sha256"},
		Update:     tables.UpdateInfo{"mod-id": "abc", "version": "v1"},
		Version:    "1.0",

		DependencyIds: datatypes.JSONSlice[uint](deps),
	}
}

func mustEncode(t *testing.T, p historyPayload) ([]byte, string) {
	t.Helper()
	raw, hash, err := encodeHistoryPayload(p)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	return raw, hash
}

func TestHistoryHashIgnoresModOrderAndDependencyOrder(t *testing.T) {
	a := []tables.Mod{testMod(1, "a", 2, 3), testMod(2, "b"), testMod(3, "c")}
	b := []tables.Mod{testMod(3, "c"), testMod(1, "a", 3, 2), testMod(2, "b")}

	_, ha := mustEncode(t, buildHistoryPayload(testPack(), a))
	_, hb := mustEncode(t, buildHistoryPayload(testPack(), b))
	if ha != hb {
		t.Fatalf("hash differs for reordered mods/deps: %s vs %s", ha, hb)
	}
}

func TestHistoryHashIgnoresVersionOnlyChange(t *testing.T) {
	a := []tables.Mod{testMod(1, "a")}
	b := []tables.Mod{testMod(1, "a")}
	b[0].Version = "2.0"

	rawA, ha := mustEncode(t, buildHistoryPayload(testPack(), a))
	rawB, hb := mustEncode(t, buildHistoryPayload(testPack(), b))
	if ha != hb {
		t.Fatalf("version-only change altered the hash")
	}
	if string(rawA) == string(rawB) {
		t.Fatalf("stored payload must still carry the version")
	}
}

func TestHistoryHashChangesOnContent(t *testing.T) {
	base := []tables.Mod{testMod(1, "a")}
	_, h0 := mustEncode(t, buildHistoryPayload(testPack(), base))

	pinned := []tables.Mod{testMod(1, "a")}
	pinned[0].Pinned = true
	_, h1 := mustEncode(t, buildHistoryPayload(testPack(), pinned))

	pack := testPack()
	pack.MCVersion = "1.21.4"
	_, h2 := mustEncode(t, buildHistoryPayload(pack, base))

	if h0 == h1 || h0 == h2 || h1 == h2 {
		t.Fatalf("content changes must change the hash: %s %s %s", h0, h1, h2)
	}
}

func TestHistoryPayloadDropsDanglingDependencies(t *testing.T) {
	p := buildHistoryPayload(testPack(), []tables.Mod{testMod(1, "a", 2, 99), testMod(2, "b")})
	got := p.Mods[0].Dependencies
	if !reflect.DeepEqual(got, []string{"b"}) {
		t.Fatalf("deps = %v, want [b]", got)
	}
}

func TestHistoryPayloadNeverEncodesNullCollections(t *testing.T) {
	m := testMod(1, "a")
	m.Update = nil
	m.DependencyIds = nil
	pack := testPack()
	pack.AcceptableGameVersions = nil

	raw, _ := mustEncode(t, buildHistoryPayload(pack, []tables.Mod{m}))
	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatal(err)
	}
	mod := generic["mods"].([]any)[0].(map[string]any)
	if mod["update"] == nil || mod["dependencies"] == nil {
		t.Fatalf("update/dependencies must encode as {} / [], got %v / %v", mod["update"], mod["dependencies"])
	}
	if generic["pack"].(map[string]any)["acceptableGameVersions"] == nil {
		t.Fatalf("acceptableGameVersions must encode as []")
	}
}

func TestHistoryPayloadRoundTripKeepsLargeUpdateIDs(t *testing.T) {
	m := testMod(1, "a")
	// CurseForge file ids are large integers; they must survive and hash identically.
	m.Update = tables.UpdateInfo{"file-id": float64(5937001), "project-id": float64(238222)}
	before := buildHistoryPayload(testPack(), []tables.Mod{m})
	raw, hash := mustEncode(t, before)

	after, err := decodeHistoryPayload(historySchemaVersion, raw)
	if err != nil {
		t.Fatal(err)
	}
	_, hash2 := mustEncode(t, after)
	if hash != hash2 {
		t.Fatalf("hash changed across a decode round trip")
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("payload changed across a decode round trip\n%+v\n%+v", before, after)
	}
}

func TestHistoryUpdateInfoNumberTypesAreNormalized(t *testing.T) {
	// an updater may leave uint32 in the map; the DB scan yields float64
	withUint := testMod(1, "a")
	withUint.Update = tables.UpdateInfo{"file-id": uint32(5937001)}
	withFloat := testMod(1, "a")
	withFloat.Update = tables.UpdateInfo{"file-id": float64(5937001)}

	_, h1 := mustEncode(t, buildHistoryPayload(testPack(), []tables.Mod{withUint}))
	_, h2 := mustEncode(t, buildHistoryPayload(testPack(), []tables.Mod{withFloat}))
	if h1 != h2 {
		t.Fatalf("numeric type difference altered the hash")
	}
}

func TestDecodeHistoryPayloadRejectsUnknownSchema(t *testing.T) {
	if _, err := decodeHistoryPayload(99, []byte(`{}`)); err == nil {
		t.Fatal("expected an error for an unknown schema version")
	}
	if _, err := decodeHistoryPayload(1, []byte(`not json`)); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}

func TestDiffHistoryFromEmptyListsEverythingAdded(t *testing.T) {
	to := buildHistoryPayload(testPack(), []tables.Mod{testMod(1, "a"), testMod(2, "b")})
	diff := diffHistory(emptyHistoryPayload(), to)

	if len(diff.Added) != 2 || len(diff.Removed) != 0 || len(diff.Changed) != 0 {
		t.Fatalf("diff = %+v", diff)
	}
	if len(diff.Pack) == 0 {
		t.Fatal("pack fields must show as changed against the empty payload")
	}
}

func TestDiffHistoryAddedRemovedChanged(t *testing.T) {
	from := buildHistoryPayload(testPack(), []tables.Mod{testMod(1, "keep"), testMod(2, "gone"), testMod(3, "edit")})

	edited := testMod(3, "edit")
	edited.Pinned = true
	edited.Download.URL = "https://x/edit-v2"
	edited.Version = "2.0"
	pack := testPack()
	pack.MCVersion = "1.21.4"
	to := buildHistoryPayload(pack, []tables.Mod{testMod(1, "keep"), edited, testMod(4, "new")})

	diff := diffHistory(from, to)

	if len(diff.Added) != 1 || diff.Added[0].Slug != "new" {
		t.Errorf("added = %+v", diff.Added)
	}
	if len(diff.Removed) != 1 || diff.Removed[0].Slug != "gone" {
		t.Errorf("removed = %+v", diff.Removed)
	}
	if len(diff.Changed) != 1 || diff.Changed[0].Slug != "edit" {
		t.Fatalf("changed = %+v", diff.Changed)
	}
	fields := map[string]bool{}
	for _, c := range diff.Changed[0].Changes {
		fields[c.Field] = true
	}
	for _, want := range []string{"pinned", "download.url", "version"} {
		if !fields[want] {
			t.Errorf("missing changed field %q in %v", want, fields)
		}
	}
	if len(diff.Pack) != 1 || diff.Pack[0].Field != "mcVersion" {
		t.Errorf("pack changes = %+v", diff.Pack)
	}

	s := summarizeHistoryDiff(diff)
	if s.Added != 1 || s.Removed != 1 || s.Changed != 1 || !reflect.DeepEqual(s.PackFields, []string{"mcVersion"}) {
		t.Errorf("summary = %+v", s)
	}
}

func TestDiffHistoryIdenticalIsEmpty(t *testing.T) {
	p := buildHistoryPayload(testPack(), []tables.Mod{testMod(1, "a")})
	diff := diffHistory(p, p)
	if len(diff.Pack)+len(diff.Added)+len(diff.Removed)+len(diff.Changed) != 0 {
		t.Fatalf("diff of identical payloads = %+v", diff)
	}
}

func TestPlanHistoryRestore(t *testing.T) {
	current := []tables.Mod{
		testMod(10, "keep"),
		testMod(11, "edit"),
		testMod(12, "extra"),
	}
	current[1].Pinned = true

	// target: keep unchanged, edit un-pinned and now depending on fresh,
	// extra gone, fresh added depending on keep
	target := buildHistoryPayload(testPack(), []tables.Mod{
		testMod(1, "keep"),
		testMod(2, "edit", 3),
		testMod(3, "fresh", 1),
	})

	plan, err := planHistoryRestore(current, target)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(plan.Deletes, []uint{12}) {
		t.Errorf("deletes = %v", plan.Deletes)
	}
	if len(plan.Inserts) != 1 || plan.Inserts[0].Slug != "fresh" {
		t.Errorf("inserts = %+v", plan.Inserts)
	}
	if len(plan.Updates) != 1 || plan.Updates[0].ID != 11 || plan.Updates[0].Mod.Pinned {
		t.Errorf("updates = %+v", plan.Updates)
	}
}

func TestPlanHistoryRestoreNoopWhenIdentical(t *testing.T) {
	current := []tables.Mod{testMod(1, "a", 2), testMod(2, "b")}
	target := buildHistoryPayload(testPack(), current)

	plan, err := planHistoryRestore(current, target)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Deletes)+len(plan.Updates)+len(plan.Inserts) != 0 {
		t.Fatalf("expected an empty plan, got %+v", plan)
	}
}

func TestPlanHistoryRestoreUpdatesOnDependencyOrVersionChange(t *testing.T) {
	current := []tables.Mod{testMod(1, "a"), testMod(2, "b")}
	withDep := []tables.Mod{testMod(1, "a", 2), testMod(2, "b")}
	plan, err := planHistoryRestore(current, buildHistoryPayload(testPack(), withDep))
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Updates) != 1 || plan.Updates[0].ID != 1 {
		t.Errorf("dependency change: updates = %+v", plan.Updates)
	}

	withVersion := []tables.Mod{testMod(1, "a"), testMod(2, "b")}
	withVersion[1].Version = "9.9"
	plan, err = planHistoryRestore(current, buildHistoryPayload(testPack(), withVersion))
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Updates) != 1 || plan.Updates[0].ID != 2 {
		t.Errorf("version change: updates = %+v", plan.Updates)
	}
}

func TestResolveDependencyIds(t *testing.T) {
	ids := map[string]uint{"a": 1, "b": 2}
	got := resolveDependencyIds(ids, []string{"b", "missing", "a"})
	if !reflect.DeepEqual(got, []uint{2, 1}) {
		t.Fatalf("got %v", got)
	}
	if resolveDependencyIds(ids, nil) == nil {
		t.Fatal("result must not be nil")
	}
}

func TestCloneDescription(t *testing.T) {
	at := time.Date(2026, time.October, 3, 9, 5, 0, 0, time.UTC)

	if got, want := cloneDescription("My Pack", at, ""), "cloned from My Pack on 2026/10/03 09:05"; got != want {
		t.Errorf("empty description: got %q want %q", got, want)
	}

	got := cloneDescription("My Pack", at, "Original text")
	want := "cloned from My Pack on 2026/10/03 09:05\n\nOriginal text"
	if got != want {
		t.Errorf("with description: got %q want %q", got, want)
	}

	// 24 hour clock, zero padded
	late := time.Date(2026, time.January, 2, 23, 7, 0, 0, time.UTC)
	if got, want := cloneDescription("x", late, ""), "cloned from x on 2026/01/02 23:07"; got != want {
		t.Errorf("padding: got %q want %q", got, want)
	}
}
