package oidc_svc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"packwiz-web/internal/tables"
)

const testSecret = "a-private-session-secret"

// fakeIdP is a minimal OIDC provider: discovery, JWKS and a token endpoint
// that enforces PKCE and returns a signed ID token built from its fields.
type fakeIdP struct {
	t      *testing.T
	srv    *httptest.Server
	key    *rsa.PrivateKey
	signer jose.Signer

	mu        sync.Mutex
	challenge string // PKCE challenge the token endpoint must see verified

	// ID token knobs
	issuer        string
	audience      string
	nonce         string
	subject       string
	expiry        time.Time
	claims        map[string]interface{}
	omitIDToken   bool
	validCode     string
	clientSecret  string
	sawBadSecret  bool
	sawClientAuth bool
}

func newFakeIdP(t *testing.T) *fakeIdP {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: key, KeyID: "k1"}},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		t.Fatal(err)
	}

	idp := &fakeIdP{
		t: t, key: key, signer: signer,
		audience: "cid", subject: "user-1", validCode: "good-code", clientSecret: "s3cret",
		expiry: time.Now().Add(time.Hour),
		claims: map[string]interface{}{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"issuer":                                idp.srv.URL,
			"authorization_endpoint":                idp.srv.URL + "/auth",
			"token_endpoint":                        idp.srv.URL + "/token",
			"jwks_uri":                              idp.srv.URL + "/jwks",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{
			{Key: &key.PublicKey, KeyID: "k1", Algorithm: "RS256", Use: "sig"},
		}})
	})
	mux.HandleFunc("/token", idp.token)

	idp.srv = httptest.NewServer(mux)
	t.Cleanup(idp.srv.Close)
	return idp
}

