package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"packwiz-web/internal/config"
	"packwiz-web/internal/log"

	"github.com/gin-gonic/gin"
)

const shutdownTimeout = 10 * time.Second

// Start serves HTTP until ctx is cancelled, then drains in-flight requests.
func Start(ctx context.Context) {
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

	srv := &http.Server{Addr: ":8080", Handler: r}

	drained := make(chan struct{})
	go func() {
		defer close(drained)
		<-ctx.Done()
		log.Info("shutdown signal received, stopping server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("error stopping server:", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server error:", err)
		return
	}

	// ListenAndServe returns as soon as Shutdown begins, not when it finishes
	<-drained
	log.Info("server stopped")
}
