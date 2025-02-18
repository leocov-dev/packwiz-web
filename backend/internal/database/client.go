package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"packwiz-web/internal/config"
	"packwiz-web/internal/logger"
	"packwiz-web/internal/services/importer"
	"packwiz-web/internal/services/packwiz_svc"
	"packwiz-web/internal/types/tables"
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
				Logger: New(gormLogLevel, logger.Log),
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
		&tables.Pack{},
		&tables.PackUsers{},
		&tables.Audit{},
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
	adminPass, _ := HashPassword("p4ckw1Z-w3b")
	defaultAdmin := tables.User{
		Username: "admin",
		Password: adminPass,
		IsAdmin:  true,
	}
	db.Create(&defaultAdmin)
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