func (i *fakeIdP) token(w http.ResponseWriter, r *http.Request) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	if id, secret, ok := r.BasicAuth(); ok {
		i.sawClientAuth = true
		if unescaped, err := url.QueryUnescape(secret); err != nil || unescaped != i.clientSecret || id != "cid" {
			i.sawBadSecret = true
		}
	}

	if r.Form.Get("grant_type") != "authorization_code" || r.Form.Get("code") != i.validCode {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
		return
	}

	sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
	if base64.RawURLEncoding.EncodeToString(sum[:]) != i.challenge {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant", "error_description": "PKCE failed"})
		return
	}

	resp := map[string]interface{}{"access_token": "at", "token_type": "Bearer"}
	if !i.omitIDToken {
		issuer := i.issuer
		if issuer == "" {
			issuer = i.srv.URL
		}
		claims := map[string]interface{}{
			"iss":   issuer,
			"aud":   i.audience,
			"sub":   i.subject,
			"iat":   time.Now().Add(-time.Minute).Unix(),
			"exp":   i.expiry.Unix(),
			"nonce": i.nonce,
		}
		for k, v := range i.claims {
			claims[k] = v
		}
		raw, err := jwt.Signed(i.signer).Claims(claims).Serialize()
		if err != nil {
			i.t.Fatal(err)
		}
		resp["id_token"] = raw
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type flowHarness struct {
	idp  *fakeIdP
	flow *FlowService
	prov tables.OidcProvider
	cl   *Client
}

func newFlowHarness(t *testing.T) *flowHarness {
	t.Helper()
	idp := newFakeIdP(t)

	providers := NewProviderService(nil, []byte(testSecret), "https://pw.example.com")
	enc, err := svcEncrypt(providers, idp.clientSecret)
	if err != nil {
		t.Fatal(err)
	}
	prov := tables.OidcProvider{
		ID: 3, Slug: "fake", Enabled: true, IssuerURL: idp.srv.URL, ClientID: "cid",
		ClientSecretEnc: enc, Scopes: "openid profile email",
	}
	cl, err := providers.ClientFor(context.Background(), prov)
	if err != nil {
		t.Fatalf("ClientFor: %v", err)
	}

	return &flowHarness{
		idp:  idp,
		flow: NewFlowService(nil, providers, []byte(testSecret)),
		prov: prov,
		cl:   cl,
	}
}

// start runs Begin and wires the fake IdP with the nonce and PKCE challenge it
// would have received on the authorization request.
func (h *flowHarness) start(t *testing.T, mode FlowMode, userID uint) (st FlowState, echoedState string) {
	t.Helper()

	res, fe := h.flow.beginWith(context.Background(), h.prov, mode, userID, "/packs")
	if fe != nil {
		t.Fatalf("begin: %v", fe)
	}

	u, err := url.Parse(res.AuthURL)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("code_challenge_method") != "S256" || q.Get("code_challenge") == "" {
		t.Fatalf("auth url lacks PKCE: %s", res.AuthURL)
	}
	if q.Get("nonce") == "" || q.Get("state") == "" {
		t.Fatalf("auth url lacks nonce/state: %s", res.AuthURL)
	}
	if q.Get("client_id") != "cid" || q.Get("response_type") != "code" {
		t.Fatalf("unexpected auth url: %s", res.AuthURL)
	}
	if q.Get("redirect_uri") != "https://pw.example.com/api/v1/auth/oidc/fake/callback" {
		t.Fatalf("redirect_uri: %s", q.Get("redirect_uri"))
	}
	if !strings.HasPrefix(res.AuthURL, h.idp.srv.URL+"/auth") {
		t.Fatalf("auth url endpoint: %s", res.AuthURL)
	}

	h.idp.mu.Lock()
	h.idp.challenge = q.Get("code_challenge")
	h.idp.nonce = q.Get("nonce")
	h.idp.mu.Unlock()

	st, err = h.flow.DecodeState(res.StateToken)
	if err != nil {
		t.Fatalf("decode state: %v", err)
	}
	return st, q.Get("state")
}

func (h *flowHarness) complete(st FlowState, code, echoed string) (Claims, error) {
	return h.flow.complete(context.Background(), h.cl, st, code, echoed)
}

func TestFlowHappyPath(t *testing.T) {
	h := newFlowHarness(t)
	h.idp.claims = map[string]interface{}{
		"email": " Jane@Example.com ", "email_verified": true,
		"preferred_username": "jane", "name": "Jane Doe",
	}

	st, echoed := h.start(t, ModeLogin, 0)
	if st.Mode != ModeLogin || st.ProviderID != 3 || st.Redirect != "/packs" {
		t.Fatalf("state: %+v", st)
	}

	claims, err := h.complete(st, "good-code", echoed)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	want := Claims{
		Subject: "user-1", Email: "jane@example.com", EmailVerified: true,
		PreferredUsername: "jane", Name: "Jane Doe",
	}
	if claims != want {
		t.Fatalf("claims = %+v, want %+v", claims, want)
	}
	if h.idp.sawBadSecret {
		t.Fatal("client secret was not presented correctly")
	}
}

func TestFlowEmailVerifiedAsString(t *testing.T) {
	h := newFlowHarness(t)
	h.idp.claims = map[string]interface{}{"email": "a@example.com", "email_verified": "true"}
	st, echoed := h.start(t, ModeLogin, 0)
	claims, err := h.complete(st, "good-code", echoed)
	if err != nil || !claims.EmailVerified {
		t.Fatalf("claims=%+v err=%v", claims, err)
	}

	h.idp.claims = map[string]interface{}{"email": "a@example.com"}
	st, echoed = h.start(t, ModeLogin, 0)
	claims, err = h.complete(st, "good-code", echoed)
	if err != nil || claims.EmailVerified {
		t.Fatalf("missing email_verified must be false: claims=%+v err=%v", claims, err)
	}
}

func TestFlowRejections(t *testing.T) {
	cases := []struct {
		name  string
		setup func(h *flowHarness, st *FlowState, echoed *string, code *string)
	}{
		{"bad state", func(h *flowHarness, st *FlowState, echoed *string, code *string) { *echoed = "forged" }},
		{"missing state", func(h *flowHarness, st *FlowState, echoed *string, code *string) { *echoed = "" }},
		{"missing code", func(h *flowHarness, st *FlowState, echoed *string, code *string) { *code = "" }},
		{"invalid code", func(h *flowHarness, st *FlowState, echoed *string, code *string) { *code = "other-code" }},
		{"bad nonce", func(h *flowHarness, st *FlowState, echoed *string, code *string) { h.idp.nonce = "replayed" }},
		{"empty nonce in token", func(h *flowHarness, st *FlowState, echoed *string, code *string) { h.idp.nonce = "" }},
		{"expired token", func(h *flowHarness, st *FlowState, echoed *string, code *string) {
			h.idp.expiry = time.Now().Add(-time.Hour)
		}},
		{"wrong issuer", func(h *flowHarness, st *FlowState, echoed *string, code *string) {
			h.idp.issuer = "https://evil.example.com"
		}},
		{"wrong audience", func(h *flowHarness, st *FlowState, echoed *string, code *string) {
			h.idp.audience = "someone-else"
		}},
		{"pkce mismatch", func(h *flowHarness, st *FlowState, echoed *string, code *string) {
			h.idp.challenge = "not-the-challenge"
		}},
		{"no id_token", func(h *flowHarness, st *FlowState, echoed *string, code *string) { h.idp.omitIDToken = true }},
		{"empty subject", func(h *flowHarness, st *FlowState, echoed *string, code *string) { h.idp.subject = " " }},
		{"expired state", func(h *flowHarness, st *FlowState, echoed *string, code *string) {
			h.flow.now = func() time.Time { return time.Now().Add(stateTTL + time.Minute) }
		}},
		{"state for other provider", func(h *flowHarness, st *FlowState, echoed *string, code *string) { st.ProviderID = 99 }},
		{"unknown mode", func(h *flowHarness, st *FlowState, echoed *string, code *string) { st.Mode = "admin" }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newFlowHarness(t)
			st, echoed := h.start(t, ModeLogin, 0)
			code := "good-code"
			c.setup(h, &st, &echoed, &code)

			if _, err := h.complete(st, code, echoed); err == nil {
				t.Fatal("expected failure")
			}
		})
	}
}

