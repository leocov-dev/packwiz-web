package dto

import (
	"github.com/go-playground/validator/v10"
)

// RemoveModsRequest removes several mods of a pack at once.
type RemoveModsRequest struct {
	ModIds []uint `json:"modIds" validate:"required,min=1,max=1000,dive,gt=0"`
}

func (r *RemoveModsRequest) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(r)
}

// RemoveModsResponse reports how many mods were removed.
type RemoveModsResponse struct {
	Removed int `json:"removed"`
}
