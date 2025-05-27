package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"path/filepath"

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
	var err error

	var gormLogLevel gormLogger.LogLevel
	if config.C.Mode == "development" {
		gormLogLevel = gormLogger.Info
	} else {
		gormLogLevel = gormLogger.Warn
	}

	switch config.C.Database {

	case "sqlite":
		db, err = gorm.Open(
			sqlite.Open(filepath.Join(config.C.DataDir, "packwiz-web.db")),
			&gorm.Config{
				Logger: newGormLogger(gormLogLevel, log.Log),
			},
		)
		if err != nil {
			panic("failed to load sqlite database")
		}
	case "postgres":
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
			config.C.PGHost,     // e.g., "localhost"
			config.C.PGUser,     // e.g., "postgres"
			config.C.PGPassword, // e.g., "yourPassword"
			config.C.PGDBName,   // e.g., "yourDB"
			config.C.PGPort,     // e.g., 5432
		)
		db, err = gorm.Open(
			postgres.Open(dsn),
			&gorm.Config{
				Logger: newGormLogger(gormLogLevel, log.Log),
			},
		)
		if err != nil {
			panic("failed to load postgres database: " + err.Error())
		}

	default:
		panic("database not configured")

	}
}

func InitDb() {
	// Run migrations to create tables and relationships
	err := db.AutoMigrate(
		&tables.User{},
		&tables.Pack{},
		&tables.Mod{},
		&tables.PackUsers{},
		&tables.Audit{},
	)
	if err != nil {
		panic("failed to migrate database")
	}

	log.Info("Database migration completed!")

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
