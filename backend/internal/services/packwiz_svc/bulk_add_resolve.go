package packwiz_svc

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/leocov-dev/packwiz-nxt/core"

	"packwiz-web/internal/tables"
)

const (
	sourceModrinth   = "modrinth"
	sourceCurseforge = "curseforge"
)

// projectRef identifies a mod project within one source.
type projectRef struct {
	Source    string
	ProjectID string
	Slug      string
	Name      string
}

func (r projectRef) key() string {
	return r.Source + ":" + r.ProjectID
}

// resolvedMod is a mod resolved for adding: the mod, its missing required
// dependencies, and the project ids (same source) its file declares
// incompatible.
type resolvedMod struct {
	Ref          projectRef
	Mod          *core.Mod
	Deps         []*core.Mod
	Incompatible []string
}

// depRefs identifies the resolved dependencies. Dependencies without a known
// source or project id are left out.
func (m resolvedMod) depRefs() []projectRef {
	refs := make([]projectRef, 0, len(m.Deps))
	for _, d := range m.Deps {
		if ref, ok := coreModRef(d); ok {
			refs = append(refs, ref)
		}
	}
	return refs
}

// lineMatch is the lookup outcome of one line, before list order is applied.
// Match is set for an exact match with a compatible version; otherwise Status
// says why there is none.
type lineMatch struct {
	Position   int
	Status     string
	Reason     string
	Match      *resolvedMod
	Ref        projectRef
	Candidates []tables.BulkAddCandidate
}

// itemOutcome is the final state of one line after list order is applied.
type itemOutcome struct {
	Position      int
	Status        string
	Reason        string
	ConflictsWith int
	// Match is set for added lines (Status BulkItemAdded).
	Match *resolvedMod
}

// acceptedMod is a mod that is (or will be) in the pack after the lines
// resolved so far. Position is 0 for a mod installed before the run.
type acceptedMod struct {
	Ref          projectRef
	Position     int
	Dependency   bool
	Incompatible []string
}

// declaresIncompatible reports whether declarer lists other as incompatible. Declarations only name
// projects of the declaring mod's own source.
func declaresIncompatible(declarer projectRef, incompatible []string, other projectRef) bool {
	return declarer.Source == other.Source && slices.Contains(incompatible, other.ProjectID)
}

func acceptedLabel(a acceptedMod) string {
	if a.Position == 0 {
		return fmt.Sprintf("installed mod %s", a.Ref.Name)
	}
	if a.Dependency {
		return fmt.Sprintf("%s (dependency of line %d)", a.Ref.Name, a.Position)
	}
	return fmt.Sprintf("%s (line %d)", a.Ref.Name, a.Position)
}

// findConflict returns the first accepted mod that conflicts with m: either
// side declares the other incompatible, or an accepted mod declares one of m's
// dependencies incompatible.
func findConflict(m resolvedMod, accepted []acceptedMod) (acceptedMod, string, bool) {
	deps := m.depRefs()
	for _, a := range accepted {
		if declaresIncompatible(m.Ref, m.Incompatible, a.Ref) || declaresIncompatible(a.Ref, a.Incompatible, m.Ref) {
			return a, "incompatible with " + acceptedLabel(a), true
		}
		for _, d := range deps {
			if declaresIncompatible(a.Ref, a.Incompatible, d) {
				return a, fmt.Sprintf("requires %s, which is incompatible with %s", d.Name, acceptedLabel(a)), true
			}
		}
	}
	return acceptedMod{}, "", false
}

