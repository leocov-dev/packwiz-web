package meta

import "github.com/gin-gonic/gin"

type TagCategory string

const (
	CategoryLogin  TagCategory = "Login"
	CategoryStatic TagCategory = "StaticFile"
)

func Tag(value TagCategory) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("meta.category", value)
	}
}
