package dto

import "github.com/go-playground/validator/v10"

type GetPackQuery struct {
	SkipMods bool `form:"skipMods"`
}

func (q *GetPackQuery) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(q)
}
