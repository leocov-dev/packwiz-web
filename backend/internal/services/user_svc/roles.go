package user_svc

import "gorm.io/gorm"

// DefaultRoleName is the system role every new user starts with.
const DefaultRoleName = "user"

// AssignDefaultRole gives a new user the default system role.
func AssignDefaultRole(tx *gorm.DB, userID uint) error {
	return tx.Exec(
		"INSERT INTO user_roles (user_id, role_id) SELECT ?, id FROM roles WHERE name = ? ON CONFLICT DO NOTHING",
		userID, DefaultRoleName,
	).Error
}
