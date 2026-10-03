package oidc_svc

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"packwiz-web/internal/log"
	"packwiz-web/internal/services/user_svc"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/utils"
)

// FlowMode says what an in-flight OIDC round trip is for.
type FlowMode string

const (
	// ModeLogin signs a user in (and may link or create one, per provider policy).
	ModeLogin FlowMode = "login"
	// ModeLink attaches an identity to the already signed-in user who started it.
	ModeLink FlowMode = "link"
)

// Error codes surfaced to the browser. They are deliberately generic: IdP
// error text and internal detail never reach the client.
const (
	CodeFailed        = "oidc_failed"
	CodeNotPermitted  = "oidc_not_permitted"
	CodeUnavailable   = "oidc_unavailable"
	CodeAlreadyLinked = "oidc_already_linked"
)

const (
	loginPath   = "/auth/login"
	profilePath = "/user/profile"

	// stateTTL bounds how long a started flow may take to complete.
	stateTTL = 10 * time.Minute
	// StateCookieMaxAge is the lifetime of the state cookie in seconds.
	StateCookieMaxAge = int(stateTTL / time.Second)
)

// FlowError is a failure carrying one of the generic Code* values.
type FlowError struct {
	Code string
	Err  error
}

func (e *FlowError) Error() string {
	if e.Err == nil {
		return e.Code
	}
	return fmt.Sprintf("%s: %v", e.Code, e.Err)
}

func (e *FlowError) Unwrap() error { return e.Err }

// Message is a safe, human-readable description for API responses.
func (e *FlowError) Message() string {
	switch e.Code {
	case CodeUnavailable:
		return "this sign-in provider is not available"
	case CodeNotPermitted:
		return "this account is not permitted to use that provider"
	case CodeAlreadyLinked:
		return "that account is already linked"
	default:
		return "sign-in with the provider failed"
	}
}

func flowErr(code string, err error) *FlowError { return &FlowError{Code: code, Err: err} }

// FlowState is the server-side record of one in-flight OIDC round trip. It is
// handed to the browser only as an encrypted, authenticated cookie value.
type FlowState struct {
	State      string   `json:"s"`
	Nonce      string   `json:"n"`
	Verifier   string   `json:"v"`
	Mode       FlowMode `json:"m"`
	UserID     uint     `json:"u,omitempty"`
	ProviderID uint     `json:"p"`
	Redirect   string   `json:"r,omitempty"`
	ExpiresAt  int64    `json:"e"`
}

// FlowService runs the OIDC authorization-code flow (with PKCE and nonce) and
// resolves the outcome to a local user.
type FlowService struct {
	providers *ProviderService
	users     *user_svc.UserService
	stateKey  []byte
	now       func() time.Time
}

// NewFlowService builds a FlowService. The state cookie is encrypted with a
// key derived from sessionSecret, domain-separated from client secrets.
func NewFlowService(db *gorm.DB, providers *ProviderService, sessionSecret []byte) *FlowService {
	return &FlowService{
		providers: providers,
		users:     user_svc.NewUserService(db),
		stateKey:  append([]byte("oidc-state:"), sessionSecret...),
		now:       time.Now,
	}
}

// EncodeState seals a FlowState into a cookie-safe string.
func (f *FlowService) EncodeState(st FlowState) (string, error) {
	raw, err := json.Marshal(st)
	if err != nil {
		return "", fmt.Errorf("encode oidc state: %w", err)
	}
	return utils.SecretboxEncrypt(f.stateKey, string(raw))
}

// DecodeState opens a sealed FlowState. Tampered or foreign values fail.
func (f *FlowService) DecodeState(token string) (FlowState, error) {
	var st FlowState
	plain, err := utils.SecretboxDecrypt(f.stateKey, token)
	if err != nil {
		return st, err
	}
	if err := json.Unmarshal([]byte(plain), &st); err != nil {
		return FlowState{}, fmt.Errorf("decode oidc state: %w", err)
	}
	return st, nil
}

// BeginResult is what the controller needs to start a round trip: where to
// send the browser, and the sealed state to store in the state cookie.
type BeginResult struct {
	AuthURL    string
	StateToken string
}

