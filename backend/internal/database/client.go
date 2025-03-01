package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"packwiz-web/internal/config"
	"packwiz-web/internal/logger"
	"packwiz-web/internal/services/importer"
	"packwiz-web/internal/services/packwiz_svc"
	tables2 "packwiz-web/internal/tables"
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
				Logger: newGormLogger(gormLogLevel, logger.Log),
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
		&tables2.User{},
		&tables2.Pack{},
		&tables2.PackUsers{},
		&tables2.Audit{},
	)
	if err != nil {
		panic("failed to migrate database")
	}

	logger.Info("Database migration completed!")

	if config.C.Database == "sqlite" {
		db.Exec("VACUUM;")
		logger.Info("Database VACUUM completed!")
	}

	createDefaultAdminUser()

	reconcileFileData()
}

func createDefaultAdminUser() {
	adminPass, _ := utils.HashPassword(config.C.AdminPassword)

	var defaultAdmin tables2.User
	db.Where("username = ?", "admin").Assign(
		tables2.User{
			Username:  "admin",
			Password:  adminPass,
			IsAdmin:   true,
			LinkToken: utils.GenerateRandomString(32),
		},
	).FirstOrCreate(&defaultAdmin)
}

func SeedDebugData() {
	createDummyPlayerUser()
}

func createDummyPlayerUser() {
	pass, _ := utils.HashPassword("password123")

	db.Create(
		&tables2.User{
			Username:  "player",
			Password:  pass,
			IsAdmin:   false,
			LinkToken: utils.GenerateRandomString(32),
		},
	)
}

func reconcileFileData() {
	reconciler := importer.NewDataReconciler(
		db,
		packwiz_svc.NewPackwizService(db),
	)
	err := reconciler.ReconcilePackwizDir()
	if err != nil {
		panic(err)
	}
}
