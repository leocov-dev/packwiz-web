package config

import (
	"github.com/spf13/viper"
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
	SessionSecret  []byte
	PGHost         string
	PGUser         string
	PGPassword     string
	PGPort         int
	PGDbName       string
}

var (
	C          Config
	versionTag string = "0.0.0-dev"
)

const (
	envMode          = "MODE"
	envAdminPassword = "ADMIN_PASSWORD"
	envProxies       = "TRUSTED_PROXIES"
	pgDbName         = "PG_DBNAME"
	pgHost           = "PG_HOST"
	pgPort           = "PG_PORT"
	pgUser           = "PG_USER"
	pgPassword       = "PG_PASSWORD"
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

	config.BindEnv(pgHost)
	config.SetDefault(pgHost, "localhost")
	config.BindEnv(pgPort)
	config.SetDefault(pgPort, 5432)
	config.BindEnv(pgUser)
	config.SetDefault(pgUser, "postgres")
	config.BindEnv(pgPassword)
	config.BindEnv(pgDbName)
	config.SetDefault(pgDbName, "packwiz")

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
		SessionSecret:  []byte(config.GetString(envSessionSecret)),
		PGHost:         config.GetString(pgHost),
		PGUser:         config.GetString(pgUser),
		PGPassword:     config.GetString(pgPassword),
		PGDbName:       config.GetString(pgDbName),
		PGPort:         config.GetInt(pgPort),
	}

	if C.AdminPassword == "" {
		panic("ADMIN_PASSWORD env var not set")
	}
	if len(C.AdminPassword) < 16 {
		panic("ADMIN_PASSWORD must be at least 16 characters")
	}
}
