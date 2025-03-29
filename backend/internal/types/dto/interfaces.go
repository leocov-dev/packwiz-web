package dto

import "github.com/go-playground/validator/v10"

type Request interface {
	Validate() error
}

type SimpleRequest struct {
}

func (r SimpleRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
