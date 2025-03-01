package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"packwiz-web/internal/utils"
	"path/filepath"
	"strings"
)

type Config struct {
	Mode           string
	AdminPassword  string
	TrustedProxies []string
	DataDir        string
	PackwizDir     string
	Database       string
	SessionSecret  []byte
}

var C Config

const (
	envMode          = "MODE"
	envAdminPassword = "ADMIN_PASSWORD"
	envProxies       = "TRUSTED_PROXIES"
	envData          = "DATA_DIR"
	envPackwiz       = "PACKWIZ_DIR"
	envDb            = "DATABASE"
	envSessionSecret = "SESSION_SECRET"
)

func init() {
	exePath, _ := os.Executable()
	packwizWebRoot := filepath.Join(filepath.Dir(exePath), "packwiz-web")

	config := viper.New()

	config.SetEnvPrefix("PWW")

	config.BindEnv(envMode)
	config.SetDefault(envMode, "production")

	config.BindEnv(envAdminPassword)
	// you can't log into the system if you don't manually set the admin password
	// env var. the default is an unknown random string to prevent people from
	// ignoring the instructions to choose their own.
	config.SetDefault(envAdminPassword, utils.GenerateRandomString(32))

	config.BindEnv(envProxies)
	config.SetDefault(envProxies, "")

	config.BindEnv(envData)
	config.SetDefault(envData, filepath.Join(packwizWebRoot, "data"))

	config.BindEnv(envPackwiz)
	config.SetDefault(envPackwiz, filepath.Join(packwizWebRoot, "packs"))

	config.BindEnv(envDb)
	config.SetDefault(envDb, "sqlite")

	config.BindEnv(envSessionSecret)
	config.SetDefault(envSessionSecret, "insecure-session-secret")

	C = Config{
		Mode:           config.GetString(envMode),
		AdminPassword:  config.GetString(envAdminPassword),
		TrustedProxies: strings.Fields(config.GetString(envProxies)),
		DataDir:        filepath.Clean(config.GetString(envData)),
		PackwizDir:     filepath.Clean(config.GetString(envPackwiz)),
		Database:       config.GetString(envDb),
		SessionSecret:  []byte(config.GetString(envSessionSecret)),
	}

	if C.AdminPassword == "" {
		panic("ADMIN_PASSWORD env var not set")
	}
	if len(C.AdminPassword) < 16 {
		panic("ADMIN_PASSWORD must be at least 16 characters")
	}

	createDirs := []string{
		C.DataDir,
		C.PackwizDir,
	}

	for _, dir := range createDirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Panicln("failed to create directory: ", dir)
		}
	}

}
