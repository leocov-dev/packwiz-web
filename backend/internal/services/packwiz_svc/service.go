package packwiz_svc

import (
	"fmt"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
	"time"

	"github.com/leocov-dev/packwiz-nxt/core"
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
	request dto.AllPacksQuery,
	userId uint,
) ([]tables.Pack, response.ServerError) {
	if len(request.Status) == 0 && !request.Archived {
		request.Status = []types.PackStatus{types.PackStatusDraft, types.PackStatusPublished}
	}

	type Result struct {
		tables.Pack
		Author     string
		IsArchived bool
	}
	var results []Result

	query := ps.db.Model(
		&tables.Pack{},
	).Select(
		"packs.*, creator.full_name AS full_name, current_user.permission AS permission, (deleted_at IS NOT NULL) AS is_archived",
	).Joins(
		"LEFT JOIN pack_users AS creator ON packs.slug = creator.pack_slug AND creator.user_id = packs.created_by",
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
		sub = sub.Or("deleted_at IS NOT NULL")
	} else {
		sub = sub.Where("deleted_at IS NULL")
	}

	query = query.Where(sub)

	if err := query.Unscoped().Scan(&results).Error; err != nil {
		return nil, response.New(http.StatusInternalServerError, "failed to query db for packs")
	}

	// transform results to raw packs since GORM handles virtual columns weirdly
	packs := make([]tables.Pack, len(results))
	for i, result := range results {
		pack := result.Pack
		pack.Author = result.Author
		pack.IsArchived = result.IsArchived
		packs[i] = pack
	}

	log.Info(fmt.Sprintf("Found %d packs", len(packs)))

	return packs, nil
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

	return true
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
			MCVersion:              request.MinecraftDef.Version,
			Loader:                 request.LoaderDef.Name,
			LoaderVersion:          request.LoaderDef.Version,
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
	type Result struct {
		tables.Pack
		IsArchived bool
	}
	var result Result

	query := ps.db.Model(
		&tables.Pack{},
	).Select(
		"packs.*, (deleted_at IS NOT NULL) AS is_archived",
	).Where(
		"packs.slug = ?", slug,
	).Order("packs.slug asc")

	if err := query.Unscoped().First(&result).Error; err != nil {
		return tables.Pack{}, response.New(http.StatusNotFound, fmt.Sprintf("pack '%s' not found", slug))
	}

	var mods []tables.Mod
	if err := ps.db.Where("pack_slug = ?", slug).Find(&mods).Error; err != nil {
		return tables.Pack{}, response.New(http.StatusInternalServerError, "failed to retrieve mods for pack")
	}

	for _, mod := range mods {
		mod.SourceLink = getModSourceLink(mod.Source, mod.ModKey, mod.VersionKey)
	}

	result.Pack.Mods = mods
	result.Pack.IsArchived = result.IsArchived

	return result.Pack, nil
}

// AddMod
// Add a new mod to an existing pack
func (ps *PackwizService) AddMod(slug string, request dto.AddModRequest, user tables.User) response.ServerError {

	//if err := ps.db.Transaction(func(tx *gorm.DB) error {
	//	return tx.Create(&tables.Mod{
	//		PackSlug:    slug,
	//		Name:        modInfo.Name,
	//		DisplayName: modInfo.DisplayName,
	//		FileName:    modInfo.Filename,
	//		Side:        modInfo.Side,
	//		Pinned:      modInfo.Pinned,
	//
	//		DownloadUrl:        modInfo.DownloadUrl,
	//		DownloadMode:       modInfo.DownloadMode,
	//		DownloadHash:       modInfo.DownloadHash,
	//		DownloadHashFormat: modInfo.DownloadHashFormat,
	//
	//		Source:     modInfo.Source,
	//		ModKey:     modInfo.ModKey,
	//		VersionKey: modInfo.VersionKey,
	//
	//		CreatedBy: user.Id,
	//		UpdatedBy: user.Id,
	//	}).Error
	//}); err != nil {
	//	return response.Wrap(err)
	//}

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

	mod.SourceLink = getModSourceLink(mod.Source, mod.ModKey, mod.VersionKey)

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
