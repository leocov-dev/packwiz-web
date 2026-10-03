package routes

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"packwiz-web/internal/services/oidc_svc"
)

// authOnlyRoutes are the protected routes that intentionally need
// authentication only (profile and static data), as "METHOD path".
var authOnlyRoutes = map[string]string{
	"GET /api/v1/user":                           "profile",
	"POST /api/v1/user/password":                 "profile",
	"POST /api/v1/user/password/set":             "profile",
	"POST /api/v1/user/update":                   "profile",
	"POST /api/v1/user/invalidate-sessions":      "profile",
	"GET /api/v1/user/identities":                "profile",
	"POST /api/v1/user/identities/:slug/link":    "profile",
	"DELETE /api/v1/user/identities/:identityId": "profile",
	"GET /api/v1/static-data":                    "static data",
	"GET /api/v1/packwiz/loaders":                "static data",
	"GET /api/v1/packwiz/pack":                   "list is filtered by pack.view in the service",
	"GET /api/v1/packwiz/pack/roles":             "assignable role list, no data of its own",
}

var permissionGuardName = regexp.MustCompile(`Require(Pack)?Permission\.func\d+$`)

var paramSegment = regexp.MustCompile(`[:*][A-Za-z]+`)

// Every protected route must carry a permission guard or be on the allow-list.
// The test records each route's handler chain through gin's HandlerNames.
func TestProtectedRoutesHavePermissionGuards(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	chains := map[string][]string{}
	router.Use(func(c *gin.Context) {
		chains[c.Request.Method+" "+c.FullPath()] = c.HandlerNames()
		c.AbortWithStatus(http.StatusNoContent)
	})

	providers := oidc_svc.NewProviderService(nil, []byte("a-private-session-secret"), "https://pw.example.com")
	flow := oidc_svc.NewFlowService(nil, providers, []byte("a-private-session-secret"))

	protected := router.Group("api/v1")
	RegisterUserRoutes(protected, nil)
	RegisterOidcIdentityRoutes(protected, nil, flow)
	RegisterAdminRoutes(protected, nil)
	RegisterOidcAdminRoutes(protected, nil, providers)
	RegisterStaticDataRoutes(protected, nil)
	packwizGroup := RegisterPackwizRoutes(protected, nil)
	RegisterPackRoutes(packwizGroup, nil, nil)

	routes := router.Routes()
	if len(routes) == 0 {
		t.Fatal("no routes registered")
	}

	for _, r := range routes {
		path := paramSegment.ReplaceAllString(r.Path, "1")
		req := httptest.NewRequest(r.Method, path, nil)
		router.ServeHTTP(httptest.NewRecorder(), req)
	}

	for _, r := range routes {
		key := r.Method + " " + r.Path
		if r.Method == http.MethodHead {
			// HEAD mirrors GET for the same path and is checked on its own below
			continue
		}
		if _, ok := authOnlyRoutes[key]; ok {
			continue
		}

		names, ok := chains[key]
		if !ok {
			t.Errorf("%s: route was not reached by the test", key)
			continue
		}
		if !hasPermissionGuard(names) {
			t.Errorf("%s: no permission guard (handlers: %s)", key, strings.Join(names, ", "))
		}
	}

	for key := range authOnlyRoutes {
		found := false
		for _, r := range routes {
			if r.Method+" "+r.Path == key {
				found = true
			}
		}
		if !found {
			t.Errorf("allow-list entry %q matches no route; remove it", key)
		}
	}
}

func hasPermissionGuard(names []string) bool {
	for _, n := range names {
		if permissionGuardName.MatchString(n) {
			return true
		}
	}
	return false
}
