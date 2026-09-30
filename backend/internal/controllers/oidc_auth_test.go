package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"packwiz-web/internal/services/oidc_svc"
)

// The sealed state is base64 and contains '+', '/' and '='; it must survive
// the trip through Set-Cookie and gin's Context.Cookie unchanged.
func TestOidcStateCookieRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)

	providers := oidc_svc.NewProviderService(nil, []byte("a-private-session-secret"), "")
	flow := oidc_svc.NewFlowService(nil, providers, []byte("a-private-session-secret"))

	for i := 0; i < 50; i++ {
		sealed, err := flow.EncodeState(oidc_svc.FlowState{State: "s", Nonce: "n", Mode: oidc_svc.ModeLogin, ProviderID: 1})
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		setOidcStateCookie(c, sealed, oidc_svc.StateCookieMaxAge)

		resp := rec.Result()
		cookies := resp.Cookies()
		if len(cookies) != 1 {
			t.Fatalf("expected one cookie, got %d", len(cookies))
		}
		got := cookies[0]
		if !got.HttpOnly || got.SameSite != http.SameSiteLaxMode || got.Path != oidcStateCookiePath || got.MaxAge != oidc_svc.StateCookieMaxAge {
			t.Fatalf("unexpected cookie attributes: %+v", got)
		}

		rec2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(rec2)
		c2.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c2.Request.AddCookie(&http.Cookie{Name: got.Name, Value: got.Value})

		read, err := c2.Cookie(oidcStateCookie)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := flow.DecodeState(read); err != nil {
			t.Fatalf("state did not survive the cookie round trip: %v", err)
		}
	}
}

func TestClearOidcStateCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	clearOidcStateCookie(c)

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge >= 0 || cookies[0].Value != "" {
		t.Fatalf("cookie not cleared: %+v", cookies)
	}
}
