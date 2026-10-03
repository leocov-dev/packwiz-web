package meta

import "github.com/gin-gonic/gin"

type TagCategory string

const (
	CategoryLogin  TagCategory = "Login"
	CategoryStatic TagCategory = "StaticFile"

	CategoryOidcLink   TagCategory = "OidcLink"
	CategoryOidcUnlink TagCategory = "OidcUnlink"

	CategoryPackRevert TagCategory = "PackRevert"
	CategoryPackClone  TagCategory = "PackClone"
	CategoryPackPrune  TagCategory = "PackPrune"
	CategoryPackRebase TagCategory = "PackRebase"
)

func Tag(value TagCategory) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("meta.category", value)
	}
}
