package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/params"
	"packwiz-web/internal/services/oidc_svc"
	"packwiz-web/internal/services/user_svc"
	"packwiz-web/internal/types/dto"
)

// OidcAdminController manages OIDC providers and users' linked accounts. It is
// mounted behind AdminGuard.
type OidcAdminController struct {
	user      *user_svc.UserService
	providers *oidc_svc.ProviderService
}

func NewOidcAdminController(db *gorm.DB, providers *oidc_svc.ProviderService) *OidcAdminController {
	return &OidcAdminController{
		user:      user_svc.NewUserService(db),
		providers: providers,
	}
}

// -----------------------------------------------------------------------------

func (ac *OidcAdminController) ListProviders(c *gin.Context) {
	providers, err := ac.providers.List()
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, providers)
}

func (ac *OidcAdminController) GetProvider(c *gin.Context) {
	id, err := mustBindIdParam(c, params.ProviderId)
	if err != nil {
		err.JSON(c)
		return
	}

	provider, err := ac.providers.Get(id)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, provider)
}

func (ac *OidcAdminController) CreateProvider(c *gin.Context) {
	var form dto.OidcProviderForm
	if err := mustBindJson(c, &form); err != nil {
		err.JSON(c)
		return
	}

	provider, err := ac.providers.Create(form)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, provider)
}

func (ac *OidcAdminController) UpdateProvider(c *gin.Context) {
	id, err := mustBindIdParam(c, params.ProviderId)
	if err != nil {
		err.JSON(c)
		return
	}

	var form dto.OidcProviderForm
	if err := mustBindJson(c, &form); err != nil {
		err.JSON(c)
		return
	}

	provider, err := ac.providers.Update(id, form)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, provider)
}

func (ac *OidcAdminController) DeleteProvider(c *gin.Context) {
	id, err := mustBindIdParam(c, params.ProviderId)
	if err != nil {
		err.JSON(c)
		return
	}

	if err := ac.providers.Delete(id); err != nil {
		err.JSON(c)
		return
	}

	isOK(c)
}

// OrphanedUsers reports how many users would lose their only login method if
// the provider were deleted, for the delete confirmation.
func (ac *OidcAdminController) OrphanedUsers(c *gin.Context) {
	id, err := mustBindIdParam(c, params.ProviderId)
	if err != nil {
		err.JSON(c)
		return
	}

	count, err := ac.providers.CountOrphanedUsers(id)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, gin.H{"count": count})
}

func (ac *OidcAdminController) TestDiscovery(c *gin.Context) {
	var request dto.OidcTestDiscoveryRequest
	if err := mustBindJson(c, &request); err != nil {
		err.JSON(c)
		return
	}

	result, err := ac.providers.TestDiscovery(c.Request.Context(), request.IssuerURL)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, result)
}

func (ac *OidcAdminController) ListUserIdentities(c *gin.Context) {
	userId, err := mustBindIdParam(c, params.UserID)
	if err != nil {
		err.JSON(c)
		return
	}

	identities, err := ac.user.ListIdentities(userId)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, identities)
}

func (ac *OidcAdminController) UnlinkUserIdentity(c *gin.Context) {
	userId, err := mustBindIdParam(c, params.UserID)
	if err != nil {
		err.JSON(c)
		return
	}

	identityId, err := mustBindIdParam(c, params.IdentityId)
	if err != nil {
		err.JSON(c)
		return
	}

	if err := ac.user.UnlinkIdentity(userId, identityId); err != nil {
		err.JSON(c)
		return
	}

	isOK(c)
}
