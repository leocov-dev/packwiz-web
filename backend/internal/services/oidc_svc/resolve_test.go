package oidc_svc

import "testing"

func facts(id uint) *UserFacts {
	return &UserFacts{ID: id, Active: true}
}

func TestDecideLogin(t *testing.T) {
	verified := Claims{Subject: "sub", Email: "a@example.com", EmailVerified: true}
	unverified := Claims{Subject: "sub", Email: "a@example.com", EmailVerified: false}
	noEmail := Claims{Subject: "sub"}

	admin := &UserFacts{ID: 1, Active: true, IsAdmin: true}
	inactive := &UserFacts{ID: 2, Active: false}
	linked := &UserFacts{ID: 3, Active: true, HasLinkAtProvider: true}

	cases := []struct {
		name string
		in   ResolveInput
		want Decision
	}{
		{
			name: "existing identity logs in",
			in:   ResolveInput{Mode: ModeLogin, Claims: verified, Identity: facts(10)},
			want: Decision{Kind: OutcomeLoggedIn, UserID: 10},
		},
		{
			name: "existing identity wins over email match",
			in: ResolveInput{
				Mode: ModeLogin, Claims: verified, Identity: facts(10), EmailUser: facts(11),
				Policy: Policy{LinkByEmail: true, AutoCreateUsers: true},
			},
			want: Decision{Kind: OutcomeLoggedIn, UserID: 10},
		},
		{
			name: "identity of inactive user rejected",
			in:   ResolveInput{Mode: ModeLogin, Claims: verified, Identity: inactive},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonUserInactive},
		},
		{
			name: "identity of admin rejected",
			in:   ResolveInput{Mode: ModeLogin, Claims: verified, Identity: admin},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonProtectedUser},
		},
		{
			name: "unknown user, nothing enabled",
			in:   ResolveInput{Mode: ModeLogin, Claims: verified},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonNoAccount},
		},
		{
			name: "email match ignored when link_by_email off",
			in:   ResolveInput{Mode: ModeLogin, Claims: verified, EmailUser: facts(20)},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonNoAccount},
		},
		{
			name: "verified email links",
			in: ResolveInput{
				Mode: ModeLogin, Claims: verified, EmailUser: facts(20),
				Policy: Policy{LinkByEmail: true},
			},
			want: Decision{Kind: OutcomeLinked, UserID: 20},
		},
		{
			name: "unverified email never links",
			in: ResolveInput{
				Mode: ModeLogin, Claims: unverified, EmailUser: facts(20),
				Policy: Policy{LinkByEmail: true},
			},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonEmailNotVerified},
		},
		{
			name: "unverified email does not fall through to auto-create",
			in: ResolveInput{
				Mode: ModeLogin, Claims: unverified, EmailUser: facts(20),
				Policy: Policy{LinkByEmail: true, AutoCreateUsers: true},
			},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonEmailNotVerified},
		},
		{
			name: "email link never targets admin",
			in: ResolveInput{
				Mode: ModeLogin, Claims: verified, EmailUser: admin,
				Policy: Policy{LinkByEmail: true},
			},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonProtectedUser},
		},
		{
			name: "email link never targets inactive user",
			in: ResolveInput{
				Mode: ModeLogin, Claims: verified, EmailUser: inactive,
				Policy: Policy{LinkByEmail: true},
			},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonUserInactive},
		},
		{
			name: "email link refused when user already has identity at provider",
			in: ResolveInput{
				Mode: ModeLogin, Claims: verified, EmailUser: linked,
				Policy: Policy{LinkByEmail: true},
			},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonProviderLinked},
		},
		{
			name: "link_by_email on but no email user falls to auto-create",
			in: ResolveInput{
				Mode: ModeLogin, Claims: verified,
				Policy: Policy{LinkByEmail: true, AutoCreateUsers: true},
			},
			want: Decision{Kind: OutcomeCreated},
		},
		{
			name: "auto-create",
			in: ResolveInput{
				Mode: ModeLogin, Claims: verified,
				Policy: Policy{AutoCreateUsers: true},
			},
			want: Decision{Kind: OutcomeCreated},
		},
		{
			name: "auto-create needs an email",
			in: ResolveInput{
				Mode: ModeLogin, Claims: noEmail,
				Policy: Policy{AutoCreateUsers: true},
			},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonEmailMissing},
		},
		{
			name: "auto-create needs a verified email",
			in: ResolveInput{
				Mode: ModeLogin, Claims: Claims{Subject: "sub", Email: "a@example.com"},
				Policy: Policy{LinkByEmail: true, AutoCreateUsers: true},
			},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonEmailNotVerified},
		},
		{
			name: "auto-create refused when email belongs to another user",
			in: ResolveInput{
				Mode: ModeLogin, Claims: verified, EmailUser: facts(20),
				Policy: Policy{AutoCreateUsers: true},
			},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonEmailInUse},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Decide(c.in); got != c.want {
				t.Fatalf("got %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestDecideLink(t *testing.T) {
	claims := Claims{Subject: "sub", Email: "a@example.com", EmailVerified: true}

	cases := []struct {
		name string
		in   ResolveInput
		want Decision
	}{
		{
			name: "links to the user who started the flow",
			in:   ResolveInput{Mode: ModeLink, Claims: claims, LinkTarget: facts(5)},
			want: Decision{Kind: OutcomeLinked, UserID: 5},
		},
		{
			name: "ignores email matches and policy toggles",
			in: ResolveInput{
				Mode: ModeLink, Claims: claims, LinkTarget: facts(5), EmailUser: facts(6),
				Policy: Policy{LinkByEmail: true, AutoCreateUsers: true},
			},
			want: Decision{Kind: OutcomeLinked, UserID: 5},
		},
		{
			name: "missing target",
			in:   ResolveInput{Mode: ModeLink, Claims: claims},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonLinkTargetMissing},
		},
		{
			name: "admin cannot link",
			in: ResolveInput{
				Mode: ModeLink, Claims: claims,
				LinkTarget: &UserFacts{ID: 1, Active: true, IsAdmin: true},
			},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonProtectedUser},
		},
		{
			name: "inactive target",
			in: ResolveInput{
				Mode: ModeLink, Claims: claims,
				LinkTarget: &UserFacts{ID: 5, Active: false},
			},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonUserInactive},
		},
		{
			name: "identity already linked to another user",
			in:   ResolveInput{Mode: ModeLink, Claims: claims, LinkTarget: facts(5), Identity: facts(9)},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonIdentityElsewhere},
		},
		{
			name: "identity already linked to this user is a no-op",
			in:   ResolveInput{Mode: ModeLink, Claims: claims, LinkTarget: facts(5), Identity: facts(5)},
			want: Decision{Kind: OutcomeLoggedIn, UserID: 5},
		},
		{
			name: "user already has another identity at the provider",
			in: ResolveInput{
				Mode: ModeLink, Claims: claims,
				LinkTarget: &UserFacts{ID: 5, Active: true, HasLinkAtProvider: true},
			},
			want: Decision{Kind: OutcomeRejected, Reason: ReasonProviderLinked},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Decide(c.in); got != c.want {
				t.Fatalf("got %+v, want %+v", got, c.want)
			}
		})
	}
}
