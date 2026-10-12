// Package compat_svc holds the jar compatibility check logic: cache keys,
// badge status and prune rules.
package compat_svc

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strings"
)

// ParserVersion is the version of the jar metadata parsing. Bump it to force
// every cached mod_jar_meta row to be parsed again.
const ParserVersion = 1

// targetHashVersion namespaces the hash input so its format can change.
const targetHashVersion = "v2"

// HashedMod is one pack mod as it enters the target hash.
type HashedMod struct {
	Slug       string
	HashFormat string
	Hash       string
	// Side is the mod's side (client/server/both); side_mismatch findings
	// depend on it.
	Side string
}

// TargetInput is everything a scan result depends on.
type TargetInput struct {
	MinecraftVersion string
	Loader           string
	LoaderVersion    string
	Mods             []HashedMod
}

// TargetHash returns a stable hex digest of the pack target and its mods. It is
// independent of mod order and changes when any input changes, so the UI can
// tell a stored result is stale.
func TargetHash(in TargetInput) string {
	lines := make([]string, 0, len(in.Mods))
	for _, m := range in.Mods {
		lines = append(lines, joinFields(m.Slug, m.HashFormat, m.Hash, m.Side))
	}
	slices.Sort(lines)

	h := sha256.New()
	write := func(s string) {
		h.Write([]byte(s))
		h.Write([]byte{'\n'})
	}
	write(targetHashVersion)
	write(joinFields(in.MinecraftVersion, in.Loader, in.LoaderVersion))
	for _, l := range lines {
		write(l)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// joinFields joins fields with NUL, so values cannot run into each other.
func joinFields(fields ...string) string {
	return strings.Join(fields, "\x00")
}
