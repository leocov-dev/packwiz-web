package tables

import (
	"time"
)

// PackUsers grants a pack-scope role to a user on one pack.
type PackUsers struct {
	PackID    uint      `json:"packId"`
	UserID    uint      `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	RoleID    uint      `json:"roleId"`
	Role      Role      `json:"-"`
}
