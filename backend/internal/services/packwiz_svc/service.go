package packwiz_svc

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"packwiz-web/internal/log"
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
		IsArchived bool
		Permission types.PackPermission
	}
	var results []Result
	packs := make([]tables.Pack, 0)

	query := ps.db.Model(
		&tables.Pack{},
	).Select(
		"packs.*, pack_users.permission AS permission, (deleted_at IS NOT NULL) AS is_archived",
	).Joins(
		"LEFT JOIN pack_users ON packs.slug = pack_users.pack_slug AND pack_users.user_id = ?",
		userId,
	).Where(
		ps.db.Where(
			"pack_users.permission >= ?", types.PackPermissionStatic,
		).Or(
			"packs.is_public",
		),
	).Order("packs.slug asc")

	if search != "" {
		query = query.Where("packs.slug LIKE ?", "%"+search+"%")
	}

	sub := ps.db
	if len(statusFilter) > 0 {
		sub = sub.Where("packs.status IN ?", statusFilter)
	}

	if includeArchived {
		sub = sub.Or("deleted_at IS NOT NULL")
	} else {
		sub = sub.Where("deleted_at IS NULL")
	}

	query = query.Where(sub)

	if err := query.Unscoped().Scan(&results).Error; err != nil {
		return nil, err
	}

	for _, result := range results {
		pack := result.Pack
		if err := ps.hydratePackData(&pack); err != nil {
			log.Warn(fmt.Sprintf("failed to hydrate data for pack %s, %w", pack.Slug, err))
		}
		pack.IsArchived = result.IsArchived
		pack.Permission = result.Permission
		packs = append(packs, pack)
	}

	log.Info(fmt.Sprintf("Found %d packs", len(packs)))

	return packs, nil
}

func (ps *PackwizService) PackExists(slug string) bool {
	if err := ps.db.Unscoped().Where("slug = ?", slug).First(&tables.Pack{}).Error; err != nil {
		log.Debug("pack not exists:", slug)
		return false
	}

	return true
}

func (ps *PackwizService) NewPack(request dto.NewPackRequest, author tables.User) error {

	name := request.Name
	if name == "" {
		name = request.Slug
	}

	return ps.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&tables.Pack{
			Slug:        request.Slug,
			CreatedBy:   author.Id,
			Status:      types.PackStatusDraft,
			Description: request.Description,
		}).Error; err != nil {
			return err
		}

		if err := tx.Create(&tables.PackUsers{
			PackSlug:   request.Slug,
			UserId:     author.Id,
			Permission: types.PackPermissionEdit,
		}).Error; err != nil {
			return err
		}

		if err := packwiz_cli.NewModpack(
			request.Slug,
			name,
			author.Username,
			request.Version,
			request.MinecraftDef.AsCliType(),
			request.LoaderDef.AsCliType(),
		); err != nil {
			return err
		}

		if request.AcceptableVersions != nil && len(request.AcceptableVersions) > 0 {
			if err := packwiz_cli.SetAcceptableVersions(
				request.Slug,
				request.AcceptableVersions...,
			); err != nil {
				return err
			}
		}

		return nil
	})
}

func (ps *PackwizService) GetPack(slug string, userId uint, hydrateData, hydrateMods bool) (tables.Pack, error) {
	type Result struct {
		tables.Pack
		IsArchived bool
		Permission types.PackPermission
	}
	var result Result

	query := ps.db.Model(
		&tables.Pack{},
	).Select(
		"packs.*, pack_users.permission AS permission, (deleted_at IS NOT NULL) AS is_archived",
	).Joins(
		"JOIN pack_users ON packs.slug = pack_users.pack_slug AND pack_users.user_id = ?", userId,
	).Where(
		"packs.slug = ?", slug,
	).Order("packs.slug asc")

	err := query.Unscoped().First(&result).Error
	if err != nil {
		return tables.Pack{}, err
	}

	result.Pack.IsArchived = result.IsArchived
	result.Pack.Permission = result.Permission

	if hydrateData {
		err = ps.hydratePackData(&result.Pack)
		if err != nil {
			log.Warn(fmt.Sprintf("failed to hydrate data for pack %s, %w", slug, err))
		}
	}

	if hydrateMods {
		err = ps.hydrateModData(&result.Pack)
		if err != nil {
			log.Warn(fmt.Sprintf("failed to hydrate mods for pack %s, %w", slug, err))
		}
	}

	return result.Pack, nil
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
	err := ps.db.Where(&tables.Pack{Slug: slug, Status: types.PackStatusPublished}).First(&tables.Pack{}).Error
	return err == nil
}

func (ps *PackwizService) IsPackPublic(slug string) bool {
	err := ps.db.Where(&tables.Pack{Slug: slug, IsPublic: true}).First(&tables.Pack{}).Error
	return err == nil
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
