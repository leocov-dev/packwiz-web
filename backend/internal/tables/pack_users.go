package tables

import (
	"packwiz-web/internal/types"
	"time"
)

type PackUsers struct {
	PackSlug   string               `gorm:"uniqueIndex:idx_pack_slug_user_id" json:"packSlug"`
	UserId     uint                 `gorm:"uniqueIndex:idx_pack_slug_user_id" json:"userId"`
	CreatedAt  time.Time            `json:"createdAt"`
	Permission types.PackPermission `json:"permission"`
}
