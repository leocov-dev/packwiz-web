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
		"LEFT JOIN pack_users ON packs.slug = pack_users.pack_slug AND pack_users.user_id = ?",
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

func (ps *PackwizService) PackExists(slug string, includeDeleted bool) bool {
	query := ps.db.Model(tables.Pack{})

	if includeDeleted {
		query = query.Unscoped()
	}

	var exists bool
	if err := query.Select("1").
		Where("slug = ?", slug).
		Limit(1).
		Find(&exists).
		Error; err != nil {
		return false
	}

	return exists
}

func (ps *PackwizService) NewPack(request dto.NewPackRequest, author tables.User) response.ServerError {

	if ps.PackExists(request.Slug, true) {
		return response.New(http.StatusBadRequest, "pack already exists")
	}

	if err := ps.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&tables.Pack{
			Slug:                   request.Slug,
			Name:                   request.Name,
			Description:            request.Description,
			CreatedBy:              author.Id,
			UpdatedBy:              author.Id,
			IsPublic:               false,
			Status:                 types.PackStatusDraft,
			MCVersion:              request.MinecraftVersion,
			Loader:                 request.LoaderName,
			LoaderVersion:          request.LoaderVersion,
			AcceptableGameVersions: request.AcceptableVersions,

			Version:    request.Version,
			PackFormat: core.CurrentPackFormat,
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

		return nil
	}); err != nil {
		return response.Wrap(err)
	}

	return nil
}

func (ps *PackwizService) GetPack(slug string) (tables.Pack, response.ServerError) {
	var result tables.Pack

	query := ps.db.Model(
		&tables.Pack{},
	).Preload(
		"Mods",
	).Where(
		"packs.slug = ?", slug,
	).Order("packs.slug asc")

	if err := query.Unscoped().First(&result).Error; err != nil {
		return result, response.New(http.StatusNotFound, fmt.Sprintf("pack '%s' not found", slug))
	}

	return result, nil
}

func (ps *PackwizService) GetPackWithPerms(slug string, userId uint) (dto.PackResponse, response.ServerError) {
	var result dto.PackResponse

	query := ps.db.Model(
		&tables.Pack{},
	).Preload(
		"Mods",
	).Select(
		"packs.*, pack_users.permission as current_user_permission",
	).Joins(
		"LEFT JOIN pack_users ON packs.slug = pack_users.pack_slug AND pack_users.user_id = ?",
		userId,
	).Where(
		"packs.slug = ?", slug,
	).Order("packs.slug asc")

	if err := query.Unscoped().First(&result).Error; err != nil {
		return result, response.New(http.StatusNotFound, fmt.Sprintf("pack '%s' not found", slug))
	}

	return result, nil
}

func (ps *PackwizService) GetMissingModDependencies(packSlug string, request dto.AddModRequest) ([]*core.Mod, response.ServerError) {
	var err error

	dbPack, err := ps.GetPack(packSlug)
	if err != nil {
		return nil, response.Wrap(err)
	}

	pack := dbPack.AsMeta()

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

	var missing []*core.Mod
	for _, mod := range missingDependencies {
		missing = append(missing, mod)
	}

	return missing, nil
}

