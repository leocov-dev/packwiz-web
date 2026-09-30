package dto

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// DefaultOidcScopes are used when the form leaves scopes blank.
const DefaultOidcScopes = "openid profile email"

var oidcSlugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?$`)

// ValidateOidcSlug checks that a provider slug is safe to use in a URL path.
func ValidateOidcSlug(slug string) error {
	if !oidcSlugPattern.MatchString(slug) {
		return errors.New("slug must be 1-64 lowercase letters, digits or hyphens, and cannot start or end with a hyphen")
	}
	return nil
}

// ValidateOidcIssuerURL checks the issuer is an absolute https URL. Plain
// http is only tolerated for loopback hosts (local development).
func ValidateOidcIssuerURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return errors.New("issuer URL must be an absolute URL")
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return errors.New("issuer URL must not contain a query or fragment")
	}
	switch u.Scheme {
	case "https":
		return nil
	case "http":
		switch u.Hostname() {
		case "localhost", "127.0.0.1", "::1":
			return nil
		}
		return errors.New("issuer URL must use https")
	default:
		return errors.New("issuer URL must use https")
	}
}

// NormalizeOidcScopes trims and de-duplicates a space separated scope list,
// falling back to the defaults when blank, and requires the openid scope.
func NormalizeOidcScopes(raw string) (string, error) {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return DefaultOidcScopes, nil
	}
	seen := make(map[string]struct{}, len(fields))
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if _, ok := seen[f]; ok {
			continue
		}
		seen[f] = struct{}{}
		out = append(out, f)
	}
	if _, ok := seen["openid"]; !ok {
		return "", errors.New("scopes must include 'openid'")
	}
	return strings.Join(out, " "), nil
}

// OidcProviderForm is the create/update payload for an OIDC provider. On
// update a blank ClientSecret keeps the stored secret.
type OidcProviderForm struct {
	Slug            string `json:"slug" validate:"required"`
	DisplayName     string `json:"displayName" validate:"required,max=255"`
	IssuerURL       string `json:"issuerUrl" validate:"required,max=512"`
	ClientID        string `json:"clientId" validate:"required,max=512"`
	ClientSecret    string `json:"clientSecret"`
	Scopes          string `json:"scopes" validate:"max=512"`
	Enabled         bool   `json:"enabled"`
	AutoCreateUsers bool   `json:"autoCreateUsers"`
	LinkByEmail     bool   `json:"linkByEmail"`
}

func (f *OidcProviderForm) Validate() error {
	f.Slug = strings.TrimSpace(f.Slug)
	f.DisplayName = strings.TrimSpace(f.DisplayName)
	f.IssuerURL = strings.TrimRight(strings.TrimSpace(f.IssuerURL), "/")
	f.ClientID = strings.TrimSpace(f.ClientID)
	f.ClientSecret = strings.TrimSpace(f.ClientSecret)

	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(f); err != nil {
		return err
	}
	if err := ValidateOidcSlug(f.Slug); err != nil {
		return err
	}
	if err := ValidateOidcIssuerURL(f.IssuerURL); err != nil {
		return err
	}
	scopes, err := NormalizeOidcScopes(f.Scopes)
	if err != nil {
		return err
	}
	f.Scopes = scopes
	return nil
}

// OidcTestDiscoveryRequest asks the server to fetch an issuer's discovery
// document without saving anything.
type OidcTestDiscoveryRequest struct {
	IssuerURL string `json:"issuerUrl" validate:"required,max=512"`
}

func (f *OidcTestDiscoveryRequest) Validate() error {
	f.IssuerURL = strings.TrimRight(strings.TrimSpace(f.IssuerURL), "/")
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(f); err != nil {
		return err
	}
	return ValidateOidcIssuerURL(f.IssuerURL)
}

// OidcProviderResponse is the admin view of a provider. It never carries the
// client secret, only whether one is stored.
type OidcProviderResponse struct {
	ID              uint   `json:"id"`
	Slug            string `json:"slug"`
	DisplayName     string `json:"displayName"`
	IssuerURL       string `json:"issuerUrl"`
	ClientID        string `json:"clientId"`
	HasSecret       bool   `json:"hasSecret"`
	Scopes          string `json:"scopes"`
	Enabled         bool   `json:"enabled"`
	AutoCreateUsers bool   `json:"autoCreateUsers"`
	LinkByEmail     bool   `json:"linkByEmail"`
	RedirectURI     string `json:"redirectUri"`
	// Broken is true when the stored client secret cannot be decrypted; the
	// provider is unusable for login until the secret is re-entered.
	Broken      bool   `json:"broken"`
	BrokenError string `json:"brokenError,omitempty"`
}

// OidcPublicProvider is what the unauthenticated login page may see.
type OidcPublicProvider struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`
}

// OidcDiscoveryResult reports a successful discovery test.
type OidcDiscoveryResult struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorizationEndpoint"`
	TokenEndpoint         string `json:"tokenEndpoint"`
}