// BeginLogin starts a sign-in round trip with the provider. redirect is the
// post-login path and is sanitized to a same-origin relative path.
func (f *FlowService) BeginLogin(ctx context.Context, slug, redirect string) (BeginResult, *FlowError) {
	return f.begin(ctx, slug, ModeLogin, 0, SanitizeRedirect(redirect))
}

// BeginLink starts a round trip that links a provider account to user. The
// user id is sealed into the state, so the callback never has to read the
// session (whose Strict cookie is not sent on the cross-site return).
func (f *FlowService) BeginLink(ctx context.Context, slug string, user tables.User) (BeginResult, *FlowError) {
	if user.Username == "admin" {
		return BeginResult{}, flowErr(CodeNotPermitted, errors.New("admin cannot link identities"))
	}
	return f.begin(ctx, slug, ModeLink, user.ID, profilePath)
}

func (f *FlowService) begin(ctx context.Context, slug string, mode FlowMode, userID uint, redirect string) (BeginResult, *FlowError) {
	p, serr := f.providers.GetBySlug(slug)
	if serr != nil {
		return BeginResult{}, flowErr(CodeUnavailable, serr)
	}
	return f.beginWith(ctx, p, mode, userID, redirect)
}

func (f *FlowService) beginWith(ctx context.Context, p tables.OidcProvider, mode FlowMode, userID uint, redirect string) (BeginResult, *FlowError) {
	cl, err := f.providers.ClientFor(ctx, p)
	if err != nil {
		log.Warn("oidc begin for provider ", p.Slug, " failed: ", err)
		return BeginResult{}, flowErr(CodeUnavailable, err)
	}

	st := FlowState{
		State:      utils.GenerateRandomString(32),
		Nonce:      utils.GenerateRandomString(32),
		Verifier:   oauth2.GenerateVerifier(),
		Mode:       mode,
		UserID:     userID,
		ProviderID: p.ID,
		Redirect:   redirect,
		ExpiresAt:  f.now().Add(stateTTL).Unix(),
	}

	token, err := f.EncodeState(st)
	if err != nil {
		return BeginResult{}, flowErr(CodeFailed, err)
	}

	authURL := cl.OAuth.AuthCodeURL(
		st.State,
		oauth2.S256ChallengeOption(st.Verifier),
		oidc.Nonce(st.Nonce),
	)
	return BeginResult{AuthURL: authURL, StateToken: token}, nil
}

// verifyState checks the state echoed back by the IdP against the sealed one.
func verifyState(st FlowState, echoed string, providerID uint, now time.Time) error {
	if st.State == "" || echoed == "" {
		return errors.New("missing state")
	}
	if subtle.ConstantTimeCompare([]byte(st.State), []byte(echoed)) != 1 {
		return errors.New("state mismatch")
	}
	if st.ProviderID != providerID {
		return errors.New("state was issued for a different provider")
	}
	if now.Unix() > st.ExpiresAt {
		return errors.New("state expired")
	}
	if st.Mode != ModeLogin && st.Mode != ModeLink {
		return errors.New("unknown flow mode")
	}
	return nil
}

// idTokenClaims are the ID token claims we read. email_verified is decoded
// leniently because some IdPs send it as a string.
type idTokenClaims struct {
	Subject           string          `json:"sub"`
	Email             string          `json:"email"`
	EmailVerified     json.RawMessage `json:"email_verified"`
	PreferredUsername string          `json:"preferred_username"`
	Name              string          `json:"name"`
}

func parseBoolClaim(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		return b
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.EqualFold(s, "true")
	}
	return false
}

