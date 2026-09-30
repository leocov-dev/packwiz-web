package routes

import (
	"testing"

	"github.com/gin-gonic/gin"
	"packwiz-web/internal/services/oidc_svc"
)

// Gin panics at registration time on conflicting routes, so building the tree
// is a meaningful check that the OIDC routes coexist with the existing ones.
func TestOidcRoutesRegisterWithoutConflicts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	providers := oidc_svc.NewProviderService(nil, []byte("a-private-session-secret"), "https://pw.example.com")
	flow := oidc_svc.NewFlowService(nil, providers, []byte("a-private-session-secret"))

	v1 := router.Group("api/v1")
	RegisterOidcAuthRoutes(v1, nil, providers, flow)
	protected := v1.Group("")
	RegisterUserRoutes(protected, nil)
	RegisterOidcIdentityRoutes(protected, nil, flow)
	RegisterAdminRoutes(protected, nil)
	RegisterOidcAdminRoutes(protected, nil, providers)

	got := map[string]bool{}
	for _, r := range router.Routes() {
		got[r.Method+" "+r.Path] = true
	}

	for _, want := range []string{
		"GET /api/v1/auth/config",
		"GET /api/v1/auth/oidc/:slug/login",
		"GET /api/v1/auth/oidc/:slug/callback",
		"GET /api/v1/user/identities",
		"POST /api/v1/user/identities/:slug/link",
		"DELETE /api/v1/user/identities/:identityId",
		"POST /api/v1/user/password/set",
		"GET /api/v1/admin/oidc/providers",
		"POST /api/v1/admin/oidc/providers",
		"POST /api/v1/admin/oidc/providers/test",
		"GET /api/v1/admin/oidc/providers/:providerId",
		"PUT /api/v1/admin/oidc/providers/:providerId",
		"DELETE /api/v1/admin/oidc/providers/:providerId",
		"GET /api/v1/admin/oidc/providers/:providerId/orphaned-users",
		"GET /api/v1/admin/users/:userId/identities",
		"DELETE /api/v1/admin/users/:userId/identities/:identityId",
	} {
		if !got[want] {
			t.Errorf("missing route %s", want)
		}
	}
}
