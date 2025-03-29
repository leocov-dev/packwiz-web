package dto

type ChangePasswordForm struct {
	SimpleRequest
	OldPassword string `form:"oldPassword" validate:"required"`
	NewPassword string `form:"newPassword" validate:"required"`
}
