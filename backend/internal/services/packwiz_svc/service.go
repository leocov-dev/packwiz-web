package packwiz_svc

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"packwiz-web/internal/logger"
	"packwiz-web/internal/services/packwiz_cli"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types"
	"packwiz-web/internal/types/dto"
	"time"
)

type PackwizService struct {
	db *gorm.DB
}

func NewPackwizService(db *gorm.DB) *PackwizService {
	return &PackwizService{
		db,
	}
}

func (ps *PackwizService) GetPacks(
	statusFilter []types.PackStatus,
	includeArchived bool,
	search string,
	userId uint,
) ([]tables.Pack, error) {
	if len(statusFilter) == 0 && !includeArchived {
		statusFilter = []types.PackStatus{types.PackStatusDraft, types.PackStatusPublished}
	}

	type Result struct {
		tables.Pack
		Permission types.PackPermission
	}
	var results []Result
	packs := make([]tables.Pack, 0)

	query := ps.db.Model(
		&tables.Pack{},
	).Select(
		"packs.*, pack_users.permission AS permission",
	).Joins(
		"LEFT JOIN pack_users ON packs.slug = pack_users.pack_slug AND pack_users.user_id = ?",
		userId,
	).Where(
		"permission >= ?", types.PackPermissionView,
	).Where(
		"packs.status IN ?", statusFilter,
	).Order("packs.slug asc")

	if search != "" {
		query = query.Where("packs.slug LIKE ?", "%"+search+"%")
	}

	if !includeArchived {
		query = query.Where("deleted_at IS NULL")
	}

	if err := query.Unscoped().Scan(&results).Error; err != nil {
		return nil, err
	}

	for _, result := range results {
		pack := result.Pack
		if err := ps.hydratePackData(&pack); err != nil {
			logger.Warn(fmt.Sprintf("failed to hydrate data for pack %s, %w", pack.Slug, err))
		}
		pack.Permission = result.Permission
		packs = append(packs, pack)
	}

	logger.Info(fmt.Sprintf("Found %d packs", len(packs)))

	return packs, nil
}

func (ps *PackwizService) PackExists(slug string) bool {
	if err := ps.db.Unscoped().Where("slug = ?", slug).First(&tables.Pack{}).Error; err != nil {
		logger.Debug("pack not exists:", slug)
		return false
	}

	return true
}

func (ps *PackwizService) NewPack(request dto.NewPackRequest, author tables.User) error {
	slug := request.Slug()

	return ps.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&tables.Pack{
			Slug:      slug,
			CreatedBy: author.Id,
			Status:    types.PackStatusDraft,
		}).Error; err != nil {
			return err
		}

		if err := tx.Create(&tables.PackUsers{
			PackSlug:   slug,
			UserId:     author.Id,
			Permission: types.PackPermissionEdit,
		}).Error; err != nil {
			return err
		}
		return packwiz_cli.NewModpack(
			slug,
			author.Username,
			request.MinecraftDef.AsCliType(),
			request.LoaderDef.AsCliType(),
		)
	})
}

func (ps *PackwizService) GetPack(slug string, userId uint, hydrateData, hydrateMods bool) (tables.Pack, error) {
	var pack tables.Pack
	err := ps.db.Unscoped().Where(&tables.Pack{Slug: slug}).First(&pack).Error
	if err != nil {
		return pack, err
	}

	if hydrateData {
		err = ps.hydratePackData(&pack)
		if err != nil {
			logger.Warn(fmt.Sprintf("failed to hydrate data for pack %s, %w", pack.Slug, err))
		}
	}

	if hydrateMods {
		err = ps.hydrateModData(&pack)
		if err != nil {
			logger.Warn(fmt.Sprintf("failed to hydrate mods for pack %s, %w", pack.Slug, err))
		}
	}

	return pack, nil
}

// AddMod
// Add a new mod to an existing pack
func (ps *PackwizService) AddMod(slug string, request dto.AddModRequest) error {
	if request.Modrinth.IsSet() {
		data := request.Modrinth
		return packwiz_cli.AddModrinthMod(slug, data.Name, data.ProjectId, data.VersionFilename, data.VersionId)
	} else if request.Curseforge.IsSet() {
		data := request.Curseforge
		return packwiz_cli.AddCurseforgeMod(slug, data.Name, data.AddonId, data.Category, data.FileId, data.Game)
	}

	return errors.New("invalid add mod request")
}

// ArchivePack
// remove a pack and all mods from disk and db
func (ps *PackwizService) ArchivePack(slug string) error {
	return ps.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(
			&tables.Pack{Slug: slug},
		).Updates(
			&tables.Pack{
				IsPublic: false,
				Status:   types.PackStatusDraft,
				DeletedAt: gorm.DeletedAt{
					Time:  time.Now(),
					Valid: true,
				},
			},
		).Error; err != nil {
			return err
		}
		return packwiz_cli.DeleteModpack(slug)
	})
}

// SetPackStatus
// change the pack status
func (ps *PackwizService) SetPackStatus(slug string, status types.PackStatus) error {
	return ps.db.Model(&tables.Pack{Slug: slug}).Update("status", status).Error
}

func (ps *PackwizService) IsPackPublished(slug string) bool {
	var pack tables.Pack
	err := ps.db.Where(&tables.Pack{Slug: slug}).First(&pack).Error
	if err != nil {
		return false
	}
	return pack.Status == types.PackStatusPublished
}

// SetAcceptableVersions
// set a mod packs acceptable minecraft versions
func (ps *PackwizService) SetAcceptableVersions(slug string, request dto.SetAcceptableVersionsRequest) error {
	return packwiz_cli.SetAcceptableVersions(slug, request.Versions...)
}

// UpdateAll
// update all the mods in a pack, skipping pinned mods
func (ps *PackwizService) UpdateAll(slug string) error {
	return packwiz_cli.UpdateAll(slug)
}

// ModExists
// check if a mod exists in a pack
func (ps *PackwizService) ModExists(slug, mod string) bool {
	return packwiz_cli.ModExists(slug, mod) == nil
}

// RemoveMod
// remove a given mod from a given pack
func (ps *PackwizService) RemoveMod(slug, mod string) error {
	return packwiz_cli.RemoveMod(slug, mod)
}

// UpdateMod
// update a given mod from a given pack
func (ps *PackwizService) UpdateMod(slug, mod string) error {
	return packwiz_cli.UpdateOne(slug, mod)
}

// GetMod
// get a single mods data
func (ps *PackwizService) GetMod(slug, mod string) (*types.ModData, error) {
	data, err := findModData(slug, mod)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (ps *PackwizService) ChangeModSide(slug, mod string, side types.ModSide) error {
	return packwiz_cli.ChangeModSide(slug, mod, side)
}

// PinMod
// pin a mod to prevent it from being updated
func (ps *PackwizService) PinMod(slug, mod string) error {
	return packwiz_cli.PinMod(slug, mod)
}

// UnpinMod
// unpin a mod to allow it to be updated
func (ps *PackwizService) UnpinMod(slug, mod string) error {
	return packwiz_cli.UnpinMod(slug, mod)
}

// ---

func (ps *PackwizService) hydratePackData(pack *tables.Pack) error {
	var err error
	pack.PackData, err = getModpackData(pack.Slug)

	return err
}

func (ps *PackwizService) hydrateModData(pack *tables.Pack) error {
	var err error
	pack.ModData, err = getModpackMods(pack.Slug)

	return err
}