// resolveInOrder applies list order to the line lookups: lines are taken first
// to last and an earlier accepted mod (or an installed one) always wins. A
// line whose mod is already accepted (also as a dependency) is skipped; one that conflicts with an
// accepted mod is left for review. Accepted lines bring their dependencies
// into the accepted set.
//
// The same mod from another source is caught by slug (installedSlugs holds the
// pack's mod slugs, lowercase).
func resolveInOrder(lines []lineMatch, installed []projectRef, installedSlugs map[string]bool) []itemOutcome {
	byKey := make(map[string]int, len(installed))
	accepted := make([]acceptedMod, 0, len(installed)+len(lines))
	// slugPos maps a lowercase slug to the line that added it, 0 if installed
	slugPos := make(map[string]int, len(installedSlugs))
	for slug := range installedSlugs {
		slugPos[slug] = 0
	}
	accept := func(a acceptedMod) {
		byKey[a.Ref.key()] = len(accepted)
		accepted = append(accepted, a)
		if slug := strings.ToLower(a.Ref.Slug); slug != "" {
			if _, ok := slugPos[slug]; !ok {
				slugPos[slug] = a.Position
			}
		}
	}
	for _, ref := range installed {
		if _, ok := byKey[ref.key()]; !ok {
			accept(acceptedMod{Ref: ref})
		}
	}

	out := make([]itemOutcome, 0, len(lines))
	for _, l := range lines {
		if l.Match == nil {
			out = append(out, itemOutcome{Position: l.Position, Status: l.Status, Reason: l.Reason})
			continue
		}
		m := *l.Match

		if i, ok := byKey[m.Ref.key()]; ok {
			a := accepted[i]
			out = append(out, itemOutcome{Position: l.Position, Status: tables.BulkItemSkipped, ConflictsWith: a.Position,
				Reason: skipReason(a)})
			continue
		}
		if pos, ok := slugPos[strings.ToLower(m.Ref.Slug)]; ok {
			out = append(out, itemOutcome{Position: l.Position, Status: tables.BulkItemSkipped, ConflictsWith: pos,
				Reason: skipReason(acceptedMod{Position: pos})})
			continue
		}

		if a, reason, ok := findConflict(m, accepted); ok {
			out = append(out, itemOutcome{Position: l.Position, Status: tables.BulkItemConflict, Reason: reason, ConflictsWith: a.Position})
			continue
		}

		accept(acceptedMod{Ref: m.Ref, Position: l.Position, Incompatible: m.Incompatible})
		for _, d := range m.depRefs() {
			if _, ok := byKey[d.key()]; !ok {
				accept(acceptedMod{Ref: d, Position: l.Position, Dependency: true})
			}
		}
		out = append(out, itemOutcome{Position: l.Position, Status: tables.BulkItemAdded, Match: l.Match})
	}
	return out
}

func skipReason(a acceptedMod) string {
	switch {
	case a.Position == 0:
		return "already in the pack"
	case a.Dependency:
		return fmt.Sprintf("already added as a dependency of line %d", a.Position)
	default:
		return fmt.Sprintf("already added by line %d", a.Position)
	}
}

// idString formats a project id stored in mod update data (string, or a
// number after a JSON round trip). Returns "" for anything else or zero.
func idString(v any) string {
	switch id := v.(type) {
	case string:
		return id
	case int:
		if id != 0 {
			return strconv.Itoa(id)
		}
	case int64:
		if id != 0 {
			return strconv.FormatInt(id, 10)
		}
	case uint32:
		if id != 0 {
			return strconv.FormatUint(uint64(id), 10)
		}
	case uint:
		if id != 0 {
			return strconv.FormatUint(uint64(id), 10)
		}
	case float64:
		if id != 0 {
			return strconv.FormatInt(int64(id), 10)
		}
	}
	return ""
}

// sourceProjectID reads the project id from a mod's update data.
func sourceProjectID(source string, data map[string]any) string {
	switch source {
	case sourceModrinth:
		return idString(data["mod-id"])
	case sourceCurseforge:
		return idString(data["project-id"])
	}
	return ""
}

// coreModRef identifies a resolved (not yet stored) mod.
func coreModRef(m *core.Mod) (projectRef, bool) {
	source, data := tables.ExtractModSource(m)
	id := sourceProjectID(source, data)
	if id == "" {
		return projectRef{}, false
	}
	return projectRef{Source: source, ProjectID: id, Slug: m.Slug, Name: m.Name}, true
}

// installedRefs identifies a pack's stored mods, and returns their slugs
// (lowercase) to catch the same mod from another source.
func installedRefs(mods []tables.Mod) ([]projectRef, map[string]bool) {
	refs := make([]projectRef, 0, len(mods))
	slugs := make(map[string]bool, len(mods))
	for _, m := range mods {
		if m.Slug != "" {
			slugs[strings.ToLower(m.Slug)] = true
		}
		if id := sourceProjectID(m.Source, m.Update); id != "" {
			refs = append(refs, projectRef{Source: m.Source, ProjectID: id, Slug: m.Slug, Name: m.Name})
		}
	}
	return refs, slugs
}
