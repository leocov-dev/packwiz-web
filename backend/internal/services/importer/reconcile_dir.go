package importer

import (
	"gorm.io/gorm"
	"packwiz-web/internal/services/packwiz_svc"
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

//func (dr *DataReconciler) ReconcilePackwizDir() error {
//	errorGroup := interfaces.NewErrorGroup()
//
//	entries, err := os.ReadDir(config.C.PackwizDir)
//	if err != nil {
//		errorGroup.Add(fmt.Errorf("failed to read PackwizDir: %w", err))
//		return errorGroup
//	}
//
//	var admin tables.User
//	dr.db.Where("username = ?", "admin").First(&admin)
//
//	for _, entry := range entries {
//		// Skip files, we only care about directories.
//		if !entry.IsDir() {
//			continue
//		}
//		if strings.HasPrefix(entry.Name(), ".") {
//			continue
//		}
//
//		slug := entry.Name()
//		_, err := packwiz_cli.GetPackFile(slug)
//		if err != nil {
//			err := dr.packwizSvc.ArchivePack(slug)
//			if err != nil {
//				errorGroup.Add(fmt.Errorf("failed to archive pack '%s': %w", slug, err))
//			}
//			errorGroup.Add(fmt.Errorf("failed to find pack.toml for modpack '%s' or data corrupted. %w", slug, err))
//			continue
//		}
//
//		pack := tables.Pack{
//			Slug:        slug,
//			Description: "pack imported from packwiz dir",
//			IsPublic:    false,
//			Status:      types.PackStatusDraft,
//		}
//		if err = dr.db.Where("slug = ?", slug).Attrs(pack).FirstOrCreate(&pack).Error; err != nil {
//			errorGroup.Add(fmt.Errorf("failed to import pack '%s': %w", slug, err))
//			continue
//		}
//
//		packUser := tables.PackUsers{
//			PackSlug:   slug,
//			UserId:     admin.ID,
//			Permission: types.PackPermissionEdit,
//		}
//		if err = dr.db.Where("pack_slug = ? AND user_id = ?", slug, admin.ID).Attrs(packUser).FirstOrCreate(&packUser).Error; err != nil {
//			errorGroup.Add(fmt.Errorf("failed to set admin user on pack '%s': %w", slug, err))
//			continue
//		}
//	}
//
//	if errorGroup.IsEmpty() {
//		return nil
//	}
//
//	return errorGroup
//}
