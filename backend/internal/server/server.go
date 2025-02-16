package server

import (
	"fmt"
	"packwiz-web/internal/config"
	"packwiz-web/internal/logger"

	"github.com/gin-gonic/gin"
)

func Start() {
	if config.C.Mode == "development" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DefaultWriter = logger.Log.Writer()
	gin.DefaultErrorWriter = logger.Log.Writer()

	r := NewRouter()

	if len(config.C.TrustedProxies) > 0 {
		r.SetTrustedProxies(config.C.TrustedProxies)
	} else {
		r.SetTrustedProxies(nil)
	}

	r.Run(fmt.Sprintf(":%s", config.C.Port))
}
