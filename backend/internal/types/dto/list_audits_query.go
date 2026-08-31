package dto

import (
	"github.com/go-playground/validator/v10"
)

type ListAuditsQuery struct {
	Action    string `form:"action"`
	UserId    uint   `form:"userId"`
	StartDate string `form:"startDate" validate:"omitempty,datetime=2006-01-02"`
	EndDate   string `form:"endDate" validate:"omitempty,datetime=2006-01-02"`
	Page      int    `form:"page" validate:"gte=1"`
	PageSize  int    `form:"pageSize" validate:"gte=1,lte=100"`
}

func (f *ListAuditsQuery) Validate() error {
	return validator.New(validator.WithRequiredStructEnabled()).Struct(f)
}
