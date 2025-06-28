package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"time"

	"packwiz-web/internal/config"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/utils"
)

var db *gorm.DB

func IsConnected() bool {
	if db == nil {
		return false
	}

	sqlDB, err := db.DB()
	if err != nil {
		return false
	}

	return sqlDB.Ping() == nil
}

func GetClient() *gorm.DB {
	if !IsConnected() {
		connect()
	}
	return db
}

func connect() {
	var gormLogLevel gormLogger.LogLevel
	if config.C.Mode == "development" {
		gormLogLevel = gormLogger.Info
	} else {
		gormLogLevel = gormLogger.Warn
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		config.C.PGHost,
		config.C.PGUser,
		config.C.PGPassword,
		config.C.PGDbName,
		config.C.PGPort,
	)
	db = retryConnection(func() (*gorm.DB, error) {
		return gorm.Open(
			postgres.Open(dsn),
			&gorm.Config{
				Logger: newGormLogger(gormLogLevel, log.Log),
			},
		)
	})
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
		log.Warn(fmt.Sprintf("DatabaseType connection failed (attempt %d/%d): %v", attempt, maxRetries, err))
		time.Sleep(retryDelay)
	}
	panic(fmt.Sprintf("failed to connect to database after %d attempts", maxRetries))
}

func UpsertDefaultAdminUser() {
	adminPass, _ := utils.HashPassword(config.C.AdminPassword)
	// update if record not found
	GetClient().Where("username = ?", "admin").Attrs(
		tables.User{
			Username:  "admin",
			LinkToken: utils.GenerateLinkToken(16),
		},
	).FirstOrCreate(&tables.User{})

	// overwrite record or create
	GetClient().Where("username = ?", "admin").Assign(
		tables.User{
			Password: adminPass,
			IsAdmin:  true,
		},
	).FirstOrCreate(&tables.User{})
}
