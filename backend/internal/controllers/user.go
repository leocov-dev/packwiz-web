package controllers

import (
	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/sessions"

	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/types/tables"
)

type UserController struct {
	db *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{
		db: db,
	}
}

func (uc *UserController) Login(c *gin.Context) {
	type LoginForm struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	var form LoginForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid form data"})
		return
	}

	var user tables.User
	if err := uc.db.Where("username = ?", form.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid username or password"})
		return
	}

	if !user.CheckPassword(form.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid username or password"})
		return
	}

	session := sessions.Default(c)
	session.Set("userId", user.ID)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to save session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (uc *UserController) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to save session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}
