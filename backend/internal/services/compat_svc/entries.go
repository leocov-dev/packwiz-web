package compat_svc

import (
	"errors"
	"strings"
	"time"

	"github.com/leocov-dev/packwiz-nxt/compat"
	"github.com/leocov-dev/packwiz-nxt/jarmeta"
	"gorm.io/datatypes"

	"packwiz-web/internal/tables"
)

const (
	// errorRetryAfter is how long a cached parse-error row is trusted before
	// the jar is downloaded again (the failure may have been a bad download).
	errorRetryAfter = 24 * time.Hour
	// maxErrorText caps the error text stored with a cached row or finding.
	maxErrorText = 2000
	// modTypeJar is the Mod.Type of jar mods; resource packs, shaders and
	// data packs are not scanned.
	modTypeJar = "mods"
)

// isJarMod reports whether a mod is a jar the scan reads. An empty type is the
// column default, "mods".
func isJarMod(m tables.Mod) bool {
	return m.Type == "" || m.Type == modTypeJar
}

// jarMods returns the mods the scan reads, in input order.
func jarMods(mods []tables.Mod) []tables.Mod {
	out := make([]tables.Mod, 0, len(mods))
	for _, m := range mods {
		if isJarMod(m) {
			out = append(out, m)
		}
	}
	return out
}

// jarKeyOf is the cache key of a mod's jar: its stored download hash, which is
// what the downloaded bytes are verified against. Case is normalised.
func jarKeyOf(m tables.Mod) JarKey {
	return JarKey{
		HashFormat: strings.ToLower(m.Download.HashFormat),
		Hash:       strings.ToLower(m.Download.Hash),
	}
}

// targetInputOf is everything a scan depends on. Only jar mods take part, so
// changing a resource pack does not make the result stale.
func targetInputOf(pack tables.Pack, mods []tables.Mod) TargetInput {
	jars := jarMods(mods)
	in := TargetInput{
		MinecraftVersion: pack.MCVersion,
		Loader:           pack.Loader,
		LoaderVersion:    pack.LoaderVersion,
		Mods:             make([]HashedMod, 0, len(jars)),
	}
	for _, m := range jars {
		key := jarKeyOf(m)
		in.Mods = append(in.Mods, HashedMod{
			Slug:       m.Slug,
			HashFormat: key.HashFormat,
			Hash:       key.Hash,
			Side:       string(m.Side),
		})
	}
	return in
}

// jarResult is what the scan knows about one jar.
type jarResult struct {
	Metas    []jarmeta.LoaderMeta
	ParseErr error
}

// cacheUsable reports whether a mod_jar_meta row can be used instead of
// downloading the jar again. A row from an older parser is never usable. A
// parse error is trusted for errorRetryAfter, except "no metadata", which is a
// property of the jar and never changes.
func cacheUsable(row tables.ModJarMeta, now time.Time) bool {
	if row.ParserVersion != ParserVersion {
		return false
	}
	if row.Error == "" || row.Error == jarmeta.ErrNoMetadata.Error() {
		return true
	}
	return now.Sub(row.ParsedAt) < errorRetryAfter
}

// resultFromRow turns a cached row back into a jarResult.
func resultFromRow(row tables.ModJarMeta) jarResult {
	var res jarResult
	if row.Meta != nil {
		res.Metas = *row.Meta
	}
	switch row.Error {
	case "":
	case jarmeta.ErrNoMetadata.Error():
		res.ParseErr = jarmeta.ErrNoMetadata
	default:
		res.ParseErr = errors.New(row.Error)
	}
	return res
}

// rowFromResult builds the cache row for a freshly parsed jar.
func rowFromResult(key JarKey, res jarResult) tables.ModJarMeta {
	row := tables.ModJarMeta{
		HashFormat:    key.HashFormat,
		Hash:          key.Hash,
		ParserVersion: ParserVersion,
	}
	if len(res.Metas) > 0 {
		metas := datatypes.JSONSlice[jarmeta.LoaderMeta](res.Metas)
		row.Meta = &metas
		for _, m := range res.Metas {
			row.Loaders = append(row.Loaders, string(m.Loader))
		}
	}
	if res.ParseErr != nil {
		row.Error = truncate(res.ParseErr.Error(), maxErrorText)
	}
	return row
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// entryFor builds the compat input of one mod.
func entryFor(m tables.Mod, res jarResult) compat.Entry {
	return compat.Entry{
		Slug:     m.Slug,
		Side:     m.Side,
		Metas:    res.Metas,
		ParseErr: res.ParseErr,
	}
}

// softenForUnscanned downgrades missing_dependency errors to warnings when
// some mods could not be scanned: the missing mod may be one of them, so the
// error could be false.
func softenForUnscanned(findings []compat.Finding, unscanned int) []compat.Finding {
	if unscanned == 0 {
		return findings
	}
	for i, f := range findings {
		if f.Code != compat.CodeMissingDependency || f.Severity != compat.SeverityError {
			continue
		}
		f.Severity = compat.SeverityWarning
		detail := make(map[string]string, len(f.Detail)+1)
		for k, v := range f.Detail {
			detail[k] = v
		}
		detail["note"] = "some mods could not be scanned, so this may be provided by one of them"
		f.Detail = detail
		findings[i] = f
	}
	return findings
}

// FindingCounts is the number of findings per severity.
type FindingCounts struct {
	Errors, Warnings, Infos int
}

// countFindings tallies findings by severity.
func countFindings(findings []compat.Finding) FindingCounts {
	var c FindingCounts
	for _, f := range findings {
		switch f.Severity {
		case compat.SeverityError:
			c.Errors++
		case compat.SeverityWarning:
			c.Warnings++
		default:
			c.Infos++
		}
	}
	return c
}

// findingRows maps compat findings to table rows. Findings whose slug is not a
// mod of the pack get no mod id.
func findingRows(packID uint, findings []compat.Finding, modIDBySlug map[string]uint) []tables.PackCompatFinding {
	rows := make([]tables.PackCompatFinding, 0, len(findings))
	for _, f := range findings {
		row := tables.PackCompatFinding{
			PackID:   packID,
			Slug:     f.Slug,
			Severity: string(f.Severity),
			Code:     string(f.Code),
			Message:  f.Message,
		}
		if id, ok := modIDBySlug[f.Slug]; ok {
			row.ModID = &id
		}
		if len(f.Detail) > 0 {
			row.Detail = make(datatypes.JSONMap, len(f.Detail))
			for k, v := range f.Detail {
				row.Detail[k] = v
			}
		}
		rows = append(rows, row)
	}
	return rows
}
