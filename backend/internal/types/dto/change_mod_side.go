package dto

import (
	"packwiz-web/internal/types"
)

type ChangeModSideRequest struct {
	SimpleRequest
	Side types.ModSide `json:"side" validate:"oneof=client server both"`
}
