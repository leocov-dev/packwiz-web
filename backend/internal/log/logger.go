package log

import (
	"github.com/sirupsen/logrus"
	"os"
	"packwiz-web/internal/config"
)

var Log = logrus.New()

func init() {
	Log.SetOutput(os.Stdout)

	if config.C.Mode == "development" {
		Log.SetLevel(logrus.DebugLevel)
		Log.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp: true,
			ForceColors:      true,
		})
	} else {
		Log.SetLevel(logrus.InfoLevel)
		Log.SetFormatter(&logrus.JSONFormatter{})
	}
}

func Debug(args ...interface{}) {
	Log.Debugln(args...)
}
func Info(args ...interface{}) {
	Log.Infoln(args...)
}
func Warn(args ...interface{}) {
	Log.Warnln(args...)
}
func Error(args ...interface{}) {
	Log.Errorln(args...)
}
func Panic(args ...interface{}) {
	Log.Panicln(args...)
}
