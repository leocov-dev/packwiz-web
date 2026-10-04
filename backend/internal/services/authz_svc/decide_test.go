package authz_svc

import (
	"errors"
	"testing"
)

func TestDecideGlobal(t *testing.T) {
	tests := []struct {
		name      string
		superuser bool
		system    Set
		want      bool
	}{
		{"superuser with no roles", true, nil, true},
		{"role holds permission", false, NewSet(PackCreate), true},
		{"role lacks permission", false, NewSet(UserLookup), false},
		{"no roles", false, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecideGlobal(tt.superuser, tt.system, PackCreate); got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestDecidePack(t *testing.T) {
	tests := []struct {
		name       string
		archived   bool
		superuser  bool
		system     Set
		pack       Set
		permission string
		want       error
	}{
		{"pack role grants", false, false, nil, NewSet(PackView), PackView, nil},
		{"system role grants on any pack", false, false, NewSet(PackModAdd), nil, PackModAdd, nil},
		{"no grant", false, false, nil, NewSet(PackView), PackModAdd, ErrForbidden},
		{"superuser needs nothing", false, true, nil, nil, PackPublish, nil},
		{"archived blocks write for pack role", true, false, nil, NewSet(PackModAdd), PackModAdd, ErrArchived},
		{"archived blocks write for superuser", true, true, nil, nil, PackModAdd, ErrArchived},
		{"archived allows view", true, false, nil, NewSet(PackView), PackView, nil},
		{"archived allows unarchive for superuser", true, true, nil, nil, PackArchive, nil},
		{"archived blocks consume", true, false, nil, NewSet(PackConsume), PackConsume, ErrArchived},
		{"archived blocks consume for superuser", true, true, nil, nil, PackConsume, ErrArchived},
		{"archived blocks link", true, false, nil, NewSet(PackLink), PackLink, ErrArchived},
		{"archived still needs the grant", true, false, nil, nil, PackView, ErrForbidden},
		{"archived allows snapshot view", true, false, nil, NewSet(PackSnapshotView), PackSnapshotView, nil},
		{"archived blocks revert for superuser", true, true, nil, nil, PackSnapshotRevert, ErrArchived},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecidePack(tt.archived, tt.superuser, tt.system, tt.pack, tt.permission)
			if !errors.Is(got, tt.want) {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestArchivedAllowedMatchesPlan(t *testing.T) {
	want := []string{PackView, PackUsersView, PackArchive, PackSnapshotView}
	if len(archivedAllowed) != len(want) {
		t.Fatalf("archivedAllowed has %d entries, want %d", len(archivedAllowed), len(want))
	}
	for _, w := range want {
		if _, ok := archivedAllowed[w]; !ok {
			t.Errorf("%s missing from archivedAllowed", w)
		}
	}
}
