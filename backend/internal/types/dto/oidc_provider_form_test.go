package dto

import "testing"

func TestValidateOidcSlug(t *testing.T) {
	cases := map[string]bool{
		"keycloak":    true,
		"my-idp-2":    true,
		"a":           true,
		"":            false,
		"-lead":       false,
		"trail-":      false,
		"UPPER":       false,
		"has space":   false,
		"has/slash":   false,
		"under_score": false,
		"../etc":      false,
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": false, // 65 chars
	}
	for slug, ok := range cases {
		err := ValidateOidcSlug(slug)
		if (err == nil) != ok {
			t.Errorf("slug %q: valid=%v, err=%v", slug, ok, err)
		}
	}
}

func TestValidateOidcIssuerURL(t *testing.T) {
	cases := map[string]bool{
		"https://accounts.google.com":      true,
		"https://kc.example.com/realms/x":  true,
		"http://localhost:8080/realms/x":   true,
		"http://127.0.0.1:9000":            true,
		"http://kc.example.com":            false,
		"ftp://kc.example.com":             false,
		"kc.example.com":                   false,
		"":                                 false,
		"https://kc.example.com/?a=b":      false,
		"https://kc.example.com/#fragment": false,
	}
	for raw, ok := range cases {
		err := ValidateOidcIssuerURL(raw)
		if (err == nil) != ok {
			t.Errorf("issuer %q: valid=%v, err=%v", raw, ok, err)
		}
	}
}

func TestNormalizeOidcScopes(t *testing.T) {
	got, err := NormalizeOidcScopes("")
	if err != nil || got != DefaultOidcScopes {
		t.Fatalf("blank: got %q, %v", got, err)
	}

	got, err = NormalizeOidcScopes("  openid  email email groups ")
	if err != nil || got != "openid email groups" {
		t.Fatalf("dedupe: got %q, %v", got, err)
	}

	if _, err = NormalizeOidcScopes("profile email"); err == nil {
		t.Fatal("expected error when openid scope is missing")
	}
}

func TestOidcProviderFormValidate(t *testing.T) {
	valid := func() OidcProviderForm {
		return OidcProviderForm{
			Slug:        " keycloak ",
			DisplayName: "Keycloak",
			IssuerURL:   "https://kc.example.com/realms/x/",
			ClientID:    "packwiz",
		}
	}

	f := valid()
	if err := f.Validate(); err != nil {
		t.Fatalf("valid form rejected: %v", err)
	}
	if f.Slug != "keycloak" || f.IssuerURL != "https://kc.example.com/realms/x" || f.Scopes != DefaultOidcScopes {
		t.Fatalf("form not normalized: %+v", f)
	}

	for name, mutate := range map[string]func(*OidcProviderForm){
		"missing name":   func(f *OidcProviderForm) { f.DisplayName = "" },
		"missing client": func(f *OidcProviderForm) { f.ClientID = "" },
		"bad slug":       func(f *OidcProviderForm) { f.Slug = "Bad Slug" },
		"http issuer":    func(f *OidcProviderForm) { f.IssuerURL = "http://kc.example.com" },
		"no openid":      func(f *OidcProviderForm) { f.Scopes = "email" },
	} {
		f := valid()
		mutate(&f)
		if err := f.Validate(); err == nil {
			t.Errorf("%s: expected validation error", name)
		}
	}
}
