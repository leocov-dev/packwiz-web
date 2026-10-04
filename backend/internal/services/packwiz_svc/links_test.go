package packwiz_svc

import (
	"testing"

	"packwiz-web/internal/tables"
)

func TestLinkKey(t *testing.T) {
	if got := linkKey(tables.Pack{IsPublic: true}, "tok"); got != publicLinkKey {
		t.Errorf("public pack: got %q", got)
	}
	if got := linkKey(tables.Pack{IsPublic: false}, "tok"); got != "tok" {
		t.Errorf("private pack: got %q", got)
	}
}

func TestConsumerLinksShareKeyAndSlug(t *testing.T) {
	if got := packTomlURL("https", "pw.example.com", "tok", "my-pack"); got != "https://pw.example.com/packwiz/tok/my-pack/pack.toml" {
		t.Errorf("pack.toml: %s", got)
	}
	pack := tables.Pack{Slug: "my-pack", Name: "My Pack #2: 100%?"}
	if got := instanceZipURL("https", "pw.example.com", "tok", pack); got != "https://pw.example.com/packwiz/tok/my-pack/multimc/My%20Pack%20%232%20100%25.zip" {
		t.Errorf("zip: %s", got)
	}
}

func TestInstanceZipFileName(t *testing.T) {
	for _, tc := range []struct{ name, want string }{
		{"My Pack", "My Pack.zip"},
		{"Pack v1.2", "Pack v1.2.zip"},
		{`a/b\c:d*e?f"g<h>i|j`, "abcdefghij.zip"},
		{"  spaced \t\n out  ", "spaced out.zip"},
		{"Ünïcødé パック", "Ünïcødé パック.zip"},
		{"...", "my-pack.zip"},
		{"", "my-pack.zip"},
	} {
		if got := InstanceZipFileName(tc.name, "my-pack"); got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.name, got, tc.want)
		}
	}
}
