package controllers

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/services/user_svc"
)

type AuthController struct {
	db  *gorm.DB
	svc *user_svc.UserService
}

func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{
		db:  db,
		svc: user_svc.NewUserService(db),
	}
}

func (ac *AuthController) Login(c *gin.Context) {
	type LoginForm struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	var form LoginForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid form data"})
		return
	}

	user, err := ac.svc.FindByUsername(form.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid username or password"})
		return
	}

	if !user.CheckPassword(form.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid username or password"})
		return
	}

	session := sessions.Default(c)
	session.Set("userId", user.Id)
	session.Set("sessionKey", ac.svc.GetOrMakeSessionKey(&user))
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to save session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func (ac *AuthController) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to save session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}