// complete exchanges the authorization code (with the PKCE verifier) and
// verifies the ID token: signature, issuer, audience, expiry and nonce.
func (f *FlowService) complete(ctx context.Context, cl *Client, st FlowState, code, echoedState string) (Claims, error) {
	if err := verifyState(st, echoedState, cl.ProviderID, f.now()); err != nil {
		return Claims{}, err
	}
	if code == "" {
		return Claims{}, errors.New("missing authorization code")
	}

	ectx, cancel := context.WithTimeout(httpContext(ctx), httpTimeout)
	defer cancel()

	token, err := cl.OAuth.Exchange(ectx, code, oauth2.VerifierOption(st.Verifier))
	if err != nil {
		return Claims{}, fmt.Errorf("code exchange: %w", err)
	}

	rawIDToken, _ := token.Extra("id_token").(string)
	if rawIDToken == "" {
		return Claims{}, errors.New("token response has no id_token")
	}

	idToken, err := cl.Verifier().Verify(ectx, rawIDToken)
	if err != nil {
		return Claims{}, fmt.Errorf("id token verification: %w", err)
	}

	if subtle.ConstantTimeCompare([]byte(idToken.Nonce), []byte(st.Nonce)) != 1 || st.Nonce == "" {
		return Claims{}, errors.New("nonce mismatch")
	}

	var raw idTokenClaims
	if err := idToken.Claims(&raw); err != nil {
		return Claims{}, fmt.Errorf("decode id token claims: %w", err)
	}
	if strings.TrimSpace(raw.Subject) == "" {
		return Claims{}, errors.New("id token has no subject")
	}

	return Claims{
		Subject:           raw.Subject,
		Email:             strings.ToLower(strings.TrimSpace(raw.Email)),
		EmailVerified:     parseBoolClaim(raw.EmailVerified),
		PreferredUsername: strings.TrimSpace(raw.PreferredUsername),
		Name:              strings.TrimSpace(raw.Name),
	}, nil
}

// CallbackInput is what the IdP redirect delivered, plus the sealed state from
// the cookie.
type CallbackInput struct {
	Slug       string
	StateToken string
	Code       string
	State      string
	// ProviderError is true when the IdP redirected back with an error
	// parameter. Its text is never used.
	ProviderError bool
}

// CallbackResult is the resolved outcome of a callback. When ErrorCode is
// empty the flow succeeded: for ModeLogin the caller must start a session for
// User; Redirect is where to send the browser next.
type CallbackResult struct {
	Mode      FlowMode
	Kind      OutcomeKind
	User      tables.User
	Redirect  string
	ErrorCode string
}

// FailureRedirect is the page a browser is sent to after a failed flow.
func FailureRedirect(mode FlowMode, code string) string {
	q := url.Values{}
	if mode == ModeLink {
		q.Set("oidcError", code)
		return profilePath + "?" + q.Encode()
	}
	q.Set("error", code)
	return loginPath + "?" + q.Encode()
}

func failed(mode FlowMode, code string) CallbackResult {
	return CallbackResult{Mode: mode, Kind: OutcomeRejected, ErrorCode: code, Redirect: FailureRedirect(mode, code)}
}

// HandleCallback finishes a round trip: it validates the sealed state, runs
// the code exchange and token verification, then resolves the identity to a
// user according to the provider's policy. It never returns IdP error text.
func (f *FlowService) HandleCallback(ctx context.Context, in CallbackInput) CallbackResult {
	st, err := f.DecodeState(in.StateToken)
	if err != nil {
		log.Warn("oidc callback for ", in.Slug, ": invalid state cookie: ", err)
		return failed(ModeLogin, CodeFailed)
	}
	mode := st.Mode
	if mode != ModeLink {
		mode = ModeLogin
	}

	if in.ProviderError {
		log.Warn("oidc callback for ", in.Slug, ": identity provider reported an error")
		return failed(mode, CodeFailed)
	}

	p, serr := f.providers.GetBySlug(in.Slug)
	if serr != nil {
		return failed(mode, CodeUnavailable)
	}

	cl, err := f.providers.ClientFor(ctx, p)
	if err != nil {
		log.Warn("oidc callback for ", in.Slug, ": provider unavailable: ", err)
		return failed(mode, CodeUnavailable)
	}

	claims, err := f.complete(ctx, cl, st, in.Code, in.State)
	if err != nil {
		log.Warn("oidc callback for ", in.Slug, " failed: ", err)
		return failed(mode, CodeFailed)
	}

	return f.resolve(p, st, claims)
}

func toFacts(u *tables.User, hasLink bool) *UserFacts {
	if u == nil {
		return nil
	}
	return &UserFacts{ID: u.ID, Active: u.IsActive, IsSuperuser: u.IsSuperuser, HasLinkAtProvider: hasLink}
}

