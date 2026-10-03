package packwiz_svc

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/leocov-dev/packwiz-nxt/core"

	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
)

// historySchemaVersion is the layout of historyPayload. Bump it, and teach
// decodeHistoryPayload the old layout, when the payload shape changes.
const historySchemaVersion = 1

// historyPack is the pack content a snapshot captures: everything pack.AsMeta
// reads. Slug, owner, status and public flag are deliberately not content.
type historyPack struct {
	Name                   string   `json:"name"`
	Description            string   `json:"description"`
	Version                string   `json:"version"`
	PackFormat             string   `json:"packFormat"`
	MCVersion              string   `json:"mcVersion"`
	Loader                 string   `json:"loader"`
	LoaderVersion          string   `json:"loaderVersion"`
	AcceptableGameVersions []string `json:"acceptableGameVersions"`
}

// historyMod is one mod's content. Dependencies are slugs, not row ids, so a
// snapshot survives mods being re-inserted under new ids.
type historyMod struct {
	Slug         string              `json:"slug"`
	Name         string              `json:"name"`
	FileName     string              `json:"fileName"`
	Side         core.ModSide        `json:"side"`
	Pinned       bool                `json:"pinned"`
	HashFormat   string              `json:"hashFormat"`
	Alias        string              `json:"alias"`
	Type         string              `json:"type"`
	Source       string              `json:"source"`
	Preserve     bool                `json:"preserve"`
	Download     tables.DownloadInfo `json:"download"`
	Update       tables.UpdateInfo   `json:"update"`
	Option       tables.OptionInfo   `json:"option"`
	Version      string              `json:"version"`
	IsDependency bool                `json:"isDependency"`
	Dependencies []string            `json:"dependencies"`
}

// historyPayload is the stored snapshot content. Mods are sorted by slug.
type historyPayload struct {
	V    int          `json:"v"`
	Pack historyPack  `json:"pack"`
	Mods []historyMod `json:"mods"`
}

func emptyHistoryPayload() historyPayload {
	return historyPayload{
		V:    historySchemaVersion,
		Pack: historyPack{AcceptableGameVersions: []string{}},
		Mods: []historyMod{},
	}
}

// buildHistoryMods maps rows to history mods, sorted by slug. Dependency ids
// that point at no mod in the list are dropped, matching filterValidDependencyIds.
func buildHistoryMods(mods []tables.Mod) []historyMod {
	slugByID := make(map[uint]string, len(mods))
	for _, m := range mods {
		slugByID[m.ID] = m.Slug
	}

	out := make([]historyMod, 0, len(mods))
	for _, m := range mods {
		deps := []string{}
		seen := map[string]struct{}{}
		for _, id := range m.DependencyIds {
			slug, ok := slugByID[id]
			if !ok {
				continue
			}
			if _, dup := seen[slug]; dup {
				continue
			}
			seen[slug] = struct{}{}
			deps = append(deps, slug)
		}
		sort.Strings(deps)

		out = append(out, historyMod{
			Slug:         m.Slug,
			Name:         m.Name,
			FileName:     m.FileName,
			Side:         m.Side,
			Pinned:       m.Pinned,
			HashFormat:   m.HashFormat,
			Alias:        m.Alias,
			Type:         m.Type,
			Source:       m.Source,
			Preserve:     m.Preserve,
			Download:     m.Download,
			Update:       normalizeUpdateInfo(m.Update),
			Option:       m.Option,
			Version:      m.Version,
			IsDependency: m.IsDependency,
			Dependencies: deps,
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Slug < out[j].Slug })
	return out
}

// normalizeUpdateInfo round-trips the update map through JSON so numbers have
// the same type (float64) whether the map came from a DB scan, an updater that
// wrote uint32s, or a decoded snapshot. Never nil, so it always encodes as {}.
func normalizeUpdateInfo(in tables.UpdateInfo) tables.UpdateInfo {
	if len(in) == 0 {
		return tables.UpdateInfo{}
	}
	raw, err := json.Marshal(in)
	if err != nil {
		return in
	}
	out := tables.UpdateInfo{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return in
	}
	return out
}

// buildHistoryPayload captures pack and its mods.
func buildHistoryPayload(pack tables.Pack, mods []tables.Mod) historyPayload {
	acceptable := []string{}
	acceptable = append(acceptable, pack.AcceptableGameVersions...)

	return historyPayload{
		V: historySchemaVersion,
		Pack: historyPack{
			Name:                   pack.Name,
			Description:            pack.Description,
			Version:                pack.Version,
			PackFormat:             pack.PackFormat,
			MCVersion:              pack.MCVersion,
			Loader:                 pack.Loader,
			LoaderVersion:          pack.LoaderVersion,
			AcceptableGameVersions: acceptable,
		},
		Mods: buildHistoryMods(mods),
	}
}

// encodeHistoryPayload returns the JSON to store and the content hash.
//
// The hash blanks every mod Version: it is display-only (never written to the
// pack files), and a version refresh alone must not count as a content change.
// Output is deterministic: struct fields keep declaration order, map keys are
// sorted by encoding/json, and mods are sorted by slug.
func encodeHistoryPayload(p historyPayload) ([]byte, string, error) {
	raw, err := json.Marshal(p)
	if err != nil {
		return nil, "", fmt.Errorf("encode snapshot payload: %w", err)
	}

	hashed := p
	hashed.Mods = make([]historyMod, len(p.Mods))
	copy(hashed.Mods, p.Mods)
	for i := range hashed.Mods {
		hashed.Mods[i].Version = ""
	}
	canonical, err := json.Marshal(hashed)
	if err != nil {
		return nil, "", fmt.Errorf("encode snapshot payload hash: %w", err)
	}

	sum := sha256.Sum256(canonical)
	return raw, hex.EncodeToString(sum[:]), nil
}

// decodeHistoryPayload reads a stored payload of the given schema version.
func decodeHistoryPayload(schemaVersion int, raw []byte) (historyPayload, error) {
	switch schemaVersion {
	case 1:
		var p historyPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return historyPayload{}, fmt.Errorf("decode snapshot payload: %w", err)
		}
		if p.Mods == nil {
			p.Mods = []historyMod{}
		}
		if p.Pack.AcceptableGameVersions == nil {
			p.Pack.AcceptableGameVersions = []string{}
		}
		for i := range p.Mods {
			if p.Mods[i].Dependencies == nil {
				p.Mods[i].Dependencies = []string{}
			}
			p.Mods[i].Update = normalizeUpdateInfo(p.Mods[i].Update)
		}
		return p, nil
	default:
		return historyPayload{}, fmt.Errorf("unsupported snapshot schema version %d", schemaVersion)
	}
}

