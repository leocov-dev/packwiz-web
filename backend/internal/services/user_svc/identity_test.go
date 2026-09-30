package user_svc

import "testing"

func TestDeriveUsernameBase(t *testing.T) {
	cases := []struct{ preferred, email, want string }{
		{"jane", "x@example.com", "jane"},
		{"Jane.Doe", "", "jane.doe"},
		{"", "Jane.Doe@example.com", "jane.doe"},
		{"  ", "bob+tag@example.com", "bobtag"},
		{"jane doe!", "", "janedoe"},
		{"admin", "a@example.com", "user"},
		{"ADMIN", "", "user"},
		{"", "admin@example.com", "user"},
		{"", "", "user"},
		{"!!!", "", "user"},
		{"üñí", "", "user"},
		{"abcdefghijklmnopqrstuvwxyz0123456789", "", "abcdefghijklmnopqrstuvwxyz012345"},
	}
	for _, c := range cases {
		if got := DeriveUsernameBase(c.preferred, c.email); got != c.want {
			t.Errorf("DeriveUsernameBase(%q, %q) = %q, want %q", c.preferred, c.email, got, c.want)
		}
	}
}

func TestCanUnlinkIdentity(t *testing.T) {
	cases := []struct {
		hasPassword bool
		others      int64
		want        bool
	}{
		{true, 0, true},
		{true, 2, true},
		{false, 1, true},
		{false, 0, false},
	}
	for _, c := range cases {
		if got := CanUnlinkIdentity(c.hasPassword, c.others); got != c.want {
			t.Errorf("CanUnlinkIdentity(%v, %d) = %v, want %v", c.hasPassword, c.others, got, c.want)
		}
	}
}