// resolve loads the facts Decide needs, applies the decision and performs the
// resulting database writes.
func (f *FlowService) resolve(p tables.OidcProvider, st FlowState, claims Claims) CallbackResult {
	mode := st.Mode
	in := ResolveInput{
		Mode:   mode,
		Claims: claims,
		Policy: Policy{AutoCreateUsers: p.AutoCreateUsers, LinkByEmail: p.LinkByEmail},
	}

	internal := func(msg string, err error) CallbackResult {
		log.Error("oidc resolve for ", p.Slug, ": ", msg, ": ", err)
		return failed(mode, CodeFailed)
	}

	identity, err := f.users.FindIdentity(p.ID, claims.Subject)
	if err != nil {
		return internal("identity lookup", err)
	}
	var identityUser *tables.User
	if identity != nil {
		u, err := f.users.FindById(identity.UserID)
		if err != nil {
			return internal("identity user lookup", err)
		}
		identityUser = &u
		in.Identity = toFacts(identityUser, false)
	}

	var emailUser *tables.User
	if mode == ModeLogin && claims.Email != "" {
		emailUser, err = f.users.FindByEmail(claims.Email)
		if err != nil {
			return internal("email lookup", err)
		}
		if emailUser != nil {
			has, err := f.users.HasIdentityAtProvider(emailUser.ID, p.ID)
			if err != nil {
				return internal("email user identity lookup", err)
			}
			in.EmailUser = toFacts(emailUser, has)
		}
	}

	var linkTarget *tables.User
	if mode == ModeLink {
		u, err := f.users.FindById(st.UserID)
		if err == nil {
			linkTarget = &u
			has, err := f.users.HasIdentityAtProvider(u.ID, p.ID)
			if err != nil {
				return internal("link target identity lookup", err)
			}
			in.LinkTarget = toFacts(linkTarget, has)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return internal("link target lookup", err)
		}
	}

	decision := Decide(in)
	if decision.Kind == OutcomeRejected {
		log.Warn("oidc login for provider ", p.Slug, " rejected: ", string(decision.Reason))
		code := CodeNotPermitted
		if mode == ModeLink && (decision.Reason == ReasonProviderLinked || decision.Reason == ReasonIdentityElsewhere) {
			code = CodeAlreadyLinked
		}
		return failed(mode, code)
	}

	ok := func(kind OutcomeKind, user tables.User) CallbackResult {
		redirect := SanitizeRedirect(st.Redirect)
		if mode == ModeLink {
			redirect = profilePath + "?oidc=linked"
		}
		return CallbackResult{Mode: mode, Kind: kind, User: user, Redirect: redirect}
	}

	switch {
	case mode == ModeLink:
		if decision.Kind == OutcomeLinked {
			if err := f.users.LinkIdentity(linkTarget.ID, p.ID, claims.Subject, claims.Email); err != nil {
				return internal("link identity", err)
			}
		}
		return ok(decision.Kind, *linkTarget)

	case decision.Kind == OutcomeLoggedIn:
		if err := f.users.TouchIdentity(*identity, claims.Email); err != nil {
			// best effort: the login itself must not fail on a bookkeeping write
			log.Warn("oidc: failed to record login time: ", err)
		}
		return ok(OutcomeLoggedIn, *identityUser)

	case decision.Kind == OutcomeLinked:
		if err := f.users.LinkIdentity(emailUser.ID, p.ID, claims.Subject, claims.Email); err != nil {
			return internal("link identity by email", err)
		}
		return ok(OutcomeLinked, *emailUser)

	case decision.Kind == OutcomeCreated:
		user, err := f.users.CreateExternalUser(user_svc.ExternalUserInput{
			ProviderID:        p.ID,
			Subject:           claims.Subject,
			Email:             claims.Email,
			PreferredUsername: claims.PreferredUsername,
			FullName:          claims.Name,
		})
		if err != nil {
			return internal("create user", err)
		}
		return ok(OutcomeCreated, user)
	}

	return failed(mode, CodeFailed)
}
