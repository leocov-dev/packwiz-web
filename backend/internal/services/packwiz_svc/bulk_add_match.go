package packwiz_svc

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"codeberg.org/jmansfield/go-modrinth/modrinth"
	"github.com/leocov-dev/packwiz-nxt/core"
	"github.com/leocov-dev/packwiz-nxt/sources"

	"packwiz-web/internal/config"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
)

const (
	// maxBulkLines caps one bulk add request (each line costs several API calls).
	maxBulkLines = 200
	// bulkSearchLimit is how many search hits per source are fetched for a
	// line, and kept as review candidates.
	bulkSearchLimit = 10
)

// parseBulkLines splits pasted text into list lines, in order: trimmed, blank
// lines and "#" comments dropped, and repeats (ignoring case and spacing)
// dropped so the first occurrence keeps its priority.
func parseBulkLines(text string) []string {
	seen := map[string]bool{}
	var out []string
	for _, raw := range strings.Split(text, "\n") {
		line := strings.Join(strings.Fields(raw), " ")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key := strings.ToLower(line)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, line)
	}
	return out
}

func normalizeName(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.\-]{1,63}$`)

// looksLikeSlug reports whether a line could be a project slug or id, so a
// direct project lookup is worth trying before searching.
func looksLikeSlug(line string) bool {
	return slugPattern.MatchString(strings.ToLower(line))
}

// parseLineURL extracts the slug from a Modrinth or CurseForge mod page URL.
func parseLineURL(line string) (source, slug string, ok bool) {
	lower := strings.ToLower(line)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		return "", "", false
	}
	switch {
	case strings.Contains(lower, "modrinth.com/"):
		if slug := sources.ParseAsModrinthSlug(line); slug != "" {
			return sourceModrinth, slug, true
		}
	case strings.Contains(lower, "curseforge.com/"):
		if _, slug, _, err := sources.CurseforgeParseUrl(line); err == nil && slug != "" {
			return sourceCurseforge, slug, true
		}
	}
	return "", "", false
}

// pickExact returns the hit that exactly matches line: one whose slug equals
// it, else the only one whose title equals it (ignoring case and spacing). Two
// hits with the same title are ambiguous and match nothing.
func pickExact(line string, hits []tables.BulkAddCandidate) (tables.BulkAddCandidate, bool) {
	want := normalizeName(line)
	for _, h := range hits {
		if strings.ToLower(h.Slug) == want {
			return h, true
		}
	}
	var found tables.BulkAddCandidate
	count := 0
	for _, h := range hits {
		if normalizeName(h.Title) == want {
			found = h
			count++
		}
	}
	return found, count == 1
}

func modrinthCandidates(hits []*modrinth.SearchResult) []tables.BulkAddCandidate {
	out := make([]tables.BulkAddCandidate, 0, len(hits))
	for _, hit := range hits {
		r := modrinthResultFromHit(hit, false)
		out = append(out, tables.BulkAddCandidate{
			Source: sourceModrinth, Slug: r.Slug, Title: r.Title, ProjectId: r.ProjectId, IconUrl: r.IconUrl,
		})
	}
	return out
}

func curseforgeCandidates(items []cfSearchItem) []tables.BulkAddCandidate {
	out := make([]tables.BulkAddCandidate, 0, len(items))
	for _, item := range items {
		iconUrl := item.Logo.ThumbnailURL
		if iconUrl == "" {
			iconUrl = item.Logo.URL
		}
		out = append(out, tables.BulkAddCandidate{
			Source:    sourceCurseforge,
			Slug:      item.Slug,
			Title:     item.Name,
			ProjectId: strconv.FormatUint(uint64(item.ID), 10),
			IconUrl:   iconUrl,
		})
	}
	return out
}

// bulkMatcher looks up the lines of a bulk add for one pack.
type bulkMatcher struct {
	pack       tables.Pack
	meta       core.Pack
	versions   []string
	curseforge bool
}

func newBulkMatcher(pack tables.Pack) bulkMatcher {
	meta := pack.AsMeta()
	versions, err := meta.GetSupportedMCVersions()
	if err != nil {
		log.Warn("bulk add: no supported MC versions, searching all versions:", err)
	}
	return bulkMatcher{pack: pack, meta: meta, versions: versions, curseforge: config.HasCurseforgeApiKey()}
}

// lineLookup collects the attempts for one line.
type lineLookup struct {
	lineMatch
	incompatible *projectRef
	firstErr     error
}

// try records one resolve attempt and reports whether it is the match.
func (l *lineLookup) try(r resolvedMod, err error) bool {
	switch {
	case err == nil:
		l.Match, l.Ref, l.Status = &r, r.Ref, tables.BulkItemAdded
		return true
	case errors.Is(err, errNoCompatibleVersion):
		if l.incompatible == nil {
			ref := r.Ref
			l.incompatible = &ref
		}
	default:
		l.fail(err)
	}
	return false
}

func (l *lineLookup) fail(err error) {
	if l.firstErr == nil {
		l.firstErr = err
	}
}

// match finds the exact match of a line, in order: Modrinth slug or id,
// Modrinth search, then CurseForge search (when configured). A line without
// an exact match with a compatible version keeps the search hits as review
// candidates.
func (bm bulkMatcher) match(position int, line string) lineMatch {
	query, only := line, ""
	if source, slug, ok := parseLineURL(line); ok {
		query, only = slug, source
	}
	l := &lineLookup{lineMatch: lineMatch{Position: position}}

	if only != sourceCurseforge && looksLikeSlug(query) {
		// a failed lookup here just means the line is not a slug
		r, err := resolveModrinthMod(query, bm.meta)
		if err == nil || errors.Is(err, errNoCompatibleVersion) {
			if l.try(r, err) {
				return l.lineMatch
			}
		}
	}

	if only != sourceCurseforge {
		hits, err := searchModrinthHits(bm.meta, query, bm.versions, bulkSearchLimit)
		if err != nil {
			l.fail(err)
		}
		candidates := modrinthCandidates(hits)
		l.Candidates = append(l.Candidates, candidates...)
		if hit, ok := pickExact(query, candidates); ok && !l.triedRef(sourceModrinth, hit.ProjectId) {
			if l.try(resolveModrinthMod(hit.Slug, bm.meta)) {
				return l.lineMatch
			}
		}
	}

	if bm.curseforge && only != sourceModrinth {
		items, err := searchCurseforgeItems(bm.pack, query, nil, bulkSearchLimit)
		if err != nil {
			l.fail(err)
		}
		candidates := curseforgeCandidates(items)
		l.Candidates = append(l.Candidates, candidates...)
		if hit, ok := pickExact(query, candidates); ok {
			id, err := strconv.ParseUint(hit.ProjectId, 10, 32)
			if err != nil {
				l.fail(err)
			} else if l.try(resolveCurseforgeByID(uint32(id), bm.meta)) {
				return l.lineMatch
			}
		}
	}

	switch {
	case l.incompatible != nil:
		l.Status, l.Ref = tables.BulkItemIncompatibleVersion, *l.incompatible
		l.Reason = "no version for this pack's Minecraft version and loader"
	case l.firstErr != nil && len(l.Candidates) == 0:
		l.Status, l.Reason = tables.BulkItemFailed, l.firstErr.Error()
	default:
		l.Status, l.Reason = tables.BulkItemPending, "no exact match"
	}
	return l.lineMatch
}

// triedRef reports whether the project was already found incompatible.
func (l *lineLookup) triedRef(source, projectID string) bool {
	return l.incompatible != nil && l.incompatible.Source == source && l.incompatible.ProjectID == projectID
}
