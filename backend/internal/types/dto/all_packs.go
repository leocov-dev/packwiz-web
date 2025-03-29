package dto

import (
	"packwiz-web/internal/types"
)

type AllPacksQuery struct {
	SimpleRequest
	Status   []types.PackStatus `form:"status" validate:"dive,oneof=draft published"`
	Archived bool               `form:"archived"`
	Search   string             `form:"search"`
}
