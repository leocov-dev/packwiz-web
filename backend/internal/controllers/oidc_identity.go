package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/params"
	"packwiz-web/internal/services/oidc_svc"
	"packwiz-web/internal/services/user_svc"
	"packwiz-web/internal/types/dto"
)

// OidcIdentityController lets a signed-in user manage their linked accounts.
type OidcIdentityController struct {
	user *user_svc.UserService
	flow *oidc_svc.FlowService
}

func NewOidcIdentityController(db *gorm.DB, flow *oidc_svc.FlowService) *OidcIdentityController {
	return &OidcIdentityController{
		user: user_svc.NewUserService(db),
		flow: flow,
	}
}

// -----------------------------------------------------------------------------

func (ic *OidcIdentityController) List(c *gin.Context) {
	user, err := mustBindCurrentUser(c)
	if err != nil {
		err.JSON(c)
		return
	}

	identities, err := ic.user.ListIdentities(user.ID)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, identities)
}

// Link starts a link round trip and returns the provider URL for the SPA to
// navigate to. The user id is sealed into the state cookie set here.
func (ic *OidcIdentityController) Link(c *gin.Context) {
	user, err := mustBindCurrentUser(c)
	if err != nil {
		err.JSON(c)
		return
	}

	slug, err := mustBindParam(c, params.OidcSlug)
	if err != nil {
		err.JSON(c)
		return
	}

	begin, fe := ic.flow.BeginLink(c.Request.Context(), slug, user)
	if fe != nil {
		flowErrorResponse(fe).JSON(c)
		return
	}

	setOidcStateCookie(c, begin.StateToken, oidc_svc.StateCookieMaxAge)
	dataOK(c, dto.LinkIdentityResponse{RedirectUrl: begin.AuthURL})
}

func (ic *OidcIdentityController) Unlink(c *gin.Context) {
	user, err := mustBindCurrentUser(c)
	if err != nil {
		err.JSON(c)
		return
	}

	identityId, err := mustBindIdParam(c, params.IdentityId)
	if err != nil {
		err.JSON(c)
		return
	}

	if err := ic.user.UnlinkIdentity(user.ID, identityId); err != nil {
		err.JSON(c)
		return
	}

	isOK(c)
}
