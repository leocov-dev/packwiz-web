package dto

type EditUserRequest struct {
	SimpleRequest
	Username string `json:"username" validate:"required"`
	FullName string `json:"fullName" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}
