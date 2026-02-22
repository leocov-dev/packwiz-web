package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/services/user_svc"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

type AdminController struct {
	db  *gorm.DB
	svc *user_svc.UserService
}

func NewAdminController(db *gorm.DB) *AdminController {
	return &AdminController{
		db:  db,
		svc: user_svc.NewUserService(db),
	}
}

func (uc *AdminController) GetUsersPaginated(c *gin.Context) {
	var query dto.ListUsersQuery
	if err := mustBindQuery(c, &query); err != nil {
		err.JSON(c)
		return
	}

	users, total, err := uc.svc.ListUsers(query)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, response.NewPaginated(
		users,
		query.Page,
		query.PageSize,
		total,
	))
}
