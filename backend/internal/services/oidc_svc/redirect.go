package oidc_svc

import (
	"net/url"
	"strings"
)

// SanitizeRedirect returns raw if it is a same-origin relative path, else "/".
// It blocks absolute URLs, scheme-relative URLs ("//host"), backslash tricks
// and control characters so the post-login redirect can never leave the app.
func SanitizeRedirect(raw string) string {
	const fallback = "/"

	if raw == "" || !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") {
		return fallback
	}
	if strings.ContainsAny(raw, "\\") {
		return fallback
	}
	for _, r := range raw {
		if r < 0x20 || r == 0x7f {
			return fallback
		}
	}

	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "" || u.Host != "" || u.User != nil {
		return fallback
	}
	return raw
}
