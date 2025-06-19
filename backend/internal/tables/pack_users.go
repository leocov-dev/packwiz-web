package tables

import (
	"packwiz-web/internal/types"
	"time"
)

type PackUsers struct {
	PackID     uint                 `json:"packId"`
	UserID     uint                 `json:"userId"`
	CreatedAt  time.Time            `json:"createdAt"`
	Permission types.PackPermission `json:"permission"`
}
