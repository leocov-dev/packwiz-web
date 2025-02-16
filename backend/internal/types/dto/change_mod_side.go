package dto

import (
	"github.com/go-playground/validator/v10"
	"packwiz-web/internal/types"
)

type ChangeModSideRequest struct {
	Side types.ModSide `json:"side" validate:"oneof=client server both"`
}

func (r ChangeModSideRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
