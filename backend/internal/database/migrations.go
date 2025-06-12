package database

import (
	"embed"
	"fmt"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	"packwiz-web/internal/config"
	"packwiz-web/internal/log"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// createMigrateInstance creates a new migrate instance with all the necessary setup
func createMigrateInstance() (*migrate.Migrate, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create database driver: %v", err)
	}

	sourceDriver, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create source driver: %v", err)
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		sourceDriver,
		config.C.PGDbName,
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %v", err)
	}

	return m, nil
}

func RunMigrations() error {
	m, err := createMigrateInstance()
	if err != nil {
		return err
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %v", err)
	}

	log.Info(fmt.Sprintf("Current migration version: %d, dirty: %t", version, dirty))

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Info("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	log.Info("Migrations applied successfully")
	return nil
}

func RollbackMigration(steps int) error {
	m, err := createMigrateInstance()
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(-steps); err != nil {
		return fmt.Errorf("failed to rollback migrations: %v", err)
	}

	log.Info(fmt.Sprintf("Rolled back %d migration(s)", steps))
	return nil
}
