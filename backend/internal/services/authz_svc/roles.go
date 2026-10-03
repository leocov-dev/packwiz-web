package authz_svc

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"packwiz-web/internal/tables"
)

var (
	// ErrSuperuserProtected means the target is the bootstrap superuser.
	ErrSuperuserProtected = errors.New("the superuser's roles cannot be changed")
	// ErrInvalidRole means a role id is unknown or not an assignable system role.
	ErrInvalidRole = errors.New("invalid role")
	// ErrUserNotFound means the target user does not exist.
	ErrUserNotFound = errors.New("user not found")
)

// RoleInfo is a role with the permission names it grants.
type RoleInfo struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Scope       string   `json:"scope"`
	Assignable  bool     `json:"assignable"`
	Permissions []string `json:"permissions"`
}

// ListRoles lists roles with their permissions. scope filters when non-empty.
func (s *Service) ListRoles(scope string) ([]RoleInfo, error) {
	query := s.db.Preload("Permissions").Order("id asc")
	if scope != "" {
		query = query.Where("scope = ?", scope)
	}

	var roles []tables.Role
	if err := query.Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}

	out := make([]RoleInfo, 0, len(roles))
	for _, r := range roles {
		names := make([]string, 0, len(r.Permissions))
		for _, p := range r.Permissions {
			names = append(names, p.Name)
		}
		out = append(out, RoleInfo{
			ID: r.ID, Name: r.Name, Description: r.Description,
			Scope: r.Scope, Assignable: r.Assignable, Permissions: names,
		})
	}
	return out, nil
}

// ListPermissions lists every permission.
func (s *Service) ListPermissions() ([]tables.Permission, error) {
	var perms []tables.Permission
	if err := s.db.Order("resource asc, name asc").Find(&perms).Error; err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}
	return perms, nil
}

// SetUserRoles replaces a user's system roles. The superuser is refused.
func (s *Service) SetUserRoles(targetID uint, roleIDs []uint) error {
	var target tables.User
	if err := s.db.Where("id = ?", targetID).First(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("find user: %w", err)
	}
	if target.IsSuperuser {
		return ErrSuperuserProtected
	}

	unique := make(map[uint]struct{}, len(roleIDs))
	for _, id := range roleIDs {
		unique[id] = struct{}{}
	}
	ids := make([]uint, 0, len(unique))
	for id := range unique {
		ids = append(ids, id)
	}

	if len(ids) > 0 {
		var count int64
		if err := s.db.Model(&tables.Role{}).
			Where("id IN ? AND scope = ? AND assignable", ids, tables.RoleScopeSystem).
			Count(&count).Error; err != nil {
			return fmt.Errorf("check roles: %w", err)
		}
		if int(count) != len(ids) {
			return ErrInvalidRole
		}
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", targetID).Delete(&tables.UserRole{}).Error; err != nil {
			return fmt.Errorf("clear roles: %w", err)
		}
		for _, id := range ids {
			if err := tx.Create(&tables.UserRole{UserID: targetID, RoleID: id}).Error; err != nil {
				return fmt.Errorf("assign role: %w", err)
			}
		}
		return nil
	})
}