func modRef(m historyMod) dto.SnapshotModRef {
	return dto.SnapshotModRef{Slug: m.Slug, Name: m.Name, Version: m.Version, FileName: m.FileName}
}

func jsonString(v any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(raw)
}

func change(field string, from, to any) dto.SnapshotFieldChange {
	return dto.SnapshotFieldChange{Field: field, From: from, To: to}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func diffHistoryPack(from, to historyPack) []dto.SnapshotFieldChange {
	out := []dto.SnapshotFieldChange{}
	add := func(field string, a, b string) {
		if a != b {
			out = append(out, change(field, a, b))
		}
	}
	add("name", from.Name, to.Name)
	add("description", from.Description, to.Description)
	add("version", from.Version, to.Version)
	add("packFormat", from.PackFormat, to.PackFormat)
	add("mcVersion", from.MCVersion, to.MCVersion)
	add("loader", from.Loader, to.Loader)
	add("loaderVersion", from.LoaderVersion, to.LoaderVersion)
	if !equalStrings(from.AcceptableGameVersions, to.AcceptableGameVersions) {
		out = append(out, change("acceptableGameVersions", from.AcceptableGameVersions, to.AcceptableGameVersions))
	}
	return out
}

func diffHistoryMod(from, to historyMod) []dto.SnapshotFieldChange {
	out := []dto.SnapshotFieldChange{}
	str := func(field string, a, b string) {
		if a != b {
			out = append(out, change(field, a, b))
		}
	}
	flag := func(field string, a, b bool) {
		if a != b {
			out = append(out, change(field, a, b))
		}
	}

	str("name", from.Name, to.Name)
	str("fileName", from.FileName, to.FileName)
	str("side", string(from.Side), string(to.Side))
	flag("pinned", from.Pinned, to.Pinned)
	str("hashFormat", from.HashFormat, to.HashFormat)
	str("alias", from.Alias, to.Alias)
	str("type", from.Type, to.Type)
	str("source", from.Source, to.Source)
	flag("preserve", from.Preserve, to.Preserve)
	str("download.url", from.Download.URL, to.Download.URL)
	str("download.mode", from.Download.Mode, to.Download.Mode)
	str("download.hash", from.Download.Hash, to.Download.Hash)
	str("download.hashFormat", from.Download.HashFormat, to.Download.HashFormat)
	if a, b := jsonString(from.Update), jsonString(to.Update); a != b {
		out = append(out, change("update", a, b))
	}
	flag("option.optional", from.Option.Optional, to.Option.Optional)
	str("option.description", from.Option.Description, to.Option.Description)
	flag("option.default", from.Option.Default, to.Option.Default)
	flag("isDependency", from.IsDependency, to.IsDependency)
	if !equalStrings(from.Dependencies, to.Dependencies) {
		out = append(out, change("dependencies", from.Dependencies, to.Dependencies))
	}
	// Version is display-only; shown in a diff but never part of the content hash.
	str("version", from.Version, to.Version)

	return out
}

// diffHistory reports what changed going from -> to. Pass emptyHistoryPayload()
// as from for a snapshot with no parent.
func diffHistory(from, to historyPayload) dto.SnapshotDiff {
	diff := dto.SnapshotDiff{
		Pack:    diffHistoryPack(from.Pack, to.Pack),
		Added:   []dto.SnapshotModRef{},
		Removed: []dto.SnapshotModRef{},
		Changed: []dto.SnapshotModChange{},
	}

	fromBySlug := make(map[string]historyMod, len(from.Mods))
	for _, m := range from.Mods {
		fromBySlug[m.Slug] = m
	}
	toBySlug := make(map[string]historyMod, len(to.Mods))
	for _, m := range to.Mods {
		toBySlug[m.Slug] = m
	}

	for _, m := range to.Mods {
		old, ok := fromBySlug[m.Slug]
		if !ok {
			diff.Added = append(diff.Added, modRef(m))
			continue
		}
		if changes := diffHistoryMod(old, m); len(changes) > 0 {
			diff.Changed = append(diff.Changed, dto.SnapshotModChange{SnapshotModRef: modRef(m), Changes: changes})
		}
	}
	for _, m := range from.Mods {
		if _, ok := toBySlug[m.Slug]; !ok {
			diff.Removed = append(diff.Removed, modRef(m))
		}
	}

	return diff
}

// summarizeHistoryDiff reduces a diff to the counts stored on the snapshot row.
func summarizeHistoryDiff(diff dto.SnapshotDiff) dto.SnapshotSummary {
	fields := make([]string, 0, len(diff.Pack))
	for _, c := range diff.Pack {
		fields = append(fields, c.Field)
	}
	return dto.SnapshotSummary{
		Added:      len(diff.Added),
		Removed:    len(diff.Removed),
		Changed:    len(diff.Changed),
		PackFields: fields,
	}
}

// historyRestoreUpdate rewrites the existing row ID to Mod's content.
type historyRestoreUpdate struct {
	ID  uint
	Mod historyMod
}

// historyRestorePlan is the row changes that turn the current mods into a snapshot.
type historyRestorePlan struct {
	Deletes []uint
	Updates []historyRestoreUpdate
	Inserts []historyMod
}

func sameHistoryMod(a, b historyMod) (bool, error) {
	ra, err := json.Marshal(a)
	if err != nil {
		return false, err
	}
	rb, err := json.Marshal(b)
	if err != nil {
		return false, err
	}
	return bytes.Equal(ra, rb), nil
}

// planHistoryRestore matches current rows to the target by slug. Matched rows
// keep their id, so mod_update_checks and other references survive; only rows
// whose content differs (Version and dependencies included) are updated, so
// untouched mods keep their updated_at.
func planHistoryRestore(current []tables.Mod, target historyPayload) (historyRestorePlan, error) {
	plan := historyRestorePlan{
		Deletes: []uint{},
		Updates: []historyRestoreUpdate{},
		Inserts: []historyMod{},
	}

	idBySlug := make(map[string]uint, len(current))
	for _, m := range current {
		idBySlug[m.Slug] = m.ID
	}
	currentBySlug := make(map[string]historyMod, len(current))
	for _, m := range buildHistoryMods(current) {
		currentBySlug[m.Slug] = m
	}

	targetSlugs := make(map[string]struct{}, len(target.Mods))
	for _, want := range target.Mods {
		targetSlugs[want.Slug] = struct{}{}

		have, ok := currentBySlug[want.Slug]
		if !ok {
			plan.Inserts = append(plan.Inserts, want)
			continue
		}
		same, err := sameHistoryMod(have, want)
		if err != nil {
			return historyRestorePlan{}, fmt.Errorf("compare mod %s: %w", want.Slug, err)
		}
		if !same {
			plan.Updates = append(plan.Updates, historyRestoreUpdate{ID: idBySlug[want.Slug], Mod: want})
		}
	}

	for _, m := range current {
		if _, ok := targetSlugs[m.Slug]; !ok {
			plan.Deletes = append(plan.Deletes, m.ID)
		}
	}
	sort.Slice(plan.Deletes, func(i, j int) bool { return plan.Deletes[i] < plan.Deletes[j] })

	return plan, nil
}

// resolveDependencyIds maps dependency slugs to row ids. Unknown slugs are
// dropped. The result is never nil.
func resolveDependencyIds(slugToID map[string]uint, deps []string) []uint {
	out := make([]uint, 0, len(deps))
	for _, slug := range deps {
		if id, ok := slugToID[slug]; ok {
			out = append(out, id)
		}
	}
	return out
}
