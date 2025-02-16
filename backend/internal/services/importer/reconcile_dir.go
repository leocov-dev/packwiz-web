package importer

import (
	"fmt"
	"gorm.io/gorm"
	"os"
	"packwiz-web/internal/config"
	"packwiz-web/internal/interfaces"
	"packwiz-web/internal/services/packwiz_cli"
	"packwiz-web/internal/services/packwiz_svc"
	"packwiz-web/internal/types/tables"
)

type DataReconciler struct {
	db         *gorm.DB
	packwizSvc *packwiz_svc.PackwizService
}

func NewDataReconciler(db *gorm.DB, packwizSvc *packwiz_svc.PackwizService) *DataReconciler {
	return &DataReconciler{
		db:         db,
		packwizSvc: packwizSvc,
	}
}

func (dr *DataReconciler) ReconcilePackwizDir() error {
	errorGroup := interfaces.NewErrorGroup()

	entries, err := os.ReadDir(config.C.PackwizDir)
	if err != nil {
		errorGroup.Add(fmt.Errorf("failed to read PackwizDir: %w", err))
		return errorGroup
	}

	var admin tables.User
	dr.db.Where("username = ?", "admin").First(&admin)

	for _, entry := range entries {
		// Skip files, we only care about directories.
		if !entry.IsDir() {
			continue
		}

		slug := entry.Name()
		_, err := packwiz_cli.GetPackFile(slug)
		if err != nil {
			errorGroup.Add(fmt.Errorf("failed to find pack.toml for modpack '%s' or data corrupted. %w", slug, err))
			continue
		}

		pack := tables.Pack{
			Slug:        slug,
			Description: "pack imported from packwiz dir",
			Users:       []tables.User{admin},
			IsPublic:    false,
		}
		if err = dr.db.Where("slug = ?", slug).Attrs(pack).FirstOrCreate(&pack).Error; err != nil {
			errorGroup.Add(fmt.Errorf("failed to import pack '%s': %w", slug, err))
			continue
		}
	}

	if errorGroup.IsEmpty() {
		return nil
	}

	return errorGroup
}
