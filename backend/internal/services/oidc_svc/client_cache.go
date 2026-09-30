package oidc_svc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/utils"
)

var (
	// ErrProviderDisabled is returned when a provider exists but is switched off.
	ErrProviderDisabled = errors.New("oidc provider is disabled")
	// ErrProviderBroken is returned when the stored client secret cannot be decrypted.
	ErrProviderBroken = errors.New("oidc provider client secret cannot be decrypted")
	// ErrNoPublicURL is returned when PWW_PUBLIC_URL is not configured, so no
	// redirect URI can be built.
	ErrNoPublicURL = errors.New("PWW_PUBLIC_URL is not configured")
)

const httpTimeout = 10 * time.Second

// Client bundles the discovered OIDC provider with the OAuth2 configuration
// used to run the authorization code flow against it.
type Client struct {
	ProviderID uint
	Provider   *oidc.Provider
	OAuth      oauth2.Config
}

// Verifier returns an ID token verifier bound to this provider's client id.
func (c *Client) Verifier() *oidc.IDTokenVerifier {
	return c.Provider.Verifier(&oidc.Config{ClientID: c.OAuth.ClientID})
}

// clientCache keeps one discovered Client per provider id. Entries are
// dropped whenever the provider is updated or deleted.
type clientCache struct {
	mu      sync.RWMutex
	entries map[uint]*Client
}

func newClientCache() *clientCache {
	return &clientCache{entries: make(map[uint]*Client)}
}

func (c *clientCache) get(id uint) (*Client, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cl, ok := c.entries[id]
	return cl, ok
}

func (c *clientCache) set(id uint, cl *Client) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[id] = cl
}

func (c *clientCache) invalidate(id uint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, id)
}

func httpContext(ctx context.Context) context.Context {
	return oidc.ClientContext(ctx, &http.Client{Timeout: httpTimeout})
}

// BuildRedirectURI returns the callback URL that must be registered with the
// identity provider. It is empty when publicURL is empty.
func BuildRedirectURI(publicURL, slug string) string {
	publicURL = strings.TrimRight(strings.TrimSpace(publicURL), "/")
	if publicURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/api/v1/auth/oidc/%s/callback", publicURL, slug)
}

// ClientFor returns the (cached) OIDC client for a provider. It fails with
// ErrProviderDisabled, ErrProviderBroken or ErrNoPublicURL without touching
// the network when the provider cannot be used.
func (s *ProviderService) ClientFor(ctx context.Context, p tables.OidcProvider) (*Client, error) {
	if !p.Enabled {
		return nil, ErrProviderDisabled
	}

	if cl, ok := s.cache.get(p.ID); ok {
		return cl, nil
	}

	redirectURI := BuildRedirectURI(s.publicURL, p.Slug)
	if redirectURI == "" {
		return nil, ErrNoPublicURL
	}

	secret := ""
	if p.ClientSecretEnc != "" {
		var err error
		secret, err = utils.SecretboxDecrypt(s.sessionSecret, p.ClientSecretEnc)
		if err != nil {
			return nil, ErrProviderBroken
		}
	}

	dctx, cancel := context.WithTimeout(httpContext(ctx), httpTimeout)
	defer cancel()

	provider, err := oidc.NewProvider(dctx, p.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery for provider %q failed: %w", p.Slug, err)
	}

	cl := &Client{
		ProviderID: p.ID,
		Provider:   provider,
		OAuth: oauth2.Config{
			ClientID:     p.ClientID,
			ClientSecret: secret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  redirectURI,
			Scopes:       strings.Fields(p.Scopes),
		},
	}
	s.cache.set(p.ID, cl)
	return cl, nil
}
