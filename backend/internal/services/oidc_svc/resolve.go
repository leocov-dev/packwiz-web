package oidc_svc

// OutcomeKind is the result of resolving an identity provider login.
type OutcomeKind string

const (
	// OutcomeLoggedIn: the identity was already linked to an active user.
	OutcomeLoggedIn OutcomeKind = "logged_in"
	// OutcomeLinked: a new identity was attached to an existing user (by
	// explicit link mode, or by verified email when the provider allows it).
	OutcomeLinked OutcomeKind = "linked"
	// OutcomeCreated: a new non-admin user was created.
	OutcomeCreated OutcomeKind = "created"
	// OutcomeRejected: the login is not permitted; see RejectReason.
	OutcomeRejected OutcomeKind = "rejected"
)

// RejectReason explains a rejection. It is for server-side logs and tests and
// is never shown to the browser.
type RejectReason string

const (
	ReasonNoAccount         RejectReason = "no matching account and auto-create is off"
	ReasonUserInactive      RejectReason = "user is deactivated"
	ReasonProtectedUser     RejectReason = "user is the protected admin account"
	ReasonEmailNotVerified  RejectReason = "email is not verified"
	ReasonEmailMissing      RejectReason = "identity provider returned no email"
	ReasonEmailInUse        RejectReason = "email already belongs to another user"
	ReasonProviderLinked    RejectReason = "user already has an identity at this provider"
	ReasonIdentityElsewhere RejectReason = "identity is linked to another user"
	ReasonLinkTargetMissing RejectReason = "link target user is missing"
)

// Claims are the verified identity attributes taken from the ID token.
type Claims struct {
	Subject           string
	Email             string
	EmailVerified     bool
	PreferredUsername string
	Name              string
}

// UserFacts is what the resolver needs to know about a candidate user.
type UserFacts struct {
	ID                uint
	Active            bool
	IsAdmin           bool // true for the protected built-in admin account
	HasLinkAtProvider bool
}

// Policy is the per-provider behavior toggles.
type Policy struct {
	AutoCreateUsers bool
	LinkByEmail     bool
}

// ResolveInput carries everything Decide needs; all lookups are done by the
// caller so Decide stays pure.
type ResolveInput struct {
	Mode   FlowMode
	Claims Claims
	Policy Policy

	// Identity is the user owning (provider, subject), if any.
	Identity *UserFacts
	// EmailUser is the user whose email matches the claims, if any.
	EmailUser *UserFacts
	// LinkTarget is the user who started a link flow (link mode only).
	LinkTarget *UserFacts
}

// Decision is the outcome of Decide.
type Decision struct {
	Kind   OutcomeKind
	UserID uint // set for LoggedIn and Linked
	Reason RejectReason
}

func reject(reason RejectReason) Decision {
	return Decision{Kind: OutcomeRejected, Reason: reason}
}

// Decide applies the resolution order: existing identity, then verified-email
// link (if the provider allows it), then auto-create (if the provider allows
// it), else reject. In link mode the identity is bound to the user who started
// the flow and nothing else is considered. The protected admin account is
// never logged into, linked or created through OIDC.
func Decide(in ResolveInput) Decision {
	if in.Mode == ModeLink {
		return decideLink(in)
	}

	if in.Identity != nil {
		switch {
		case in.Identity.IsAdmin:
			return reject(ReasonProtectedUser)
		case !in.Identity.Active:
			return reject(ReasonUserInactive)
		}
		return Decision{Kind: OutcomeLoggedIn, UserID: in.Identity.ID}
	}

	if in.Policy.LinkByEmail && in.EmailUser != nil {
		u := in.EmailUser
		switch {
		case in.Claims.Email == "":
			return reject(ReasonEmailMissing)
		case !in.Claims.EmailVerified:
			return reject(ReasonEmailNotVerified)
		case u.IsAdmin:
			return reject(ReasonProtectedUser)
		case !u.Active:
			return reject(ReasonUserInactive)
		case u.HasLinkAtProvider:
			return reject(ReasonProviderLinked)
		}
		return Decision{Kind: OutcomeLinked, UserID: u.ID}
	}

	if in.Policy.AutoCreateUsers {
		switch {
		case in.Claims.Email == "":
			return reject(ReasonEmailMissing)
		case in.EmailUser != nil:
			return reject(ReasonEmailInUse)
		}
		return Decision{Kind: OutcomeCreated}
	}

	return reject(ReasonNoAccount)
}

func decideLink(in ResolveInput) Decision {
	t := in.LinkTarget
	switch {
	case t == nil:
		return reject(ReasonLinkTargetMissing)
	case t.IsAdmin:
		return reject(ReasonProtectedUser)
	case !t.Active:
		return reject(ReasonUserInactive)
	}

	if in.Identity != nil {
		if in.Identity.ID != t.ID {
			return reject(ReasonIdentityElsewhere)
		}
		// already linked to this very user: nothing to do
		return Decision{Kind: OutcomeLoggedIn, UserID: t.ID}
	}

	if t.HasLinkAtProvider {
		return reject(ReasonProviderLinked)
	}
	return Decision{Kind: OutcomeLinked, UserID: t.ID}
}
