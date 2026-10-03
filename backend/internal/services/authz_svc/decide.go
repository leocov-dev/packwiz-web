package authz_svc

import "errors"

var (
	// ErrForbidden means the user lacks the permission.
	ErrForbidden = errors.New("permission denied")
	// ErrArchived means the pack is archived and the permission is not allowed on archived packs.
	ErrArchived = errors.New("pack is archived")
	// ErrPackNotFound means the pack does not exist.
	ErrPackNotFound = errors.New("pack not found")
)

// Set is a set of permission names.
type Set map[string]struct{}

// NewSet builds a Set from names.
func NewSet(names ...string) Set {
	s := make(Set, len(names))
	for _, n := range names {
		s[n] = struct{}{}
	}
	return s
}

// Has reports whether the set holds the permission. Safe on a nil Set.
func (s Set) Has(name string) bool {
	_, ok := s[name]
	return ok
}

// Names returns the permission names in the set, unsorted.
func (s Set) Names() []string {
	out := make([]string, 0, len(s))
	for n := range s {
		out = append(out, n)
	}
	return out
}

// DecideGlobal is the decision for a global permission. The superuser
// short-circuits before the set is consulted.
func DecideGlobal(superuser bool, system Set, permission string) bool {
	return superuser || system.Has(permission)
}

// DecidePack is the decision for a pack permission. The archived-state rule
// runs first and applies to the superuser too; then the superuser
// short-circuit; then the system set (all packs) or the pack role set (this pack).
func DecidePack(archived, superuser bool, system, pack Set, permission string) error {
	if archived {
		if _, ok := archivedAllowed[permission]; !ok {
			return ErrArchived
		}
	}
	if superuser || system.Has(permission) || pack.Has(permission) {
		return nil
	}
	return ErrForbidden
}
