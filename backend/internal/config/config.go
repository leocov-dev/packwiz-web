package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Mode           string
	Port           string
	TrustedProxies []string
	DataDir        string
	PackwizDir     string
	Database       string
}

var C Config

const (
	envMode    = "MODE"
	envPort    = "PORT"
	envProxies = "TRUSTED_PROXIES"
	envData    = "DATA_DIR"
	envPackwiz = "PACKWIZ_DIR"
	envDb      = "DATABASE"
)

func init() {
	exePath, _ := os.Executable()
	packwizWebRoot := filepath.Join(filepath.Dir(exePath), "packwiz-web")

	config := viper.New()

	config.SetEnvPrefix("PWW")

	config.BindEnv(envMode)
	config.SetDefault(envMode, "production")

	config.BindEnv(envPort)
	config.SetDefault(envPort, "8080")

	config.BindEnv(envProxies)
	config.SetDefault(envProxies, "")

	config.BindEnv(envData)
	config.SetDefault(envData, filepath.Join(packwizWebRoot, "data"))

	config.BindEnv(envPackwiz)
	config.SetDefault(envPackwiz, filepath.Join(packwizWebRoot, "packs"))

	config.BindEnv(envDb)
	config.SetDefault(envDb, "sqlite")

	C = Config{
		Mode:           config.GetString(envMode),
		Port:           config.GetString(envPort),
		TrustedProxies: strings.Fields(config.GetString(envProxies)),
		DataDir:        config.GetString(envData),
		PackwizDir:     config.GetString(envPackwiz),
		Database:       config.GetString(envDb),
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
