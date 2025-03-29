package packwiz_svc

import (
	"fmt"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"packwiz-web/internal/log"
	"packwiz-web/internal/services/packwiz_cli"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
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
	request dto.AllPacksQuery,
	userId uint,
) ([]tables.Pack, response.ServerError) {
	if len(request.Status) == 0 && !request.Archived {
		request.Status = []types.PackStatus{types.PackStatusDraft, types.PackStatusPublished}
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

func (ps *PackwizService) PackExists(slug string, includeDeleted bool) bool {
	query := ps.db

	if includeDeleted {
		query = query.Unscoped()
	}

	if err := query.Where("slug = ?", slug).First(&tables.Pack{}).Error; err != nil {
		log.Debug("pack not exists:", slug)
		return false
	}

	return true
}

func (ps *PackwizService) NewPack(request dto.NewPackRequest, author tables.User) response.ServerError {

	if ps.PackExists(request.Slug, true) {
		return response.New(http.StatusBadRequest, "pack already exists")
	}

	name := request.Name
	if name == "" {
		name = request.Slug
	}

	if err := ps.db.Transaction(func(tx *gorm.DB) error {
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
	}); err != nil {
		return response.New(http.StatusInternalServerError, "failed to create db pack")
	}

	return nil
}

func (ps *PackwizService) GetPack(slug string, userId uint, hydrateData, hydrateMods bool) (tables.Pack, response.ServerError) {
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
		return tables.Pack{}, response.New(http.StatusNotFound, fmt.Sprintf("pack '%s' not found", slug))
	}

	result.Pack.IsArchived = result.IsArchived
	result.Pack.Permission = result.Permission

	if hydrateData {
		err = ps.hydratePackData(&result.Pack)
		if err != nil {
			log.Warn(fmt.Sprintf("failed to hydrate data for pack %s, %s", slug, err))
		}
	}

	if hydrateMods {
		err = ps.hydrateModData(&result.Pack)
		if err != nil {
			log.Warn(fmt.Sprintf("failed to hydrate mods for pack %s, %s", slug, err))
		}
	}

	return result.Pack, nil
}

// AddMod
// Add a new mod to an existing pack
func (ps *PackwizService) AddMod(slug string, request dto.AddModRequest) response.ServerError {
	if request.Modrinth.IsSet() {
		data := request.Modrinth
		return response.Wrap(packwiz_cli.AddModrinthMod(slug, data.Name, data.ProjectId, data.VersionFilename, data.VersionId))
	} else if request.Curseforge.IsSet() {
		data := request.Curseforge
		return response.Wrap(packwiz_cli.AddCurseforgeMod(slug, data.Name, data.AddonId, data.Category, data.FileId, data.Game))
	}

	return response.New(http.StatusBadRequest, "invalid add mod request")
}

// ArchivePack
// soft delete a pack
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
	if err := packwiz_cli.SetAcceptableVersions(slug, request.Versions...); err != nil {
		return response.New(http.StatusInternalServerError, "failed to set acceptable versions")
	}

	return nil
}

// UpdateAll
// update all the mods in a pack, skipping pinned mods
func (ps *PackwizService) UpdateAll(slug string) response.ServerError {
	if packwiz_cli.UpdateAll(slug) != nil {
		return response.New(http.StatusInternalServerError, "failed to update all mods")
	}

	return nil
}

// ModExists
// check if a mod exists in a pack
func (ps *PackwizService) ModExists(slug, mod string) bool {
	return packwiz_cli.ModExists(slug, mod) == nil
}

// RemoveMod
// remove a given mod from a given pack
func (ps *PackwizService) RemoveMod(slug, mod string) response.ServerError {
	if packwiz_cli.RemoveMod(slug, mod) != nil {
		return response.New(http.StatusInternalServerError, "failed to remove mod")
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

	if packwiz_cli.UpdateOne(slug, mod) != nil {
		return response.New(http.StatusInternalServerError, "failed to update mod")
	}

	return nil
}

// GetMod
// get a single mods data
func (ps *PackwizService) GetMod(slug, mod string) (*types.ModData, response.ServerError) {
	data, err := findModData(slug, mod)
	if err != nil {
		return nil, response.New(http.StatusInternalServerError, "failed to get mod file data")
	}

	return &data, nil
}

func (ps *PackwizService) ChangeModSide(slug, mod string, side types.ModSide) response.ServerError {
	if packwiz_cli.ChangeModSide(slug, mod, side) != nil {
		return response.New(http.StatusInternalServerError, "failed to change mod side")
	}

	return nil
}

// PinMod
// pin a mod to prevent it from being updated
func (ps *PackwizService) PinMod(slug, mod string) response.ServerError {
	if packwiz_cli.PinMod(slug, mod) != nil {
		return response.New(http.StatusInternalServerError, "failed to pin mod")
	}

	return nil
}

// UnpinMod
// unpin a mod to allow it to be updated
func (ps *PackwizService) UnpinMod(slug, mod string) response.ServerError {
	if packwiz_cli.UnpinMod(slug, mod) != nil {
		return response.New(http.StatusInternalServerError, "failed to unpin mod")
	}

	return nil
}

func (ps *PackwizService) GetPersonalLink(
	user tables.User,
	slug string,
	scheme string,
	host string,
) (url.URL, response.ServerError) {
	link, err := url.Parse(fmt.Sprintf("%s://%s/packwiz/%s/pack.toml", scheme, host, slug))
	if err != nil {
		return url.URL{}, response.New(http.StatusInternalServerError, "failed to build link url")
	}

	if !ps.IsPackPublic(slug) {
		query := link.Query()
		query.Add("token", user.LinkToken)
		link.RawQuery = query.Encode()

	}

	return *link, nil
}

// -----------------------------------------------------------------------------

func (ps *PackwizService) hydratePackData(pack *tables.Pack) response.ServerError {
	var err error
	pack.PackData, err = getModpackData(pack.Slug)

	if err != nil {
		return response.New(http.StatusInternalServerError, "failed to get pack file data")
	}

	return nil
}

func (ps *PackwizService) hydrateModData(pack *tables.Pack) response.ServerError {
	var err error
	pack.ModData, err = getModpackMods(pack.Slug)

	if err != nil {
		return response.New(http.StatusInternalServerError, "failed to get pack mods file data")
	}

	return nil
}
