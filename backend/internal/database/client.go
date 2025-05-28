package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"path/filepath"
	"time"

	"packwiz-web/internal/config"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/utils"
)

var db *gorm.DB

func GetClient() *gorm.DB {
	return db
}

func init() {
	var gormLogLevel gormLogger.LogLevel
	if config.C.Mode == "development" {
		gormLogLevel = gormLogger.Info
	} else {
		gormLogLevel = gormLogger.Warn
	}

	switch config.C.Database {

	case "sqlite":
		db = retryConnection(func() (*gorm.DB, error) {
			return gorm.Open(
				sqlite.Open(filepath.Join(config.C.DataDir, "packwiz-web.db")),
				&gorm.Config{
					Logger: newGormLogger(gormLogLevel, log.Log),
				},
			)
		})
	case "postgres":
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			config.C.PGHost,     // e.g., "localhost"
			config.C.PGUser,     // e.g., "postgres"
			config.C.PGPassword, // e.g., "yourPassword"
			config.C.PGDBName,   // e.g., "yourDB"
			config.C.PGPort,     // e.g., 5432
		)
		db = retryConnection(func() (*gorm.DB, error) {
			return gorm.Open(
				postgres.Open(dsn),
				&gorm.Config{
					Logger: newGormLogger(gormLogLevel, log.Log),
				},
			)
		})

	default:
		panic("database not configured")

	}
}

const (
	maxRetries = 50
	retryDelay = 3 * time.Second
)

func retryConnection(openFn func() (*gorm.DB, error)) *gorm.DB {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		conn, err := openFn()
		if err == nil {
			return conn
		}
		log.Warn(fmt.Sprintf("Database connection failed (attempt %d/%d): %v", attempt, maxRetries, err))
		time.Sleep(retryDelay)
	}
	panic(fmt.Sprintf("failed to connect to database after %d attempts", maxRetries))
}

func InitDb() {

	// todo migrations

	//log.Info("Database migration completed!")

	if config.C.Database == "sqlite" {
		db.Exec("VACUUM;")
		log.Info("Database VACUUM completed!")
	}
}

func CreateDefaultAdminUser() {
	adminPass, _ := utils.HashPassword(config.C.AdminPassword)
	// update if record not found
	db.Where("username = ?", "admin").Attrs(
		tables.User{
			Username:  "admin",
			LinkToken: utils.GenerateLinkToken(16),
		},
	).FirstOrCreate(&tables.User{})

	// overwrite record or create
	db.Where("username = ?", "admin").Assign(
		tables.User{
			Password: adminPass,
			IsAdmin:  true,
		},
	).FirstOrCreate(&tables.User{})
}
