package oidc_svc

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"gorm.io/gorm"
	"packwiz-web/internal/config"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
	"packwiz-web/internal/utils"
)

// ProviderService manages admin-configured OIDC providers.
type ProviderService struct {
	db            *gorm.DB
	sessionSecret []byte
	publicURL     string
	cache         *clientCache
}

// NewProviderService builds a ProviderService. sessionSecret keys the
// client-secret encryption; publicURL is the external base URL of the app.
func NewProviderService(db *gorm.DB, sessionSecret []byte, publicURL string) *ProviderService {
	return &ProviderService{
		db:            db,
		sessionSecret: sessionSecret,
		publicURL:     publicURL,
		cache:         newClientCache(),
	}
}

func (s *ProviderService) usingDefaultSecret() bool {
	return string(s.sessionSecret) == config.DefaultSessionSecret
}

func (s *ProviderService) requireSecureSecret() response.ServerError {
	if s.usingDefaultSecret() {
		return response.New(
			http.StatusBadRequest,
			"OIDC providers cannot be saved while the default session secret is in use; set PWW_SESSION_SECRET to a private random value and restart",
		)
	}
	return nil
}

// secretState reports whether a client secret is stored and whether it can
// still be decrypted with the current session secret.
func (s *ProviderService) secretState(p tables.OidcProvider) (hasSecret bool, broken bool) {
	if p.ClientSecretEnc == "" {
		return false, false
	}
	if _, err := utils.SecretboxDecrypt(s.sessionSecret, p.ClientSecretEnc); err != nil {
		return true, true
	}
	return true, false
}

func (s *ProviderService) toResponse(p tables.OidcProvider) dto.OidcProviderResponse {
	hasSecret, broken := s.secretState(p)
	res := dto.OidcProviderResponse{
		ID:              p.ID,
		Slug:            p.Slug,
		DisplayName:     p.DisplayName,
		IssuerURL:       p.IssuerURL,
		ClientID:        p.ClientID,
		HasSecret:       hasSecret,
		Scopes:          p.Scopes,
		Enabled:         p.Enabled,
		AutoCreateUsers: p.AutoCreateUsers,
		LinkByEmail:     p.LinkByEmail,
		RedirectURI:     BuildRedirectURI(s.publicURL, p.Slug),
		Broken:          broken,
	}
	if broken {
		res.BrokenError = "The stored client secret can no longer be decrypted (the session secret changed). Re-enter the client secret to restore this provider."
	}
	return res
}

// List returns every provider for the admin page.
func (s *ProviderService) List() ([]dto.OidcProviderResponse, response.ServerError) {
	var providers []tables.OidcProvider
	if err := s.db.Order("display_name ASC").Find(&providers).Error; err != nil {
		return nil, response.New(http.StatusInternalServerError, "failed to list oidc providers")
	}

	out := make([]dto.OidcProviderResponse, 0, len(providers))
	for _, p := range providers {
		out = append(out, s.toResponse(p))
	}
	return out, nil
}

// ListEnabled returns the providers that may be offered on the login page:
// enabled and not broken. Nothing is offered while PWW_PUBLIC_URL is unset,
// because no redirect URI can be built and every login would fail.
func (s *ProviderService) ListEnabled() ([]dto.OidcPublicProvider, response.ServerError) {
	if s.publicURL == "" {
		return []dto.OidcPublicProvider{}, nil
	}

	var providers []tables.OidcProvider
	if err := s.db.Where("enabled = ?", true).Order("display_name ASC").Find(&providers).Error; err != nil {
		return nil, response.New(http.StatusInternalServerError, "failed to list oidc providers")
	}

	out := make([]dto.OidcPublicProvider, 0, len(providers))
	for _, p := range providers {
		if _, broken := s.secretState(p); broken {
			continue
		}
		out = append(out, dto.OidcPublicProvider{Slug: p.Slug, DisplayName: p.DisplayName})
	}
	return out, nil
}

func (s *ProviderService) find(id uint) (tables.OidcProvider, response.ServerError) {
	var p tables.OidcProvider
	if err := s.db.Where("id = ?", id).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return p, response.New(http.StatusNotFound, fmt.Sprintf("oidc provider %d not found", id))
		}
		return p, response.New(http.StatusInternalServerError, "failed to load oidc provider")
	}
	return p, nil
}

// Get returns one provider for the admin page.
func (s *ProviderService) Get(id uint) (dto.OidcProviderResponse, response.ServerError) {
	p, err := s.find(id)
	if err != nil {
		return dto.OidcProviderResponse{}, err
	}
	return s.toResponse(p), nil
}

// GetBySlug returns the raw provider row for the login flow.
func (s *ProviderService) GetBySlug(slug string) (tables.OidcProvider, response.ServerError) {
	var p tables.OidcProvider
	if err := s.db.Where("slug = ?", slug).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return p, response.New(http.StatusNotFound, "oidc provider not found")
		}
		return p, response.New(http.StatusInternalServerError, "failed to load oidc provider")
	}
	return p, nil
}

