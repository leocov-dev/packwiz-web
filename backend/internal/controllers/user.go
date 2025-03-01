package controllers

import (
	"github.com/gin-gonic/gin"
	"packwiz-web/internal/tables"

	"gorm.io/gorm"
	"net/http"
)

type UserController struct {
	db *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{
		db: db,
	}
}

func (uc *UserController) CurrentUser(c *gin.Context) {
	user := c.MustGet("user").(tables.User)
	c.JSON(http.StatusOK, user)
}