func TestFlowSigningKeyFromAnotherIdP(t *testing.T) {
	h := newFlowHarness(t)
	st, echoed := h.start(t, ModeLogin, 0)

	other, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	h.idp.signer, err = jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: other, KeyID: "k1"}},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := h.complete(st, "good-code", echoed); err == nil {
		t.Fatal("token signed with an unknown key must be rejected")
	}
}

func TestFlowLinkStateBindsUser(t *testing.T) {
	h := newFlowHarness(t)
	st, _ := h.start(t, ModeLink, 42)
	if st.Mode != ModeLink || st.UserID != 42 {
		t.Fatalf("state: %+v", st)
	}
}

func TestBeginLinkRefusesAdmin(t *testing.T) {
	h := newFlowHarness(t)
	_, fe := h.flow.BeginLink(context.Background(), "fake", tables.User{ID: 1, Username: "admin"})
	if fe == nil || fe.Code != CodeNotPermitted {
		t.Fatalf("expected not-permitted, got %v", fe)
	}
}

func TestStateTokenTamperAndForeignKey(t *testing.T) {
	h := newFlowHarness(t)
	token, err := h.flow.EncodeState(FlowState{State: "s", Nonce: "n", Mode: ModeLogin, ProviderID: 3})
	if err != nil {
		t.Fatal(err)
	}

	got, err := h.flow.DecodeState(token)
	if err != nil || got.State != "s" {
		t.Fatalf("round trip: %+v %v", got, err)
	}

	raw, _ := base64.StdEncoding.DecodeString(token)
	raw[len(raw)-1] ^= 0x01
	if _, err := h.flow.DecodeState(base64.StdEncoding.EncodeToString(raw)); err == nil {
		t.Fatal("tampered state must fail")
	}

	other := NewFlowService(nil, NewProviderService(nil, []byte("another-private-secret"), ""), []byte("another-private-secret"))
	if _, err := other.DecodeState(token); err == nil {
		t.Fatal("state from another secret must fail")
	}

	// a stored client secret must not be usable as a state cookie
	providers := NewProviderService(nil, []byte(testSecret), "")
	enc, _ := svcEncrypt(providers, `{"s":"x","n":"y","m":"login","p":3,"e":9999999999}`)
	if _, err := h.flow.DecodeState(enc); err == nil {
		t.Fatal("client-secret ciphertext must not decode as state")
	}
}

func TestHandleCallbackWithBadStateCookie(t *testing.T) {
	h := newFlowHarness(t)
	res := h.flow.HandleCallback(context.Background(), CallbackInput{
		Slug: "fake", StateToken: "garbage", Code: "good-code", State: "x",
	})
	if res.ErrorCode != CodeFailed || res.Mode != ModeLogin {
		t.Fatalf("result: %+v", res)
	}
	if res.Redirect != "/auth/login?error=oidc_failed" {
		t.Fatalf("redirect: %q", res.Redirect)
	}
}

func TestHandleCallbackProviderErrorNeverEchoed(t *testing.T) {
	h := newFlowHarness(t)
	token, _ := h.flow.EncodeState(FlowState{State: "s", Mode: ModeLink, ProviderID: 3, ExpiresAt: time.Now().Add(time.Minute).Unix()})

	res := h.flow.HandleCallback(context.Background(), CallbackInput{
		Slug: "fake", StateToken: token, ProviderError: true,
	})
	if res.ErrorCode != CodeFailed || res.Mode != ModeLink {
		t.Fatalf("result: %+v", res)
	}
	if res.Redirect != "/user/profile?oidcError=oidc_failed" {
		t.Fatalf("redirect: %q", res.Redirect)
	}
}

func TestFailureRedirect(t *testing.T) {
	if got := FailureRedirect(ModeLogin, CodeNotPermitted); got != "/auth/login?error=oidc_not_permitted" {
		t.Errorf("login: %q", got)
	}
	if got := FailureRedirect(ModeLink, CodeAlreadyLinked); got != "/user/profile?oidcError=oidc_already_linked" {
		t.Errorf("link: %q", got)
	}
}
