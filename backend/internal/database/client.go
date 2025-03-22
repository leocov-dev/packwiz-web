package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"packwiz-web/internal/config"
	"packwiz-web/internal/log"
	"packwiz-web/internal/services/importer"
	"packwiz-web/internal/services/packwiz_svc"
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

	createDefaultAdminUser()

	reconcileFileData()
}

func createDefaultAdminUser() {
	adminPass, _ := utils.HashPassword(config.C.AdminPassword)

	// overwrite record or create
	db.Where("username = ?", "admin").Assign(
		tables.User{
			Username: "admin",
			Password: adminPass,
			IsAdmin:  true,
		},
	).FirstOrCreate(&tables.User{})

	// update if record not found
	db.Where("username = ?", "admin").Attrs(
		tables.User{
			LinkToken: utils.GenerateRandomString(32),
		},
	).FirstOrCreate(&tables.User{})
}

func SeedDebugData() {
	createDummyPlayerUser()
}

func createDummyPlayerUser() {
	pass, _ := utils.HashPassword("password123")

	db.Create(
		&tables.User{
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
		log.Error(err)
	}
}
