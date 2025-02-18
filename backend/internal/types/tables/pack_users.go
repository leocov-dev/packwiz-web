package tables

import (
	"packwiz-web/internal/types"
	"time"
)

type PackUsers struct {
	PackSlug   string               `json:"packSlug"`
	UserID     uint                 `json:"userId"`
	CreatedAt  time.Time            `json:"createdAt"`
	Permission types.PackPermission `json:"permission"`
}
