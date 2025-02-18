package server

import (
	"embed"
	"fmt"
	"packwiz-web/internal/config"
	"packwiz-web/internal/logger"

	"github.com/gin-gonic/gin"
)

func Start(publicFiles *embed.FS) {
	if config.C.Mode == "development" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DefaultWriter = logger.Log.Writer()
	gin.DefaultErrorWriter = logger.Log.Writer()

	r := NewRouter(publicFiles)

	if len(config.C.TrustedProxies) > 0 {
		r.SetTrustedProxies(config.C.TrustedProxies)
	} else {
		r.SetTrustedProxies(nil)
	}

	r.Run(fmt.Sprintf(":%s", config.C.Port))
}
