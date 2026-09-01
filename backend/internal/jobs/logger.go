package jobs

import (
	"log/slog"
	"os"

	"packwiz-web/internal/config"
)

// newSlogLogger builds a standalone *slog.Logger for River, since its
// Config.Logger field requires log/slog rather than logrus. Mirrors the
// dev/prod split already used by internal/log without depending on it.
func newSlogLogger() *slog.Logger {
	var handler slog.Handler
	if config.C.Mode == "development" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	return slog.New(handler)
}
