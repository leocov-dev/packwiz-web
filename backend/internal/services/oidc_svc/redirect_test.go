package oidc_svc

import "testing"

func TestSanitizeRedirect(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", "/"},
		{"/", "/"},
		{"/packs/1/edit", "/packs/1/edit"},
		{"/packs?tab=mods#x", "/packs?tab=mods#x"},
		{"https://evil.example.com/", "/"},
		{"http://evil.example.com", "/"},
		{"//evil.example.com", "/"},
		{"///evil.example.com", "/"},
		{"/\\evil.example.com", "/"},
		{"\\\\evil.example.com", "/"},
		{"evil.example.com", "/"},
		{"javascript:alert(1)", "/"},
		{"/ok\r\nSet-Cookie: a=b", "/"},
		{"/ok\tx", "/"},
		{"/%0d%0a", "/%0d%0a"}, // percent-encoded stays inert in a Location path
	}
	for _, c := range cases {
		if got := SanitizeRedirect(c.in); got != c.want {
			t.Errorf("SanitizeRedirect(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
