package user_svc

import (
	"fmt"

	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
)

// WithRoles pairs each user with their system roles using one query.
func (s *UserService) WithRoles(users []tables.User) ([]dto.AdminUserResponse, error) {
	out := make([]dto.AdminUserResponse, len(users))
	if len(users) == 0 {
		return out, nil
	}

	ids := make([]uint, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}

	var rows []struct {
		UserID uint
		RoleID uint
		Name   string
	}
	if err := s.db.Table("user_roles").
		Select("user_roles.user_id, roles.id as role_id, roles.name").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id IN ?", ids).
		Order("roles.id asc").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("load user roles: %w", err)
	}

	byUser := make(map[uint][]dto.RoleRef, len(users))
	for _, r := range rows {
		byUser[r.UserID] = append(byUser[r.UserID], dto.RoleRef{ID: r.RoleID, Name: r.Name})
	}

	for i, u := range users {
		roles := byUser[u.ID]
		if roles == nil {
			roles = []dto.RoleRef{}
		}
		out[i] = dto.AdminUserResponse{User: u, Roles: roles}
	}

	return out, nil
}
