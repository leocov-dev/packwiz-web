package dto

import (
	"github.com/go-playground/validator/v10"
	"github.com/leocov-dev/packwiz-nxt/core"
)

type ChangeModSideRequest struct {
	Side core.ModSide `json:"side" validate:"oneof=client server both"`
}

func (f *ChangeModSideRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}