// AddMod
// Add a new mod to an existing pack
func (ps *PackwizService) AddMod(packSlug string, request dto.AddModRequest, user tables.User) response.ServerError {

	var err error

	dbPack, err := ps.GetPack(packSlug)
	if err != nil {
		return response.Wrap(err)
	}

	pack := dbPack.AsMeta()

	var newMods []*core.Mod

	if request.Modrinth != nil {
		newMods, err = addModrinthMod(request.Modrinth.Url, pack)
		if err != nil {
			return response.Wrap(err)
		}
	} else if request.Curseforge != nil {
		newMods, err = addCurseforgeMod(request.Curseforge.Url, pack)
		if err != nil {
			return response.Wrap(err)
		}
	} else if request.GitHub != nil {
		newMods, err = addGithubMod(request.GitHub.Url, pack)
		if err != nil {
			return response.Wrap(err)
		}
	} else {
		return response.New(http.StatusBadRequest, "invalid mod type")
	}

	if err := ps.db.Transaction(func(tx *gorm.DB) error {

		for _, mod := range newMods {

			source, update := tables.ExtractModSource(mod)
			if source == "" {
				return response.New(http.StatusBadRequest, fmt.Sprintf("invalid mod data found: %v", mod.Update))
			}

			if err := tx.Create(&tables.Mod{
				Slug:     mod.Slug,
				PackSlug: packSlug,
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
				Source:    source,
				Update:    update,
				CreatedBy: user.Id,
				UpdatedBy: user.Id,
			}).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return response.Wrap(err)
	}

	return nil
}

// ArchivePack
// soft-delete a pack
func (ps *PackwizService) ArchivePack(slug string) response.ServerError {
	if err := ps.db.Transaction(func(tx *gorm.DB) error {
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
		return nil
	}); err != nil {
		return response.New(http.StatusInternalServerError, "failed to archive pack")
	}

	return nil
}

// UnArchivePack
// remove soft delete from a pack
func (ps *PackwizService) UnArchivePack(slug string) response.ServerError {
	if err := ps.db.Transaction(func(tx *gorm.DB) error {
		return tx.Unscoped().Model(
			&tables.Pack{Slug: slug},
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
func (ps *PackwizService) SetPackStatus(slug string, status types.PackStatus) response.ServerError {
	if err := ps.db.Model(&tables.Pack{Slug: slug}).Update("status", status).Error; err != nil {
		return response.New(http.StatusInternalServerError, "failed to set pack status")
	}

	return nil
}

func (ps *PackwizService) IsPackPublished(slug string) bool {
	err := ps.db.Where(&tables.Pack{Slug: slug, Status: types.PackStatusPublished}).First(&tables.Pack{}).Error
	return err == nil
}

func (ps *PackwizService) IsPackPublic(slug string) bool {
	err := ps.db.Where(&tables.Pack{Slug: slug, IsPublic: true}).First(&tables.Pack{}).Error
	return err == nil
}

func (ps *PackwizService) MakePackPublic(slug string) response.ServerError {
	if err := ps.db.Model(&tables.Pack{Slug: slug}).Update("is_public", true).Error; err != nil {
		return response.New(http.StatusInternalServerError, "failed to make pack public")
	}

	return nil
}

func (ps *PackwizService) MakePackPrivate(slug string) response.ServerError {
	if err := ps.db.Model(&tables.Pack{Slug: slug}).Update("is_public", false).Error; err != nil {
		return response.New(http.StatusInternalServerError, "failed to make pack private")
	}

	return nil
}

// SetAcceptableVersions
// set a mod packs acceptable minecraft versions
func (ps *PackwizService) SetAcceptableVersions(slug string, request dto.SetAcceptableVersionsRequest) response.ServerError {
	if err := ps.db.
		Model(tables.Pack{}).
		Where(tables.Pack{Slug: slug}).
		Update("acceptableGameVersions", request.Versions).
		Error; err != nil {
		return response.Wrap(err)
	}

	return nil
}

// UpdateAll
// update all the mods in a pack, skipping pinned mods
func (ps *PackwizService) UpdateAll(slug string) response.ServerError {

	// TODO: expose update in lib
	//if packwiz_cli.UpdateAll(slug) != nil {
	//	return response.New(http.StatusInternalServerError, "failed to update all mods")
	//}

	return nil
}

// ModExists
// check if a mod exists in a pack
func (ps *PackwizService) ModExists(slug, mod string) bool {
	var exists bool
	if err := ps.db.
		Model(tables.Mod{}).
		Select("1").
		Where(tables.Mod{Slug: mod, PackSlug: slug}).
		Limit(1).
		Find(&exists).
		Error; err != nil {
		return false
	}

	return exists
}

// RemoveMod
// remove a given mod from a given pack
func (ps *PackwizService) RemoveMod(slug, mod string) response.ServerError {

	if err := ps.db.
		Model(tables.Mod{}).
		Delete(tables.Mod{PackSlug: slug, Slug: mod}).
		Error; err != nil {
		return response.Wrap(err)
	}

	return nil
}

// UpdateMod
// update a given mod from a given pack
func (ps *PackwizService) UpdateMod(slug, mod string) response.ServerError {
	modInfo, err := ps.GetMod(slug, mod)
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

	return nil
}

// GetMod
// get a single mods data
func (ps *PackwizService) GetMod(slug, name string) (tables.Mod, response.ServerError) {
	var mod tables.Mod
	if err := ps.db.Where("pack_slug = ? AND name = ?", slug, name).First(&mod).Error; err != nil {
		return mod, response.Wrap(err)
	}

	return mod, nil
}

func (ps *PackwizService) ChangeModSide(slug, mod string, side core.ModSide) response.ServerError {
	if err := ps.db.
		Model(tables.Mod{}).
		Where(tables.Mod{Slug: mod, PackSlug: slug}).
		Update("side", side).
		Error; err != nil {
		return response.Wrap(err)
	}

	return nil
}

func (ps *PackwizService) SetModPinnedValue(slug, mod string, value bool) response.ServerError {
	if err := ps.db.
		Model(tables.Mod{}).
		Where(tables.Mod{Slug: mod, PackSlug: slug}).
		Update("pinned", value).
		Error; err != nil {
		return response.Wrap(err)
	}

	return nil
}

func (ps *PackwizService) GetPersonalLink(
	user tables.User,
	slug string,
	scheme string,
	host string,
) (url.URL, response.ServerError) {

	var key string
	if ps.IsPackPublic(slug) {
		key = "public"
	} else {
		key = user.LinkToken
	}

	link, err := url.Parse(fmt.Sprintf("%s://%s/packwiz/%s/%s/pack.toml", scheme, host, key, slug))
	if err != nil {
		return url.URL{}, response.New(http.StatusInternalServerError, "failed to build link url")
	}

	return *link, nil
}

func (ps *PackwizService) EditPack(slug string, request dto.EditPackRequest) response.ServerError {

	pack, err := ps.GetPack(slug)
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
