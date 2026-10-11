package compat_svc

import (
	"slices"
	"time"

	"github.com/leocov-dev/packwiz-nxt/compat"
	"github.com/leocov-dev/packwiz-nxt/jarmeta"
)

// PruneAfter is how long a mod_jar_meta row must be unused before it may be
// pruned.
const PruneAfter = 180 * 24 * time.Hour

// JarKey identifies a mod_jar_meta row.
type JarKey struct {
	HashFormat string
	Hash       string
}

// PruneRow is a mod_jar_meta row as the prune decision sees it.
type PruneRow struct {
	Key        JarKey
	LastUsedAt time.Time
	// Loaders are the loader names found in the jar (empty on parse error).
	Loaders []string
	// Meta is the parsed metadata; empty when the jar could not be parsed.
	Meta []jarmeta.LoaderMeta
}

// PackTarget is the (Minecraft version, loader) a pack runs on.
type PackTarget struct {
	MinecraftVersion string
	Loader           string
}

// pruneCandidates returns the keys of rows that may be deleted. A row is
// deletable only if it was last used more than PruneAfter ago, no current mod
// references it (referenced), and it matches no current pack target.
func pruneCandidates(rows []PruneRow, referenced map[JarKey]struct{}, targets []PackTarget, now time.Time) []JarKey {
	var out []JarKey
	for _, r := range rows {
		if now.Sub(r.LastUsedAt) <= PruneAfter {
			continue
		}
		if _, ok := referenced[r.Key]; ok {
			continue
		}
		if matchesAnyTarget(r, targets) {
			continue
		}
		out = append(out, r.Key)
	}
	return out
}

func matchesAnyTarget(r PruneRow, targets []PackTarget) bool {
	return slices.ContainsFunc(targets, func(t PackTarget) bool { return jarMatchesTarget(r, t) })
}

// jarMatchesTarget reports whether the jar could be used by a pack with target
// t. It errs towards true: a missing, unparseable or unevaluable range counts
// as a match so the row is kept. Rows without loaders (parse errors) match
// nothing.
func jarMatchesTarget(r PruneRow, t PackTarget) bool {
	for _, loader := range r.Loaders {
		if !loaderAccepts(t.Loader, loader) {
			continue
		}
		meta, ok := findMeta(r.Meta, loader)
		if !ok {
			return true
		}
		if mcRangesAccept(meta, t.MinecraftVersion) {
			return true
		}
	}
	return false
}

// loaderAccepts mirrors compat: a Quilt pack also accepts Fabric jars.
func loaderAccepts(packLoader, jarLoader string) bool {
	return packLoader == jarLoader ||
		(packLoader == string(jarmeta.LoaderQuilt) && jarLoader == string(jarmeta.LoaderFabric))
}

func findMeta(metas []jarmeta.LoaderMeta, loader string) (jarmeta.LoaderMeta, bool) {
	for _, m := range metas {
		if string(m.Loader) == loader {
			return m, true
		}
	}
	return jarmeta.LoaderMeta{}, false
}

// mcRangesAccept reports whether mcVersion satisfies any of the jar's declared
// Minecraft ranges. No declared range, or a range that cannot be evaluated,
// accepts.
func mcRangesAccept(m jarmeta.LoaderMeta, mcVersion string) bool {
	ranges, declared := m.MinecraftRanges()
	if !declared || len(ranges) == 0 {
		return true
	}
	for _, rng := range ranges {
		ok, err := compat.Satisfies(m.Loader, rng, mcVersion)
		if err != nil || ok {
			return true
		}
	}
	return false
}
