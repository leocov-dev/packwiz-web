package authz_svc

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"packwiz-web/internal/tables"
)

// Service answers every access question. All checks flow through
// permission <- role_permissions <- role.
type Service struct {
	db *gorm.DB
}

// NewService creates the authorization service.
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// SystemPermissions loads the permissions a user holds through system roles.
// The superuser gets every permission without touching the database.
func (s *Service) SystemPermissions(user tables.User) (Set, error) {
	if user.IsSuperuser {
		return NewSet(AllPermissions...), nil
	}

	var names []string
	if err := s.db.Table("user_roles").
		Distinct("permissions.name").
		Joins("JOIN role_permissions ON role_permissions.role_id = user_roles.role_id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("user_roles.user_id = ?", user.ID).
		Scan(&names).Error; err != nil {
		return nil, fmt.Errorf("load system permissions: %w", err)
	}

	return NewSet(names...), nil
}

// Can reports whether the user holds a global permission. system is the set
// from SystemPermissions.
func (s *Service) Can(user tables.User, system Set, permission string) bool {
	return DecideGlobal(user.IsSuperuser, system, permission)
}

// PackRole is a user's role on one pack and the permissions it grants.
type PackRole struct {
	Name        string
	Permissions Set
}

// PackRole loads the user's pack-scope role on a pack. A nil result means no grant.
func (s *Service) PackRole(userID, packID uint) (*PackRole, error) {
	var rows []struct {
		RoleName string
		Perm     string
	}
	if err := s.db.Table("pack_users").
		Select("roles.name as role_name, permissions.name as perm").
		Joins("JOIN roles ON roles.id = pack_users.role_id").
		Joins("LEFT JOIN role_permissions ON role_permissions.role_id = roles.id").
		Joins("LEFT JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("pack_users.pack_id = ? AND pack_users.user_id = ?", packID, userID).
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("load pack role: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}

	role := &PackRole{Name: rows[0].RoleName, Permissions: Set{}}
	for _, r := range rows {
		if r.Perm != "" {
			role.Permissions[r.Perm] = struct{}{}
		}
	}
	return role, nil
}

// PackArchived reports whether the pack is archived (soft deleted). It returns
// ErrPackNotFound when the pack does not exist.
func (s *Service) PackArchived(packID uint) (bool, error) {
	var rows []struct{ Archived bool }
	if err := s.db.Unscoped().Table("packs").
		Select("deleted_at IS NOT NULL as archived").
		Where("id = ?", packID).
		Limit(1).
		Scan(&rows).Error; err != nil {
		return false, fmt.Errorf("load pack state: %w", err)
	}
	if len(rows) == 0 {
		return false, ErrPackNotFound
	}
	return rows[0].Archived, nil
}

// CanOnPack returns nil when the user may use the permission on the pack, else
// ErrArchived, ErrForbidden or ErrPackNotFound.
func (s *Service) CanOnPack(user tables.User, system Set, packID uint, permission string) error {
	archived, err := s.PackArchived(packID)
	if err != nil {
		return err
	}

	// the superuser needs no pack role lookup
	var packSet Set
	if !user.IsSuperuser && !system.Has(permission) {
		role, err := s.PackRole(user.ID, packID)
		if err != nil {
			return err
		}
		if role != nil {
			packSet = role.Permissions
		}
	}

	return DecidePack(archived, user.IsSuperuser, system, packSet, permission)
}

// EffectivePackPermissions returns what the user can do on each of the given
// packs, keyed by pack id, plus the pack role name for display. archived says
// which packs are archived; they are reduced to the archived-allowed set.
func (s *Service) EffectivePackPermissions(
	user tables.User,
	system Set,
	archived map[uint]bool,
) (perms map[uint][]string, roleNames map[uint]string, err error) {
	perms = make(map[uint][]string, len(archived))
	roleNames = make(map[uint]string, len(archived))

	var rows []struct {
		PackID   uint
		RoleName string
		Perm     string
	}
	if !user.IsSuperuser {
		if err := s.db.Table("pack_users").
			Select("pack_users.pack_id, roles.name as role_name, permissions.name as perm").
			Joins("JOIN roles ON roles.id = pack_users.role_id").
			Joins("LEFT JOIN role_permissions ON role_permissions.role_id = roles.id").
			Joins("LEFT JOIN permissions ON permissions.id = role_permissions.permission_id").
			Where("pack_users.user_id = ?", user.ID).
			Scan(&rows).Error; err != nil {
			return nil, nil, fmt.Errorf("load pack roles: %w", err)
		}
	}

	byPack := map[uint]Set{}
	for _, r := range rows {
		roleNames[r.PackID] = r.RoleName
		if byPack[r.PackID] == nil {
			byPack[r.PackID] = Set{}
		}
		if r.Perm != "" {
			byPack[r.PackID][r.Perm] = struct{}{}
		}
	}

	for packID, isArchived := range archived {
		out := []string{}
		for _, name := range AllPermissions {
			if DecidePack(isArchived, user.IsSuperuser, system, byPack[packID], name) == nil && packPermission(name) {
				out = append(out, name)
			}
		}
		perms[packID] = out
	}

	return perms, roleNames, nil
}

// packPermission reports whether name is a pack-resource permission.
func packPermission(name string) bool {
	_, global := globalPermissions[name]
	return !global
}

// VisiblePackIDs returns the pack ids on which the user holds pack.view, or
// nil with all=true when it is held globally.
func (s *Service) VisiblePackIDs(user tables.User, system Set) (ids []uint, all bool, err error) {
	if DecideGlobal(user.IsSuperuser, system, PackView) {
		return nil, true, nil
	}
	if err := s.db.Table("pack_users").
		Distinct("pack_users.pack_id").
		Joins("JOIN role_permissions ON role_permissions.role_id = pack_users.role_id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("pack_users.user_id = ? AND permissions.name = ?", user.ID, PackView).
		Scan(&ids).Error; err != nil {
		return nil, false, fmt.Errorf("load visible packs: %w", err)
	}
	return ids, false, nil
}

// IsNotFound reports whether err is ErrPackNotFound.
func IsNotFound(err error) bool { return errors.Is(err, ErrPackNotFound) }
