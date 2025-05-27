package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"packwiz-web/internal/config"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/utils"
	"path/filepath"
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
		// TODO
		panic("postgres database not implemented")
	default:
		panic("database not configured")

	}
}

func InitDb() {
	// Run migrations to create tables and relationships
	err := db.AutoMigrate(
		&tables.User{},
		&tables.Mod{},
		&tables.Pack{},
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
