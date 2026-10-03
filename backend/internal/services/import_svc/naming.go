package import_svc

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	slugSuffix = "imported"
	nameSuffix = "imported"
)

var nonSlugChars = regexp.MustCompile(`[^a-z0-9._-]+`)

// slugFromName turns a pack name into a valid pack slug, falling back to the
// pack url's host when the name has nothing usable.
func slugFromName(name, packUrl string) string {
	slug := strings.Trim(nonSlugChars.ReplaceAllString(strings.ToLower(name), "-"), "-._")
	if slug != "" {
		return slug
	}
	if u, err := url.Parse(packUrl); err == nil {
		if host := strings.Trim(nonSlugChars.ReplaceAllString(strings.ToLower(u.Hostname()), "-"), "-._"); host != "" {
			return host
		}
	}
	return "pack"
}

// uniqueSlug returns base, or base-imported[-N] for the first value taken
// reports as free.
func uniqueSlug(base string, taken func(string) bool) string {
	if !taken(base) {
		return base
	}
	for n := 1; ; n++ {
		candidate := base + "-" + slugSuffix
		if n > 1 {
			candidate = fmt.Sprintf("%s-%d", candidate, n)
		}
		if !taken(candidate) {
			return candidate
		}
	}
}

// uniqueName returns base, or "base (imported[ N])" for the first value taken
// reports as free.
func uniqueName(base string, taken func(string) bool) string {
	if !taken(base) {
		return base
	}
	for n := 1; ; n++ {
		candidate := fmt.Sprintf("%s (%s)", base, nameSuffix)
		if n > 1 {
			candidate = fmt.Sprintf("%s (%s %d)", base, nameSuffix, n)
		}
		if !taken(candidate) {
			return candidate
		}
	}
}

// importedDescription prefixes the pack's own description with where and when
// it was imported from.
func importedDescription(packUrl string, at time.Time, description string) string {
	prefix := fmt.Sprintf("imported from %s on %s", packUrl, at.UTC().Format("2006/01/02 15:04"))
	if strings.TrimSpace(description) == "" {
		return prefix
	}
	return prefix + "\n\n" + description
}