func (s *ProviderService) slugTaken(slug string, exceptId uint) (bool, error) {
	var count int64
	if err := s.db.Model(&tables.OidcProvider{}).
		Where("slug = ? AND id <> ?", slug, exceptId).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Create stores a new provider, encrypting its client secret.
func (s *ProviderService) Create(form dto.OidcProviderForm) (dto.OidcProviderResponse, response.ServerError) {
	if err := s.requireSecureSecret(); err != nil {
		return dto.OidcProviderResponse{}, err
	}

	taken, err := s.slugTaken(form.Slug, 0)
	if err != nil {
		return dto.OidcProviderResponse{}, response.New(http.StatusInternalServerError, "failed to check for existing oidc provider")
	}
	if taken {
		return dto.OidcProviderResponse{}, response.New(http.StatusConflict, "an oidc provider with this slug already exists")
	}

	p := tables.OidcProvider{
		Slug:            form.Slug,
		DisplayName:     form.DisplayName,
		IssuerURL:       form.IssuerURL,
		ClientID:        form.ClientID,
		Scopes:          form.Scopes,
		Enabled:         form.Enabled,
		AutoCreateUsers: form.AutoCreateUsers,
		LinkByEmail:     form.LinkByEmail,
	}

	if form.ClientSecret != "" {
		enc, err := utils.SecretboxEncrypt(s.sessionSecret, form.ClientSecret)
		if err != nil {
			return dto.OidcProviderResponse{}, response.New(http.StatusInternalServerError, "failed to encrypt client secret")
		}
		p.ClientSecretEnc = enc
	}

	if err := s.db.Create(&p).Error; err != nil {
		return dto.OidcProviderResponse{}, response.New(http.StatusInternalServerError, "failed to create oidc provider")
	}

	return s.toResponse(p), nil
}

// Update changes a provider. A blank form.ClientSecret keeps the stored one.
func (s *ProviderService) Update(id uint, form dto.OidcProviderForm) (dto.OidcProviderResponse, response.ServerError) {
	if err := s.requireSecureSecret(); err != nil {
		return dto.OidcProviderResponse{}, err
	}

	p, serr := s.find(id)
	if serr != nil {
		return dto.OidcProviderResponse{}, serr
	}

	taken, err := s.slugTaken(form.Slug, id)
	if err != nil {
		return dto.OidcProviderResponse{}, response.New(http.StatusInternalServerError, "failed to check for existing oidc provider")
	}
	if taken {
		return dto.OidcProviderResponse{}, response.New(http.StatusConflict, "an oidc provider with this slug already exists")
	}

	p.Slug = form.Slug
	p.DisplayName = form.DisplayName
	p.IssuerURL = form.IssuerURL
	p.ClientID = form.ClientID
	p.Scopes = form.Scopes
	p.Enabled = form.Enabled
	p.AutoCreateUsers = form.AutoCreateUsers
	p.LinkByEmail = form.LinkByEmail

	if form.ClientSecret != "" {
		enc, err := utils.SecretboxEncrypt(s.sessionSecret, form.ClientSecret)
		if err != nil {
			return dto.OidcProviderResponse{}, response.New(http.StatusInternalServerError, "failed to encrypt client secret")
		}
		p.ClientSecretEnc = enc
	}

	if err := s.db.Save(&p).Error; err != nil {
		return dto.OidcProviderResponse{}, response.New(http.StatusInternalServerError, "failed to update oidc provider")
	}
	s.cache.invalidate(id)

	return s.toResponse(p), nil
}

// Delete removes a provider. Linked identities are removed by the database
// cascade; use CountOrphanedUsers first to warn about stranded users.
func (s *ProviderService) Delete(id uint) response.ServerError {
	if _, serr := s.find(id); serr != nil {
		return serr
	}

	if err := s.db.Delete(&tables.OidcProvider{}, id).Error; err != nil {
		return response.New(http.StatusInternalServerError, "failed to delete oidc provider")
	}
	s.cache.invalidate(id)
	return nil
}

// CountOrphanedUsers counts users for whom this provider is the only login
// method: they have no password and no identity at any other provider, so
// deleting the provider would lock them out.
func (s *ProviderService) CountOrphanedUsers(providerID uint) (int64, response.ServerError) {
	var count int64
	err := s.db.Raw(`
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		JOIN user_identities i ON i.user_id = u.id AND i.provider_id = ?
		WHERE u.deleted_at IS NULL
		  AND COALESCE(u.password, '') = ''
		  AND NOT EXISTS (
		      SELECT 1 FROM user_identities o
		      WHERE o.user_id = u.id AND o.provider_id <> ?
		  )`, providerID, providerID).Scan(&count).Error
	if err != nil {
		return 0, response.New(http.StatusInternalServerError, "failed to count affected users")
	}
	return count, nil
}

// TestDiscovery fetches the issuer's discovery document without saving
// anything, so admins can validate the issuer URL before creating a provider.
func (s *ProviderService) TestDiscovery(ctx context.Context, issuerURL string) (dto.OidcDiscoveryResult, response.ServerError) {
	dctx, cancel := context.WithTimeout(httpContext(ctx), httpTimeout)
	defer cancel()

	provider, err := oidc.NewProvider(dctx, issuerURL)
	if err != nil {
		log.Warn("oidc discovery test failed for ", issuerURL, ": ", err)
		return dto.OidcDiscoveryResult{}, response.New(
			http.StatusBadGateway,
			fmt.Sprintf("discovery failed: %s", err.Error()),
		)
	}

	endpoint := provider.Endpoint()
	return dto.OidcDiscoveryResult{
		Issuer:                issuerURL,
		AuthorizationEndpoint: endpoint.AuthURL,
		TokenEndpoint:         endpoint.TokenURL,
	}, nil
}
