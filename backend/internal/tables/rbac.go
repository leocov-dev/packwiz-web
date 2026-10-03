package tables

// Role is a named set of permissions. System roles are attached to a user via
// UserRole, pack roles via PackUsers.
type Role struct {
	ID          uint         `gorm:"primarykey" json:"id"`
	Name        string       `gorm:"unique" json:"name"`
	Description string       `json:"description"`
	Scope       string       `json:"scope"`
	Assignable  bool         `json:"assignable"`
	Permissions []Permission `gorm:"many2many:role_permissions" json:"-"`
}

// Permission is a flat named capability. Resource is "pack" or "global".
type Permission struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	Name        string `gorm:"unique" json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
}

// UserRole attaches a system role to a user.
type UserRole struct {
	UserID uint `gorm:"primaryKey" json:"userId"`
	RoleID uint `gorm:"primaryKey" json:"roleId"`
}

const (
	RoleScopeSystem = "system"
	RoleScopePack   = "pack"

	PermissionResourcePack   = "pack"
	PermissionResourceGlobal = "global"
)
