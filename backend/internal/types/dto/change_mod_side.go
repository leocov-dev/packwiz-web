package dto

import (
	"github.com/go-playground/validator/v10"
	"packwiz-web/internal/types"
)

type ChangeModSideRequest struct {
	Side types.ModSide `json:"side" validate:"oneof=client server both"`
}

func (f *ChangeModSideRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}
