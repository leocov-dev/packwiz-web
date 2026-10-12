package compat_svc

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/leocov-dev/packwiz-nxt/compat"
	"github.com/leocov-dev/packwiz-nxt/jarmeta"
	"gorm.io/datatypes"

	"packwiz-web/internal/tables"
)

const (
	// errorRetryAfter is how long a cached parse-error row is trusted before
	// the jar is downloaded again (the failure may have been a bad download).
	errorRetryAfter = 24 * time.Hour
	// maxDetailText caps one value of a finding's detail map.
	maxDetailText = 500
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
		metas := datatypes.JSONSlice[jarmeta.LoaderMeta](sanitizeMetas(res.Metas))
		row.Meta = &metas
		for _, m := range res.Metas {
			row.Loaders = append(row.Loaders, string(m.Loader))
		}
	}
	if res.ParseErr != nil {
		row.Error = sanitizeText(res.ParseErr.Error(), maxErrorText)
	}
	return row
}

// sanitizeText makes text from a jar, a parse error or a download error safe
// for a Postgres text/jsonb column and caps it at max bytes. Postgres rejects
// NUL and invalid UTF-8, so NULs are dropped and invalid bytes become U+FFFD;
// the cut never splits a multibyte character.
func sanitizeText(s string, max int) string {
	s = strings.ToValidUTF8(s, "\uFFFD")
	s = strings.ReplaceAll(s, "\x00", "")
	if len(s) <= max {
		return s
	}
	cut := max
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut]
}

// jsonNul is how encoding/json writes a NUL inside a string. Postgres jsonb
// rejects it.
var jsonNul = []byte(`\u0000`)

// stripJSONNul removes NUL escapes from encoded JSON. It walks escape pairs,
// so an escaped backslash followed by the text "u0000" is left alone.
func stripJSONNul(b []byte) []byte {
	if !bytes.Contains(b, jsonNul) {
		return b
	}
	out := make([]byte, 0, len(b))
	for i := 0; i < len(b); {
		if b[i] == '\\' {
			if bytes.HasPrefix(b[i:], jsonNul) {
				i += len(jsonNul)
				continue
			}
			if i+1 < len(b) {
				out = append(out, b[i], b[i+1])
				i += 2
				continue
			}
		}
		out = append(out, b[i])
		i++
	}
	return out
}

// sanitizeMetas returns metas that can be stored in a jsonb column: NULs that
// jars declare inside strings are removed. Metas without any are returned as is.
func sanitizeMetas(metas []jarmeta.LoaderMeta) []jarmeta.LoaderMeta {
	raw, err := json.Marshal(metas)
	if err != nil {
		return metas
	}
	clean := stripJSONNul(raw)
	if len(clean) == len(raw) {
		return metas
	}
	var out []jarmeta.LoaderMeta
	if err := json.Unmarshal(clean, &out); err != nil {
		return metas
	}
	return out
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
			Slug:     sanitizeText(f.Slug, maxErrorText),
			Severity: string(f.Severity),
			Code:     string(f.Code),
			Message:  sanitizeText(f.Message, maxErrorText),
		}
		if id, ok := modIDBySlug[f.Slug]; ok {
			row.ModID = &id
		}
		if len(f.Detail) > 0 {
			row.Detail = make(datatypes.JSONMap, len(f.Detail))
			for k, v := range f.Detail {
				row.Detail[sanitizeText(k, maxDetailText)] = sanitizeText(v, maxDetailText)
			}
		}
		rows = append(rows, row)
	}
	return rows
}
