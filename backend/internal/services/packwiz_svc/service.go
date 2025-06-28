package packwiz_svc

import (
	"fmt"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"time"

	"github.com/leocov-dev/packwiz-nxt/core"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

type PackwizService struct {
	db *gorm.DB
}

func NewPackwizService(db *gorm.DB) *PackwizService {
	return &PackwizService{
		db,
	}
}

func (ps *PackwizService) GetPacksWithPerms(
	request dto.AllPacksQuery,
	userId uint,
) ([]dto.PackResponse, response.ServerError) {
	if len(request.Status) == 0 && !request.Archived {
		request.Status = []types.PackStatus{types.PackStatusDraft, types.PackStatusPublished}
	}

	var results []dto.PackResponse

	query := ps.db.Model(
		&tables.Pack{},
	).Select(
		"packs.*, pack_users.permission as current_user_permission",
	).Preload(
		"User",
	).Joins(
		"LEFT JOIN pack_users ON packs.id = pack_users.pack_id AND pack_users.user_id = ?",
		userId,
	).Order("packs.slug asc")

	if request.Search != "" {
		query = query.Where("packs.slug LIKE ?", "%"+request.Search+"%")
	}

	sub := ps.db
	if len(request.Status) > 0 {
		sub = sub.Where("packs.status IN ?", request.Status)
	}

	if request.Archived {
		sub = sub.Or("packs.deleted_at IS NOT NULL")
	} else {
		sub = sub.Where("packs.deleted_at IS NULL")
	}

	query = query.Where(sub)

	if err := query.Unscoped().Scan(&results).Error; err != nil {
		return nil, response.New(http.StatusInternalServerError, "failed to query db for packs")
	}

	log.Debug(fmt.Sprintf("Found %d packs", len(results)))

	return results, nil
}

func (ps *PackwizService) PackExists(packId uint, includeDeleted bool) bool {
	query := ps.db.Model(tables.Pack{})

	if includeDeleted {
		query = query.Unscoped()
	}

	var exists bool
	if err := query.Select("1").
		Where("id = ?", packId).
		Limit(1).
		Find(&exists).
		Error; err != nil {
		return false
	}

	return exists
}

func (ps *PackwizService) PackExistsBySlug(packSlug string, includeDeleted bool) bool {
	query := ps.db.Model(tables.Pack{})

	if includeDeleted {
		query = query.Unscoped()
	}

	var exists bool
	if err := query.Select("1").
		Where("slug = ?", packSlug).
		Limit(1).
		Find(&exists).
		Error; err != nil {
		return false
	}

	return exists
}

func (ps *PackwizService) NewPack(request dto.NewPackRequest, author tables.User) response.ServerError {

	if ps.PackExistsBySlug(request.Slug, true) {
		return response.New(http.StatusBadRequest, "pack already exists")
	}

	if err := ps.db.Transaction(func(tx *gorm.DB) error {
		newPack := &tables.Pack{
			Slug:                   request.Slug,
			Name:                   request.Name,
			Description:            request.Description,
			CreatedBy:              author.ID,
			UpdatedBy:              author.ID,
			IsPublic:               false,
			Status:                 types.PackStatusDraft,
			MCVersion:              request.MinecraftVersion,
			Loader:                 request.LoaderName,
			LoaderVersion:          request.LoaderVersion,
			AcceptableGameVersions: request.AcceptableVersions,

			Version:    request.Version,
			PackFormat: core.CurrentPackFormat,
		}

		if err := tx.Create(newPack).Error; err != nil {
			return err
		}

		if err := tx.Create(&tables.PackUsers{
			PackID:     newPack.ID,
			UserID:     author.ID,
			Permission: types.PackPermissionEdit,
		}).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return response.Wrap(err)
	}

	return nil
}
func (ps *PackwizService) GetPackById(packId uint) (tables.Pack, response.ServerError) {
	var result tables.Pack

	query := ps.db.Model(
		&tables.Pack{},
	).Preload(
		"Mods",
	).Where(
		&tables.Pack{ID: packId},
	)

	if err := query.Unscoped().First(&result).Error; err != nil {
		return result, response.New(http.StatusNotFound, fmt.Sprintf("pack '%d' not found", packId))
	}

	return result, nil
}
func (ps *PackwizService) GetPackBySlug(slug string) (tables.Pack, response.ServerError) {
	var result tables.Pack

	query := ps.db.Model(
		&tables.Pack{},
	).Preload(
		"Mods",
	).Where(
		&tables.Pack{Slug: slug},
	)

	if err := query.Unscoped().First(&result).Error; err != nil {
		return result, response.New(http.StatusNotFound, fmt.Sprintf("pack '%s' not found", slug))
	}

	return result, nil
}

func (ps *PackwizService) GetPackWithPerms(packId, userId uint) (dto.PackResponse, response.ServerError) {
	var result dto.PackResponse

	query := ps.db.Model(
		&tables.Pack{},
	).Preload(
		"Mods",
	).Select(
		"packs.*, pack_users.permission as current_user_permission",
	).Joins(
		"LEFT JOIN pack_users ON packs.id = pack_users.pack_id AND pack_users.user_id = ?",
		userId,
	).Where(
		"packs.id = ?", packId,
	)

	if err := query.Unscoped().First(&result).Error; err != nil {
		return result, response.New(http.StatusNotFound, fmt.Sprintf("pack '%d' not found", packId))
	}

	return result, nil
}

func (ps *PackwizService) GetMissingModDependencies(packId uint, request dto.AddModRequest) ([]*core.Mod, response.ServerError) {

	dbPack, gErr := ps.GetPackById(packId)
	if gErr != nil {
		return nil, gErr
	}

	pack := dbPack.AsMeta()

	var err error
	var missingDependencies []*core.Mod

	if request.Modrinth != nil {
		missingDependencies, err = lookupModrinthDependencies(request.Modrinth.Url, pack)
	} else if request.Curseforge != nil {
		missingDependencies, err = lookupCurseforgeDependencies(request.Curseforge.Url, pack)
	} else if request.GitHub != nil {
		// can't resolve dependencies for github mods
		return nil, nil
	} else {
		return nil, response.New(http.StatusBadRequest, "invalid mod type")
	}

	if err != nil {
		return nil, response.Wrap(err)
	}

	return missingDependencies, nil
}

// AddMod
// Add a new mod to an existing pack
func (ps *PackwizService) AddMod(packId uint, request dto.AddModRequest, user tables.User) response.ServerError {

	var err error

	dbPack, err := ps.GetPackById(packId)
	if err != nil {
		return response.Wrap(err)
	}

	pack := dbPack.AsMeta()

	var newMod *core.Mod
	var dependencies []*core.Mod

	if request.Modrinth != nil {
		newMod, dependencies, err = addModrinthMod(request.Modrinth.Url, pack)
		if err != nil {
			return response.Wrap(err)
		}
	} else if request.Curseforge != nil {
		newMod, dependencies, err = addCurseforgeMod(request.Curseforge.Url, pack)
		if err != nil {
			return response.Wrap(err)
		}
	} else if request.GitHub != nil {
		newMod, dependencies, err = addGithubMod(request.GitHub.Url, pack)
		if err != nil {
			return response.Wrap(err)
		}
	} else {
		return response.New(http.StatusBadRequest, "invalid mod type")
	}

	if err := ps.db.Transaction(func(tx *gorm.DB) error {

		var dependencyIds []uint

		for _, mod := range dependencies {

			existingMod, existsErr := ps.GetModBySlug(dbPack.Slug, mod.Slug)
			if existsErr == nil {
				dependencyIds = append(dependencyIds, existingMod.ID)
				log.Info(fmt.Sprintf("mod '%s' already exists in pack '%s'", mod.Slug, dbPack.Slug))
				continue
			}

			dbMod, err := createMod(mod, dbPack, user, tx, true, nil)
			if err != nil {
				return err
			}
			dependencyIds = append(dependencyIds, dbMod.ID)
		}

		_, err = createMod(newMod, dbPack, user, tx, false, dependencyIds)

		return nil
	}); err != nil {
		return response.Wrap(err)
	}

	return nil
}

func createMod(mod *core.Mod, dbPack tables.Pack, user tables.User, db *gorm.DB, isDependency bool, dependencyIds []uint) (*tables.Mod, error) {
	source, update := tables.ExtractModSource(mod)
	if source == "" {
		return nil, response.New(http.StatusBadRequest, fmt.Sprintf("invalid mod data found: %v", mod.Update))
	}

	newMod := &tables.Mod{
		Slug:     mod.Slug,
		PackID:   dbPack.ID,
		Name:     mod.Name,
		FileName: mod.FileName,
		Side:     mod.Side,
		Pinned:   mod.Pin,
		Type:     mod.ModType,
		Download: tables.DownloadInfo{
			URL:        mod.Download.URL,
			Mode:       mod.Download.Mode,
			Hash:       mod.Download.Hash,
			HashFormat: mod.Download.HashFormat,
		},
		Source:        source,
		Update:        update,
		CreatedBy:     user.ID,
		UpdatedBy:     user.ID,
		IsDependency:  isDependency,
		DependencyIds: dependencyIds,
	}

	if err := db.Create(newMod).Error; err != nil {
		return nil, err
	}

	return newMod, nil
}

// ArchivePack
// soft-delete a pack
func (ps *PackwizService) ArchivePack(packId uint) response.ServerError {
	if err := ps.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(
			&tables.Pack{ID: packId},
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
		return nil
	}); err != nil {
		return response.New(http.StatusInternalServerError, "failed to archive pack")
	}

	return nil
}

// UnArchivePack
// remove soft delete from a pack
func (ps *PackwizService) UnArchivePack(packId uint) response.ServerError {
	if err := ps.db.Transaction(func(tx *gorm.DB) error {
		return tx.Unscoped().Model(
			&tables.Pack{ID: packId},
		).Update(
			"deleted_at", nil,
		).Error
	}); err != nil {
		return response.New(http.StatusInternalServerError, "failed to unarchive pack")
	}

	return nil
}

// SetPackStatus
// change the pack status
func (ps *PackwizService) SetPackStatus(packId uint, status types.PackStatus) response.ServerError {
	if err := ps.db.Model(&tables.Pack{ID: packId}).Update("status", status).Error; err != nil {
		return response.New(http.StatusInternalServerError, "failed to set pack status")
	}

	return nil
}

func (ps *PackwizService) IsPackPublished(packId uint) bool {
	err := ps.db.Where(&tables.Pack{ID: packId, Status: types.PackStatusPublished}).First(&tables.Pack{}).Error
	return err == nil
}

func (ps *PackwizService) IsPackPublicById(packId uint) bool {
	err := ps.db.Where(&tables.Pack{ID: packId, IsPublic: true}).First(&tables.Pack{}).Error
	return err == nil
}

func (ps *PackwizService) MakePackPublic(packId uint) response.ServerError {
	if err := ps.db.Model(&tables.Pack{ID: packId}).Update("is_public", true).Error; err != nil {
		return response.New(http.StatusInternalServerError, "failed to make pack public")
	}

	return nil
}

func (ps *PackwizService) MakePackPrivate(packId uint) response.ServerError {
	if err := ps.db.Model(&tables.Pack{ID: packId}).Update("is_public", false).Error; err != nil {
		return response.New(http.StatusInternalServerError, "failed to make pack private")
	}

	return nil
}

// SetAcceptableVersions
// set a mod packs acceptable minecraft versions
func (ps *PackwizService) SetAcceptableVersions(packId uint, request dto.SetAcceptableVersionsRequest) response.ServerError {
	if err := ps.db.
		Model(tables.Pack{}).
		Where(tables.Pack{ID: packId}).
		Update("acceptableGameVersions", request.Versions).
		Error; err != nil {
		return response.Wrap(err)
	}

	return nil
}

// UpdateAll
// update all the mods in a pack, skipping pinned mods
func (ps *PackwizService) UpdateAll(packId uint) response.ServerError {

	_, err := ps.GetPackById(packId)
	if err != nil {
		return err
	}

	// TODO: expose update in lib
	//if packwiz_cli.UpdateAll(slug) != nil {
	//	return response.New(http.StatusInternalServerError, "failed to update all mods")
	//}

	return response.New(http.StatusInternalServerError, "not implemented")
}

// ModExistsById
// check if a mod exists in a pack
func (ps *PackwizService) ModExistsById(modId uint) bool {
	var exists bool

	if err := ps.db.
		Model(tables.Mod{}).
		Select("1").
		Where(tables.Mod{ID: modId}).
		Limit(1).
		Find(&exists).
		Error; err != nil {
		return false
	}

	return exists
}

func (ps *PackwizService) ModExistsBySlug(packSlug, modSlug string) bool {
	var count int64
	err := ps.db.
		Model(&tables.Mod{}).
		Joins("JOIN packs ON mods.pack_id = packs.id").
		Where("packs.slug = ? AND mods.slug = ?", packSlug, modSlug).
		Count(&count).Error
	return err == nil && count > 0
}

// RemoveModById
// remove a given mod from a given pack
func (ps *PackwizService) RemoveModById(modId uint) response.ServerError {

	if err := ps.db.
		Model(tables.Mod{}).
		Delete(tables.Mod{ID: modId}).
		Error; err != nil {
		return response.Wrap(err)
	}

	return nil
}

// UpdateMod
// update a given mod from a given pack
func (ps *PackwizService) UpdateMod(modId uint) response.ServerError {
	modInfo, err := ps.GetMod(modId)
	if err != nil {
		return err
	}

	if modInfo.Pinned {
		return response.New(http.StatusBadRequest, "cannot update pinned mod")
	}

	// TODO: expose in lib
	//if packwiz_cli.UpdateOne(slug, mod) != nil {
	//	return response.New(http.StatusInternalServerError, "failed to update mod")
	//}

	return response.New(http.StatusInternalServerError, "not implemented")
}

// GetMod
// get a single mods data
func (ps *PackwizService) GetMod(modId uint) (tables.Mod, response.ServerError) {
	var mod tables.Mod
	if err := ps.db.Where("id = ?", modId).First(&mod).Error; err != nil {
		return mod, response.Wrap(err)
	}

	return mod, nil
}

func (ps *PackwizService) GetModBySlug(packSlug, modSlug string) (tables.Mod, response.ServerError) {
	var mod tables.Mod
	if err := ps.db.
		Model(tables.Mod{}).
		Joins("JOIN packs ON mods.pack_id = packs.id").
		Where("packs.slug = ? AND mods.slug = ?", packSlug, modSlug).
		First(&mod).Error; err != nil {
		return mod, response.Wrap(err)
	}
	return mod, nil
}

func (ps *PackwizService) ChangeModSide(modId uint, side core.ModSide) response.ServerError {
	if err := ps.db.
		Model(tables.Mod{}).
		Where(tables.Mod{ID: modId}).
		Update("side", side).
		Error; err != nil {
		return response.Wrap(err)
	}

	return nil
}

func (ps *PackwizService) SetModPinnedValue(modId uint, value bool) response.ServerError {
	if err := ps.db.
		Model(tables.Mod{}).
		Where(tables.Mod{ID: modId}).
		Update("pinned", value).
		Error; err != nil {
		return response.Wrap(err)
	}

	return nil
}

func (ps *PackwizService) GetPersonalLink(
	user tables.User,
	packId uint,
	scheme string,
	host string,
) (url.URL, response.ServerError) {

	var key string
	if ps.IsPackPublicById(packId) {
		key = "public"
	} else {
		key = user.LinkToken
	}

	pack, err := ps.GetPackById(packId)
	if err != nil {
		return url.URL{}, err
	}

	link, parseErr := url.Parse(fmt.Sprintf("%s://%s/packwiz/%s/%s/pack.toml", scheme, host, key, pack.Slug))
	if parseErr != nil {
		return url.URL{}, response.New(http.StatusInternalServerError, "failed to build link url")
	}

	return *link, nil
}

func (ps *PackwizService) EditPack(packId uint, request dto.EditPackRequest) response.ServerError {

	pack, err := ps.GetPackById(packId)
	if err != nil {
		return response.Wrap(err)
	}

	if request.Name != "" {
		pack.Name = request.Name
	}

	pack.Description = request.Description
	pack.AcceptableGameVersions = request.AcceptableVersions

	if err := ps.db.Save(pack).Error; err != nil {
		return response.Wrap(err)
	}

	return nil
}
