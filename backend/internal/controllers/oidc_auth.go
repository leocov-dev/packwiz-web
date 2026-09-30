package controllers

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/config"
	"packwiz-web/internal/params"
	"packwiz-web/internal/services/oidc_svc"
	"packwiz-web/internal/services/user_svc"
	"packwiz-web/internal/types/response"
)

const (
	oidcStateCookie     = "oidc_state"
	oidcStateCookiePath = "/api/v1/auth/oidc"
)

// setOidcStateCookie stores the sealed flow state. It is SameSite=Lax (not
// Strict like the session cookie) because it must be sent when the IdP
// redirects the browser back from another site. The value is query-escaped
// because gin's Context.Cookie unescapes it when reading (base64 contains '+').
func setOidcStateCookie(c *gin.Context, value string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     oidcStateCookie,
		Value:    url.QueryEscape(value),
		Path:     oidcStateCookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   config.C.Mode != "development",
		SameSite: http.SameSiteLaxMode,
	})
}

func clearOidcStateCookie(c *gin.Context) {
	setOidcStateCookie(c, "", -1)
}

// flowErrorResponse maps a flow failure onto an API error for JSON endpoints.
func flowErrorResponse(fe *oidc_svc.FlowError) response.ServerError {
	status := http.StatusInternalServerError
	switch fe.Code {
	case oidc_svc.CodeNotPermitted:
		// not 403: the SPA treats every 403 as "not an admin" and navigates away
		status = http.StatusBadRequest
	case oidc_svc.CodeUnavailable:
		status = http.StatusServiceUnavailable
	case oidc_svc.CodeAlreadyLinked:
		status = http.StatusConflict
	}
	return response.New(status, fe.Message())
}

// OidcAuthController serves the public OIDC sign-in endpoints.
type OidcAuthController struct {
	user      *user_svc.UserService
	providers *oidc_svc.ProviderService
	flow      *oidc_svc.FlowService
}

func NewOidcAuthController(
	db *gorm.DB,
	providers *oidc_svc.ProviderService,
	flow *oidc_svc.FlowService,
) *OidcAuthController {
	return &OidcAuthController{
		user:      user_svc.NewUserService(db),
		providers: providers,
		flow:      flow,
	}
}

// -----------------------------------------------------------------------------

// Config lists the providers the login page may offer.
func (oc *OidcAuthController) Config(c *gin.Context) {
	providers, err := oc.providers.ListEnabled()
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, providers)
}

// Login starts a sign-in round trip. It is a full-page navigation, so failures
// redirect back to the login page instead of returning JSON.
func (oc *OidcAuthController) Login(c *gin.Context) {
	slug, err := mustBindParam(c, params.OidcSlug)
	if err != nil {
		c.Redirect(http.StatusFound, oidc_svc.FailureRedirect(oidc_svc.ModeLogin, oidc_svc.CodeUnavailable))
		return
	}

	begin, fe := oc.flow.BeginLogin(c.Request.Context(), slug, c.Query("redirect"))
	if fe != nil {
		c.Redirect(http.StatusFound, oidc_svc.FailureRedirect(oidc_svc.ModeLogin, fe.Code))
		return
	}

	setOidcStateCookie(c, begin.StateToken, oidc_svc.StateCookieMaxAge)
	c.Redirect(http.StatusFound, begin.AuthURL)
}

// Callback completes a round trip. On success for a sign-in it starts a
// session; it always ends in a redirect and never echoes IdP error text.
func (oc *OidcAuthController) Callback(c *gin.Context) {
	// the authorization code and state are in the URL; keep them out of
	// caches and Referer headers
	c.Header("Cache-Control", "no-store")
	c.Header("Referrer-Policy", "no-referrer")

	slug, err := mustBindParam(c, params.OidcSlug)
	if err != nil {
		c.Redirect(http.StatusFound, oidc_svc.FailureRedirect(oidc_svc.ModeLogin, oidc_svc.CodeFailed))
		return
	}

	// the state is single-use: clear it whatever happens next
	stateToken, _ := c.Cookie(oidcStateCookie)
	clearOidcStateCookie(c)

	result := oc.flow.HandleCallback(c.Request.Context(), oidc_svc.CallbackInput{
		Slug:          slug,
		StateToken:    stateToken,
		Code:          c.Query("code"),
		State:         c.Query("state"),
		ProviderError: c.Query("error") != "",
	})

	if result.ErrorCode == "" && result.Mode == oidc_svc.ModeLogin {
		user := result.User
		sessionKey := oc.user.GetOrMakeSessionKey(&user)
		if err := newSession(c, user.ID, sessionKey); err != nil {
			c.Redirect(http.StatusFound, oidc_svc.FailureRedirect(oidc_svc.ModeLogin, oidc_svc.CodeFailed))
			return
		}
	}

	c.Redirect(http.StatusFound, result.Redirect)
}
