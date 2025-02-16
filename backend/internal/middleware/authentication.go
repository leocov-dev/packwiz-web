package middleware

import (
	"github.com/gin-gonic/gin"
	"packwiz-web/internal/database"
	"packwiz-web/internal/types/tables"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		var admin tables.User
		db := database.GetClient()
		err := db.Where("username = ?", "admin").First(&admin).Error
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}

		c.Set("user", admin)

		c.Next()
	}
}
