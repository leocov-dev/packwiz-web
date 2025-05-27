package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"strings"

	libConfig "github.com/leocov-dev/packwiz-nxt/config"
	"packwiz-web/internal/utils"
)

type Config struct {
	Name           string
	Version        string
	Mode           string
	AdminPassword  string
	TrustedProxies []string
	DataDir        string
	Database       string
	SessionSecret  []byte
	PGHost         string
	PGUser         string
	PGPassword     string
	PGDBName       string
	PGPort         int
}

var (
	C          Config
	versionTag string = "0.0.0-dev"
)

const (
	envMode          = "MODE"
	envAdminPassword = "ADMIN_PASSWORD"
	envProxies       = "TRUSTED_PROXIES"
	envData          = "DATA_DIR"
	envDb            = "DATABASE"
	pgHost           = "PG_HOST"
	pgPort           = "PG_PORT"
	pgUser           = "PG_USER"
	pgPassword       = "PG_PASSWORD"
	pgDBName         = "PG_DBNAME"
	envSessionSecret = "SESSION_SECRET"
	curseforgeApiKey = "CF_API_KEY"
	githubApiKey     = "GH_API_KEY"
)

func SetVersionTag(tag string) {
	if tag == "" {
		return
	}
	versionTag = tag
}

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

	config.BindEnv(envData)
	config.SetDefault(envData, filepath.Join(packwizWebRoot, "data"))

	config.BindEnv(envDb)
	config.SetDefault(envDb, "sqlite")

	config.BindEnv(pgHost)
	config.SetDefault(pgHost, "localhost")
	config.BindEnv(pgPort)
	config.SetDefault(pgPort, 5432)
	config.BindEnv(pgUser)
	config.SetDefault(pgUser, "postgres")
	config.BindEnv(pgPassword)
	config.BindEnv(pgDBName)
	config.SetDefault(pgDBName, "packwiz")

	config.BindEnv(envSessionSecret)
	config.SetDefault(envSessionSecret, "insecure-session-secret")

	config.BindEnv(curseforgeApiKey)
	config.BindEnv(githubApiKey)

	if cfApiKey := os.Getenv(curseforgeApiKey); cfApiKey != "" {
		libConfig.SetCurseforgeApiKey(cfApiKey)
	}
	if ghApiKey := os.Getenv(githubApiKey); ghApiKey != "" {
		libConfig.SetGitHubApiKey(ghApiKey)
	}

	C = Config{
		Name:           filepath.Base(exePath),
		Version:        versionTag,
		Mode:           config.GetString(envMode),
		AdminPassword:  config.GetString(envAdminPassword),
		TrustedProxies: strings.Fields(config.GetString(envProxies)),
		DataDir:        filepath.Clean(config.GetString(envData)),
		Database:       config.GetString(envDb),
		SessionSecret:  []byte(config.GetString(envSessionSecret)),
		PGHost:         config.GetString(pgHost),
		PGUser:         config.GetString(pgUser),
		PGPassword:     config.GetString(pgPassword),
		PGDBName:       config.GetString(pgDBName),
		PGPort:         config.GetInt(pgPort),
	}

	if C.AdminPassword == "" {
		panic("ADMIN_PASSWORD env var not set")
	}
	if len(C.AdminPassword) < 16 {
		panic("ADMIN_PASSWORD must be at least 16 characters")
	}

	createDirs := []string{
		C.DataDir,
	}

	for _, dir := range createDirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Panicln("failed to create directory: ", dir)
		}
	}

}
