package dto

type LoginForm struct {
	SimpleRequest
	Username string `form:"username" validate:"required"`
	Password string `form:"password" validate:"required"`
}
