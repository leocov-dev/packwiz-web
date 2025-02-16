package tables

import (
	"packwiz-web/internal/types"
	"time"
)

type Pack struct {
	Slug        string          `gorm:"primarykey" json:"slug"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	Description string          `json:"description"`
	CreatedBy   uint            `json:"createdBy"`
	Users       []User          `gorm:"many2many:pack_users;" json:"users"`
	IsPublic    bool            `json:"isPublic"`
	PackData    types.PackData  `gorm:"-" json:"packData"`
	ModData     []types.ModData `gorm:"-" json:"modData"`
}
