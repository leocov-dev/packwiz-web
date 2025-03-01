package tables

import (
	"gorm.io/gorm"
	"packwiz-web/internal/types"
	"time"
)

type Pack struct {
	Slug        string               `gorm:"primarykey" json:"slug"`
	CreatedAt   time.Time            `json:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt       `gorm:"index" json:"deletedAt"`
	Description string               `json:"description"`
	CreatedBy   uint                 `json:"createdBy"`
	IsPublic    bool                 `json:"isPublic"`
	Status      types.PackStatus     `gorm:"default:draft" json:"status"`
	IsArchived  bool                 `gorm:"-" json:"isArchived"`
	Permission  types.PackPermission `gorm:"-" json:"permission"`
	PackData    *types.PackData      `gorm:"-" json:"packData"`
	ModData     []types.ModData      `gorm:"-" json:"modData"`
}
