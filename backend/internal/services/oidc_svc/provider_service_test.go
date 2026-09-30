package oidc_svc

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"packwiz-web/internal/config"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/utils"
)

func TestBuildRedirectURI(t *testing.T) {
	cases := []struct {
		publicURL, slug, want string
	}{
		{"https://packwiz.example.com", "keycloak", "https://packwiz.example.com/api/v1/auth/oidc/keycloak/callback"},
		{"https://packwiz.example.com/", "google", "https://packwiz.example.com/api/v1/auth/oidc/google/callback"},
		{"  https://packwiz.example.com  ", "x", "https://packwiz.example.com/api/v1/auth/oidc/x/callback"},
		{"", "keycloak", ""},
	}
	for _, c := range cases {
		if got := BuildRedirectURI(c.publicURL, c.slug); got != c.want {
			t.Errorf("BuildRedirectURI(%q, %q) = %q, want %q", c.publicURL, c.slug, got, c.want)
		}
	}
}

func TestSecureSecretGuard(t *testing.T) {
	insecure := NewProviderService(nil, []byte(config.DefaultSessionSecret), "https://x")
	if err := insecure.requireSecureSecret(); err == nil {
		t.Fatal("expected refusal with default session secret")
	}
	// the guard runs before any db access, so create/update must fail cleanly
	if _, err := insecure.Create(dto.OidcProviderForm{}); err == nil {
		t.Fatal("Create should refuse with default session secret")
	}
	if _, err := insecure.Update(1, dto.OidcProviderForm{}); err == nil {
		t.Fatal("Update should refuse with default session secret")
	}

	secure := NewProviderService(nil, []byte("a-private-session-secret"), "https://x")
	if err := secure.requireSecureSecret(); err != nil {
		t.Fatalf("unexpected refusal: %v", err)
	}
}

func TestToResponseSecretState(t *testing.T) {
	svc := NewProviderService(nil, []byte("a-private-session-secret"), "https://pw.example.com")

	enc, err := svcEncrypt(svc, "s3cret")
	if err != nil {
		t.Fatal(err)
	}

	ok := svc.toResponse(tables.OidcProvider{Slug: "kc", ClientSecretEnc: enc})
	if !ok.HasSecret || ok.Broken {
		t.Fatalf("healthy provider: %+v", ok)
	}
	if ok.RedirectURI != "https://pw.example.com/api/v1/auth/oidc/kc/callback" {
		t.Fatalf("redirect uri: %q", ok.RedirectURI)
	}

	none := svc.toResponse(tables.OidcProvider{Slug: "kc"})
	if none.HasSecret || none.Broken {
		t.Fatalf("no secret: %+v", none)
	}

	rotated := NewProviderService(nil, []byte("a-different-session-secret"), "https://pw.example.com")
	broken := rotated.toResponse(tables.OidcProvider{Slug: "kc", ClientSecretEnc: enc})
	if !broken.HasSecret || !broken.Broken || broken.BrokenError == "" {
		t.Fatalf("rotated secret should mark provider broken: %+v", broken)
	}
}

func TestClientForFailsWithoutNetwork(t *testing.T) {
	enc, _ := svcEncrypt(NewProviderService(nil, []byte("a-private-session-secret"), ""), "s3cret")

	svc := NewProviderService(nil, []byte("a-private-session-secret"), "https://pw.example.com")
	ctx := context.Background()

	if _, err := svc.ClientFor(ctx, tables.OidcProvider{ID: 1, Slug: "kc", Enabled: false}); err != ErrProviderDisabled {
		t.Errorf("disabled: got %v", err)
	}

	noURL := NewProviderService(nil, []byte("a-private-session-secret"), "")
	if _, err := noURL.ClientFor(ctx, tables.OidcProvider{ID: 1, Slug: "kc", Enabled: true}); err != ErrNoPublicURL {
		t.Errorf("no public url: got %v", err)
	}

	rotated := NewProviderService(nil, []byte("a-different-session-secret"), "https://pw.example.com")
	if _, err := rotated.ClientFor(ctx, tables.OidcProvider{ID: 1, Slug: "kc", Enabled: true, ClientSecretEnc: enc}); err != ErrProviderBroken {
		t.Errorf("broken: got %v", err)
	}
}

func TestClientForAndDiscoveryWithFakeIdP(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"issuer": %q,
			"authorization_endpoint": %q,
			"token_endpoint": %q,
			"jwks_uri": %q,
			"id_token_signing_alg_values_supported": ["RS256"]
		}`, srv.URL, srv.URL+"/auth", srv.URL+"/token", srv.URL+"/jwks")
	}))
	defer srv.Close()

	svc := NewProviderService(nil, []byte("a-private-session-secret"), "https://pw.example.com")
	ctx := context.Background()

	res, serr := svc.TestDiscovery(ctx, srv.URL)
	if serr != nil {
		t.Fatalf("discovery: %v", serr)
	}
	if res.AuthorizationEndpoint != srv.URL+"/auth" || res.TokenEndpoint != srv.URL+"/token" {
		t.Fatalf("unexpected endpoints: %+v", res)
	}

	enc, _ := svcEncrypt(svc, "s3cret")
	p := tables.OidcProvider{
		ID: 7, Slug: "fake", Enabled: true, IssuerURL: srv.URL, ClientID: "cid",
		ClientSecretEnc: enc, Scopes: "openid email",
	}
	cl, err := svc.ClientFor(ctx, p)
	if err != nil {
		t.Fatalf("ClientFor: %v", err)
	}
	if cl.OAuth.ClientSecret != "s3cret" || cl.OAuth.ClientID != "cid" {
		t.Fatalf("oauth config: %+v", cl.OAuth)
	}
	if cl.OAuth.RedirectURL != "https://pw.example.com/api/v1/auth/oidc/fake/callback" {
		t.Fatalf("redirect url: %q", cl.OAuth.RedirectURL)
	}
	if len(cl.OAuth.Scopes) != 2 {
		t.Fatalf("scopes: %v", cl.OAuth.Scopes)
	}

	again, _ := svc.ClientFor(ctx, p)
	if again != cl {
		t.Fatal("expected cached client")
	}
	svc.cache.invalidate(p.ID)
	if fresh, _ := svc.ClientFor(ctx, p); fresh == cl {
		t.Fatal("expected fresh client after invalidate")
	}
}

func TestTestDiscoveryFailure(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()

	svc := NewProviderService(nil, []byte("a-private-session-secret"), "")
	if _, err := svc.TestDiscovery(context.Background(), srv.URL); err == nil {
		t.Fatal("expected discovery error")
	}
}

func svcEncrypt(svc *ProviderService, plain string) (string, error) {
	return utils.SecretboxEncrypt(svc.sessionSecret, plain)
}
