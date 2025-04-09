package dto

import (
	"github.com/go-playground/validator/v10"
	"packwiz-web/internal/types"
)

type AllPacksQuery struct {
	Status   []types.PackStatus `form:"status" validate:"dive,oneof=draft published"`
	Archived bool               `form:"archived"`
	Search   string             `form:"search" validate:"omitempty,max=255"`
}

func (f *AllPacksQuery) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}
