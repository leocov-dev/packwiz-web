package server

import (
	"packwiz-web/internal/config"
	"packwiz-web/internal/log"

	"github.com/gin-gonic/gin"
)

func Start() {
	if config.C.Mode == "development" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DefaultWriter = log.Log.Writer()
	gin.DefaultErrorWriter = log.Log.Writer()

	r := NewRouter()

	if len(config.C.TrustedProxies) > 0 {
		r.SetTrustedProxies(config.C.TrustedProxies)
	} else {
		r.SetTrustedProxies(nil)
	}

	r.Run(":8080")
}
