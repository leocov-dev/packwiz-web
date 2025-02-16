package tables

import (
	"packwiz-web/internal/types"
	"time"
)

type PackUsers struct {
	PackSlug   string
	UserID     uint
	CreatedAt  time.Time
	Permission types.PackPermission
}
