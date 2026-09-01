package packwiz_svc

import (
	"context"
	"database/sql"
	"fmt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/leocov-dev/packwiz-nxt/core"
	"github.com/leocov-dev/packwiz-nxt/fileio"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"packwiz-web/internal/jobs"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

type PackwizService struct {
	db *gorm.DB
	// riverClient is used to enqueue background jobs (e.g. MigrateModsArgs).
	// It's nil for PackwizService instances constructed as a job resolver
	// (internal/jobs.MigrateModsResolver) - those never need to enqueue.
	riverClient *river.Client[*sql.Tx]
}

func NewPackwizService(db *gorm.DB, riverClient *river.Client[*sql.Tx]) *PackwizService {
	return &PackwizService{
		db:          db,
		riverClient: riverClient,
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

// filterValidDependencyIds trims each mod's DependencyIds to only reference
// mod IDs present in the given slice (i.e. other mods in the same pack).
// Mutates and returns the input slice. Guards against stale references left
// behind when a mod a dependency pointed at is later removed.
func filterValidDependencyIds(mods []tables.Mod) []tables.Mod {
	valid := make(map[uint]bool, len(mods))
	for _, m := range mods {
		valid[m.ID] = true
	}
	for i := range mods {
		filtered := make([]uint, 0, len(mods[i].DependencyIds))
		for _, id := range mods[i].DependencyIds {
			if valid[id] {
				filtered = append(filtered, id)
			}
		}
		mods[i].DependencyIds = filtered
	}
	return mods
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

	result.Mods = filterValidDependencyIds(result.Mods)

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

	result.Mods = filterValidDependencyIds(result.Mods)

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

	result.Mods = filterValidDependencyIds(result.Mods)

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

		if _, err := createMod(newMod, dbPack, user, tx, false, dependencyIds); err != nil {
			return err
		}

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
		).Select(
			"IsPublic", "Status", "DeletedAt",
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
		Update("acceptable_game_versions", request.Versions).
		Error; err != nil {
		return response.Wrap(err)
	}

	return nil
}

// UpdateAll
// update all the mods in a pack, skipping pinned mods
func (ps *PackwizService) UpdateAll(packId uint, user tables.User) response.ServerError {

	dbPack, err := ps.GetPackById(packId)
	if err != nil {
		return err
	}

	pack := dbPack.AsMeta()

	if updateErr := core.UpdateAllMods(nil, pack); updateErr != nil {
		return response.Wrap(updateErr)
	}

	if txErr := ps.db.Transaction(func(tx *gorm.DB) error {
		for _, dbMod := range dbPack.Mods {
			updatedMod, ok := pack.Mods[dbMod.Slug]
			if !ok {
				continue
			}
			if err := applyModUpdate(tx, dbMod.ID, updatedMod, user); err != nil {
				return err
			}
		}
		return nil
	}); txErr != nil {
		return response.Wrap(txErr)
	}

	return nil
}

// RehashAll
// recompute and persist a single hash format for every mod in a pack
func (ps *PackwizService) RehashAll(ctx context.Context, packId uint, format string, user tables.User) response.ServerError {
	dbPack, err := ps.GetPackById(packId)
	if err != nil {
		return err
	}

	pack := dbPack.AsMeta()

	session, sessErr := fileio.CreateDownloadSession(nil, pack.GetModsList(), []string{format})
	if sessErr != nil {
		return response.Wrap(sessErr)
	}

	for dl := range session.StartDownloads(ctx) {
		if dl.Error != nil {
			// leave mods that fail to rehash (e.g. manual-download mods) untouched
			continue
		}
		dl.Mod.Download.HashFormat = format
		dl.Mod.Download.Hash = dl.Hashes[format]
	}

	if idxErr := session.SaveIndex(); idxErr != nil {
		// non-fatal: this is a shared local download cache, not persisted pack state
		log.Debug(idxErr)
	}

	if txErr := ps.db.Transaction(func(tx *gorm.DB) error {
		for _, dbMod := range dbPack.Mods {
			updatedMod, ok := pack.Mods[dbMod.Slug]
			if !ok || updatedMod.Download.Hash == "" {
				continue
			}
			if err := tx.Model(&tables.Mod{ID: dbMod.ID}).Select(
				"Download", "HashFormat", "UpdatedBy",
			).Updates(tables.Mod{
				Download: tables.DownloadInfo{
					URL:        updatedMod.Download.URL,
					Mode:       updatedMod.Download.Mode,
					Hash:       updatedMod.Download.Hash,
					HashFormat: updatedMod.Download.HashFormat,
				},
				HashFormat: format,
				UpdatedBy:  user.ID,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	}); txErr != nil {
		return response.Wrap(txErr)
	}

	return nil
}

// applyModUpdate
// persist the fields of an updated core.Mod back onto its corresponding tables.Mod row
func applyModUpdate(tx *gorm.DB, dbModID uint, updated *core.Mod, user tables.User) error {
	source, update := tables.ExtractModSource(updated)

	return tx.Model(&tables.Mod{ID: dbModID}).Select(
		"FileName", "Download", "Source", "Update", "UpdatedBy",
	).Updates(tables.Mod{
		FileName: updated.FileName,
		Download: tables.DownloadInfo{
			URL:        updated.Download.URL,
			Mode:       updated.Download.Mode,
			Hash:       updated.Download.Hash,
			HashFormat: updated.Download.HashFormat,
		},
		Source:    source,
		Update:    update,
		UpdatedBy: user.ID,
	}).Error
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
func (ps *PackwizService) UpdateMod(modId uint, user tables.User) response.ServerError {
	modInfo, err := ps.GetMod(modId)
	if err != nil {
		return err
	}

	if modInfo.Pinned {
		return response.New(http.StatusBadRequest, "cannot update pinned mod")
	}

	dbPack, err := ps.GetPackById(modInfo.PackID)
	if err != nil {
		return err
	}

	pack := dbPack.AsMeta()

	mod, ok := pack.Mods[modInfo.Slug]
	if !ok {
		return response.New(http.StatusNotFound, fmt.Sprintf("mod '%s' not found in pack", modInfo.Slug))
	}

	if updateErr := core.UpdateSingleMod(nil, pack, mod); updateErr != nil {
		return response.Wrap(updateErr)
	}

	if txErr := ps.db.Transaction(func(tx *gorm.DB) error {
		return applyModUpdate(tx, modInfo.ID, mod, user)
	}); txErr != nil {
		return response.Wrap(txErr)
	}

	return nil
}

// GetMod
// get a single mods data
func (ps *PackwizService) GetMod(modId uint) (tables.Mod, response.ServerError) {
	var mod tables.Mod
	if err := ps.db.Where("id = ?", modId).First(&mod).Error; err != nil {
		return mod, response.Wrap(err)
	}

	if err := ps.filterModDependencyIds(&mod); err != nil {
		return mod, response.Wrap(err)
	}

	return mod, nil
}

// filterModDependencyIds trims mod.DependencyIds to only reference mod IDs
// that still exist, for the single-mod case where sibling mods aren't
// already loaded (see filterValidDependencyIds for the in-memory variant
// used when a pack's full Mods slice is available).
func (ps *PackwizService) filterModDependencyIds(mod *tables.Mod) error {
	if len(mod.DependencyIds) == 0 {
		return nil
	}

	var existing []uint
	if err := ps.db.Model(&tables.Mod{}).
		Where("id IN ?", []uint(mod.DependencyIds)).
		Pluck("id", &existing).Error; err != nil {
		return err
	}

	valid := make(map[uint]bool, len(existing))
	for _, id := range existing {
		valid[id] = true
	}

	filtered := make([]uint, 0, len(mod.DependencyIds))
	for _, id := range mod.DependencyIds {
		if valid[id] {
			filtered = append(filtered, id)
		}
	}
	mod.DependencyIds = filtered

	return nil
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

func (ps *PackwizService) ChangeModOption(modId uint, req dto.ChangeModOptionRequest) response.ServerError {
	option := tables.OptionInfo{
		Optional:    req.Optional,
		Description: req.Description,
		Default:     req.Default,
	}

	if err := ps.db.
		Model(tables.Mod{}).
		Where(tables.Mod{ID: modId}).
		Update("option", option).
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

// PackUserInfo
// a user's access information for a given pack
type PackUserInfo struct {
	UserID     uint                 `json:"userId"`
	Username   string               `json:"username"`
	FullName   string               `json:"fullName"`
	Email      string               `json:"email"`
	Permission types.PackPermission `json:"permission"`
	CreatedAt  time.Time            `json:"createdAt"`
}

// ListPackUsers
// list all users with access to a pack
func (ps *PackwizService) ListPackUsers(packId uint) ([]PackUserInfo, response.ServerError) {
	var results []PackUserInfo

	if err := ps.db.Model(&tables.PackUsers{}).
		Select(
			"pack_users.user_id, users.username, users.full_name, users.email, pack_users.permission, pack_users.created_at",
		).
		Joins("JOIN users ON users.id = pack_users.user_id").
		Where("pack_users.pack_id = ?", packId).
		Order("users.username asc").
		Scan(&results).Error; err != nil {
		return nil, response.New(http.StatusInternalServerError, "failed to query db for pack users")
	}

	return results, nil
}

// PackUserSearchResult
// a minimal user shape for the "add collaborator" search picker
type PackUserSearchResult struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

// SearchPackUsers
// search for users by username/full name/email who do not already have access to a pack
func (ps *PackwizService) SearchPackUsers(packId uint, query string) ([]PackUserSearchResult, response.ServerError) {
	var results []PackUserSearchResult

	like := "%" + query + "%"
	if err := ps.db.Model(&tables.User{}).
		Select("users.id as user_id, users.username, users.full_name, users.email").
		Where("users.username ILIKE ? OR users.full_name ILIKE ? OR users.email ILIKE ?", like, like, like).
		Where("users.id NOT IN (SELECT user_id FROM pack_users WHERE pack_id = ?)", packId).
		Order("users.username asc").
		Limit(20).
		Scan(&results).Error; err != nil {
		return nil, response.New(http.StatusInternalServerError, "failed to query db for users")
	}

	return results, nil
}

// GrantPackUser
// grant a user access to a pack
func (ps *PackwizService) GrantPackUser(packId, userId uint, permission types.PackPermission) response.ServerError {
	if _, err := ps.GetPackById(packId); err != nil {
		return err
	}

	var userExists bool
	if err := ps.db.Model(&tables.User{}).
		Select("1").
		Where("id = ?", userId).
		Limit(1).
		Find(&userExists).Error; err != nil {
		return response.New(http.StatusInternalServerError, "failed to query db for user")
	}
	if !userExists {
		return response.New(http.StatusNotFound, fmt.Sprintf("user '%d' not found", userId))
	}

	var alreadyExists bool
	if err := ps.db.Model(&tables.PackUsers{}).
		Select("1").
		Where("pack_id = ? AND user_id = ?", packId, userId).
		Limit(1).
		Find(&alreadyExists).Error; err != nil {
		return response.New(http.StatusInternalServerError, "failed to query db for pack user")
	}
	if alreadyExists {
		return response.New(http.StatusConflict, "user already has access to this pack, use edit instead")
	}

	if err := ps.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&tables.PackUsers{
			PackID:     packId,
			UserID:     userId,
			Permission: permission,
		}).Error
	}); err != nil {
		return response.Wrap(err)
	}

	return nil
}

// RevokePackUser
// revoke a user's access to a pack
func (ps *PackwizService) RevokePackUser(packId, userId uint) response.ServerError {
	result := ps.db.Where("pack_id = ? AND user_id = ?", packId, userId).Delete(&tables.PackUsers{})
	if result.Error != nil {
		return response.Wrap(result.Error)
	}
	if result.RowsAffected == 0 {
		return response.New(http.StatusNotFound, "user does not have access to this pack")
	}

	return nil
}

// ChangePackUserPermission
// change a user's permission level for a pack
func (ps *PackwizService) ChangePackUserPermission(packId, userId uint, permission types.PackPermission) response.ServerError {
	result := ps.db.Model(&tables.PackUsers{}).
		Where("pack_id = ? AND user_id = ?", packId, userId).
		Update("permission", permission)
	if result.Error != nil {
		return response.Wrap(result.Error)
	}
	if result.RowsAffected == 0 {
		return response.New(http.StatusNotFound, "user does not have access to this pack")
	}

	return nil
}

func (ps *PackwizService) EditPack(packId uint, request dto.EditPackRequest) response.ServerError {

	pack, err := ps.GetPackById(packId)
	if err != nil {
		return response.Wrap(err)
	}

	if request.Name != "" {
		pack.Name = request.Name
	}

	if request.Version != "" {
		pack.Version = request.Version
	}

	if request.Description != "" {
		pack.Description = request.Description
	}

	mcVersion, mcErr := resolveMinecraftVersion(request.MinecraftDef, false)
	if mcErr != nil {
		return mcErr
	}
	if mcVersion != "" {
		pack.MCVersion = mcVersion
	}

	if loaderName := strings.ToLower(request.LoaderDef.Name); loaderName != "" {
		pack.Loader = loaderName
	}

	loaderVersion, lvErr := resolveLoaderVersion(pack.Loader, pack.MCVersion, request.LoaderDef, false)
	if lvErr != nil {
		return lvErr
	}
	if loaderVersion != "" {
		pack.LoaderVersion = loaderVersion
	}

	if len(request.AcceptableVersions) > 0 {
		pack.AcceptableGameVersions = request.AcceptableVersions
	}

	if err := ps.db.Save(pack).Error; err != nil {
		return response.Wrap(err)
	}

	return nil
}

// resolveMinecraftVersion
// turns a MinecraftDef into a concrete Minecraft version string. Returns "" (no
// error) when the def carries no version/latest/snapshot, signaling "leave
// unchanged" for partial-update callers such as EditPack. Latest/snapshot always
// require a live fetch of the Mojang version manifest to resolve; validateExplicit
// additionally validates a literal Version against that manifest (skipped for
// EditPack to avoid a network round-trip on every plain metadata edit, since that
// path predates real version validation; used for Migrate, which is a deliberate,
// infrequent action where validation is worth the cost).
func resolveMinecraftVersion(def dto.MinecraftDef, validateExplicit bool) (string, response.ServerError) {
	if !def.Latest && !def.Snapshot && def.Version == "" {
		return "", nil
	}

	if !def.Latest && !def.Snapshot && !validateExplicit {
		return def.Version, nil
	}

	mcv, err := core.GetMinecraftVersions()
	if err != nil {
		return "", response.Wrap(err)
	}

	switch {
	case def.Snapshot:
		return mcv.LatestSnapshot, nil
	case def.Latest:
		return mcv.Latest, nil
	default:
		if !mcv.CheckValid(def.Version) {
			return "", response.New(http.StatusBadRequest, fmt.Sprintf("'%s' is not a valid minecraft version", def.Version))
		}
		return def.Version, nil
	}
}

// resolveLoaderVersion
// turns a LoaderDef (plus useRecommended, meaningful for forge only) into a
// concrete loader version for the given loader name and Minecraft version.
// Returns "" (no error) when the def carries no version/latest and useRecommended
// is false, signaling "leave unchanged" for partial-update callers such as
// EditPack.
func resolveLoaderVersion(loaderName, mcVersion string, def dto.LoaderDef, useRecommended bool) (string, response.ServerError) {
	if !def.Latest && !useRecommended && def.Version == "" {
		return "", nil
	}

	loaderComp, ok := core.ModLoaders[loaderName]
	if !ok {
		return "", response.New(http.StatusBadRequest, fmt.Sprintf("unknown loader '%s'", loaderName))
	}

	versions, latest, err := loaderComp.VersionListGetter(mcVersion)
	if err != nil {
		return "", response.Wrap(err)
	}

	switch {
	case loaderName == "forge" && useRecommended:
		recommended, recErr := core.GetForgeRecommended(mcVersion)
		if recErr != nil {
			return "", response.Wrap(recErr)
		}
		if recommended != "" {
			return recommended, nil
		}
		return latest, nil
	case def.Latest || useRecommended:
		return latest, nil
	default:
		// liteloader has exactly one version per Minecraft version and isn't
		// represented in its own version list, so it's exempt from containment
		// validation (matches packwiz-nxt's CLI migrate behavior).
		if loaderName != "liteloader" && !slices.Contains(versions, def.Version) {
			return "", response.New(http.StatusBadRequest, fmt.Sprintf("'%s' is not a valid %s version for minecraft '%s'", def.Version, loaderName, mcVersion))
		}
		return def.Version, nil
	}
}

// migrationTarget is the resolved outcome of a MigratePackRequest: the
// concrete Minecraft/loader versions it names, plus an in-memory core.Pack
// (built from the pack's current mods) with those versions already applied,
// ready for a compatibility check or a cascading mod update. Building this
// performs no persistence.
type migrationTarget struct {
	Pack          core.Pack
	MCVersion     string
	LoaderName    string
	LoaderVersion string
}

// resolveMigrationTarget resolves request against dbPack's current state into
// a migrationTarget, shared by Migrate (which may persist it) and
// MigrateDryRun (which only checks it).
func resolveMigrationTarget(dbPack tables.Pack, request dto.MigratePackRequest) (migrationTarget, response.ServerError) {
	mcVersion, mcErr := resolveMinecraftVersion(request.MinecraftDef, true)
	if mcErr != nil {
		return migrationTarget{}, mcErr
	}
	if mcVersion == "" {
		return migrationTarget{}, response.New(http.StatusBadRequest, "minecraft version or latest/snapshot flag is required")
	}

	loaderName := strings.ToLower(request.LoaderDef.Name)
	if loaderName == "" {
		loaderName = dbPack.Loader
	}

	loaderVersion, lvErr := resolveLoaderVersion(loaderName, mcVersion, request.LoaderDef, request.UseRecommended)
	if lvErr != nil {
		return migrationTarget{}, lvErr
	}
	if loaderVersion == "" {
		return migrationTarget{}, response.New(http.StatusBadRequest, "loader version, latest, or recommended flag is required")
	}

	pack := dbPack.AsMeta()
	pack.Versions = map[string]string{
		"minecraft": mcVersion,
		loaderName:  loaderVersion,
	}

	if len(request.AcceptableVersions) > 0 {
		pack.SetAcceptableGameVersions(request.AcceptableVersions)
	}

	return migrationTarget{
		Pack:          pack,
		MCVersion:     mcVersion,
		LoaderName:    loaderName,
		LoaderVersion: loaderVersion,
	}, nil
}

// Migrate
// validates and applies a new Minecraft version / loader combination to a
// pack, persisting the version change immediately. If request.UpdateMods is
// set, mod re-resolution is not done inline (it makes one sequential
// external HTTP call per mod) - instead a MigrateModsArgs job is enqueued to
// do that work in the background, mirroring packwiz-nxt's CLI `migrate
// minecraft`/`migrate loader` flow but decoupled from the request lifecycle.
func (ps *PackwizService) Migrate(ctx context.Context, packId uint, request dto.MigratePackRequest, user tables.User) (dto.MigrateResponse, response.ServerError) {

	dbPack, err := ps.GetPackById(packId)
	if err != nil {
		return dto.MigrateResponse{}, err
	}

	target, tErr := resolveMigrationTarget(dbPack, request)
	if tErr != nil {
		return dto.MigrateResponse{}, tErr
	}

	unchanged := target.MCVersion == dbPack.MCVersion &&
		target.LoaderName == dbPack.Loader &&
		target.LoaderVersion == dbPack.LoaderVersion &&
		len(request.AcceptableVersions) == 0

	if unchanged {
		return dto.MigrateResponse{}, nil
	}

	acceptableVersions, avErr := target.Pack.GetAcceptableGameVersions()
	if avErr != nil {
		return dto.MigrateResponse{}, response.Wrap(avErr)
	}

	if err := ps.db.Model(&tables.Pack{ID: packId}).Select(
		"MCVersion", "Loader", "LoaderVersion", "AcceptableGameVersions", "UpdatedBy",
	).Updates(tables.Pack{
		MCVersion:              target.MCVersion,
		Loader:                 target.LoaderName,
		LoaderVersion:          target.LoaderVersion,
		AcceptableGameVersions: datatypes.JSONSlice[string](acceptableVersions),
		UpdatedBy:              user.ID,
	}).Error; err != nil {
		return dto.MigrateResponse{}, response.Wrap(err)
	}

	if !request.UpdateMods {
		return dto.MigrateResponse{}, nil
	}

	if ps.riverClient == nil {
		return dto.MigrateResponse{}, response.New(http.StatusInternalServerError, "background jobs are not available")
	}

	result, insertErr := ps.riverClient.Insert(ctx, jobs.MigrateModsArgs{
		PackID:             packId,
		MCVersion:          target.MCVersion,
		LoaderName:         target.LoaderName,
		LoaderVersion:      target.LoaderVersion,
		AcceptableVersions: request.AcceptableVersions,
		UserID:             user.ID,
	}, nil)
	if insertErr != nil {
		// the version bump above already committed; a failed enqueue just
		// means mods won't be auto-updated, not a reason to fail the request.
		log.Error("failed to enqueue migrate_mods job:", insertErr)
		return dto.MigrateResponse{ModsQueued: false}, nil
	}

	jobId := result.Job.ID
	return dto.MigrateResponse{ModsQueued: true, JobId: &jobId}, nil
}

// ResolveMigratedMods implements jobs.MigrateModsResolver. It re-checks a
// pack's mods (already migrated to a new MC/loader target by Migrate)
// against that target using core.CheckAllMods - resilient per-mod, unlike
// core.UpdateAllMods - applies updates for mods that pass, and records a
// per-mod result row for every mod so the migrate job status endpoint can
// report outcomes.
func (ps *PackwizService) ResolveMigratedMods(ctx context.Context, args jobs.MigrateModsArgs, jobId int64) error {
	dbPack, err := ps.GetPackById(args.PackID)
	if err != nil {
		return err
	}

	pack := dbPack.AsMeta()

	results, checkErr := core.CheckAllMods(nil, pack)
	if checkErr != nil {
		return checkErr
	}

	bySlug := make(map[string]tables.Mod, len(dbPack.Mods))
	for _, m := range dbPack.Mods {
		bySlug[m.Slug] = m
	}

	type updateBatch struct {
		results []core.UpdateCheckResult
	}
	batches := make(map[string]*updateBatch)
	for _, r := range results {
		if r.Err != nil || !r.UpdateAvailable || r.Mod.Pin {
			continue
		}
		b, ok := batches[r.Source]
		if !ok {
			b = &updateBatch{}
			batches[r.Source] = b
		}
		b.results = append(b.results, r)
	}

	// sourceErrors carries a batch-level failure back to individual mods in
	// that batch, since Updater.DoUpdate reports one error for the whole
	// batch rather than per mod.
	sourceErrors := make(map[string]error)
	for source, batch := range batches {
		updater, ok := core.GetUpdater(source)
		if !ok {
			sourceErrors[source] = fmt.Errorf("no updater registered for source: %s", source)
			continue
		}
		mods := make([]*core.Mod, len(batch.results))
		cachedState := make([]interface{}, len(batch.results))
		for i, r := range batch.results {
			mods[i] = r.Mod
			cachedState[i] = r.CachedState
		}
		if doErr := updater.DoUpdate(mods, cachedState); doErr != nil {
			sourceErrors[source] = doErr
		}
	}

	return ps.db.Transaction(func(tx *gorm.DB) error {
		for _, r := range results {
			dbMod, ok := bySlug[r.Mod.Slug]
			if !ok {
				continue
			}

			resultRow := tables.ModMigrationResult{
				JobId:  jobId,
				PackID: args.PackID,
				ModId:  dbMod.ID,
				Slug:   r.Mod.Slug,
				Name:   r.Mod.Name,
				Pinned: r.Mod.Pin,
			}

			switch {
			case r.Err != nil:
				resultRow.Incompatible = true
				resultRow.Error = r.Err.Error()
			case !r.UpdateAvailable || r.Mod.Pin:
				resultRow.UpdateAvailable = r.UpdateAvailable
				resultRow.UpdateString = r.UpdateString
			case sourceErrors[r.Source] != nil:
				resultRow.Incompatible = true
				resultRow.Error = sourceErrors[r.Source].Error()
			default:
				resultRow.UpdateAvailable = true
				resultRow.UpdateString = r.UpdateString
				if err := applyModUpdate(tx, dbMod.ID, r.Mod, tables.User{ID: args.UserID}); err != nil {
					return err
				}
			}

			if err := tx.Create(&resultRow).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetMigrateJobStatus reports a MigrateModsArgs job's lifecycle state, plus
// its per-mod results once it has completed.
func (ps *PackwizService) GetMigrateJobStatus(ctx context.Context, jobId int64) (dto.MigrateJobStatusResponse, response.ServerError) {
	if ps.riverClient == nil {
		return dto.MigrateJobStatusResponse{}, response.New(http.StatusInternalServerError, "background jobs are not available")
	}

	jobRow, err := ps.riverClient.JobGet(ctx, jobId)
	if err != nil {
		return dto.MigrateJobStatusResponse{}, response.Wrap(err)
	}

	out := dto.MigrateJobStatusResponse{State: string(jobRow.State)}

	if jobRow.State != rivertype.JobStateCompleted {
		return out, nil
	}

	var rows []tables.ModMigrationResult
	if err := ps.db.Where("job_id = ?", jobId).Find(&rows).Error; err != nil {
		return dto.MigrateJobStatusResponse{}, response.Wrap(err)
	}

	out.Mods = make([]dto.MigrateDryRunMod, 0, len(rows))
	for _, row := range rows {
		out.Mods = append(out.Mods, dto.MigrateDryRunMod{
			ModId:           row.ModId,
			Slug:            row.Slug,
			Name:            row.Name,
			Pinned:          row.Pinned,
			UpdateAvailable: row.UpdateAvailable,
			UpdateString:    row.UpdateString,
			Incompatible:    row.Incompatible,
			Error:           row.Error,
		})
	}

	return out, nil
}

// MigrateDryRun
// resolves request the same way Migrate does, then checks (but does not
// apply) each mod's compatibility with the candidate Minecraft version /
// loader target, so a caller can preview what Migrate would do before
// committing to it. Performs no persistence.
func (ps *PackwizService) MigrateDryRun(packId uint, request dto.MigratePackRequest) (dto.MigrateDryRunResponse, response.ServerError) {
	dbPack, err := ps.GetPackById(packId)
	if err != nil {
		return dto.MigrateDryRunResponse{}, err
	}

	target, tErr := resolveMigrationTarget(dbPack, request)
	if tErr != nil {
		return dto.MigrateDryRunResponse{}, tErr
	}

	results, checkErr := core.CheckAllMods(nil, target.Pack)
	if checkErr != nil {
		return dto.MigrateDryRunResponse{}, response.Wrap(checkErr)
	}

	bySlug := make(map[string]tables.Mod, len(dbPack.Mods))
	for _, m := range dbPack.Mods {
		bySlug[m.Slug] = m
	}

	out := dto.MigrateDryRunResponse{Mods: make([]dto.MigrateDryRunMod, 0, len(results))}
	for _, r := range results {
		dbMod := bySlug[r.Mod.Slug]
		item := dto.MigrateDryRunMod{
			ModId:  dbMod.ID,
			Slug:   r.Mod.Slug,
			Name:   r.Mod.Name,
			Pinned: r.Mod.Pin,
		}
		if r.Err != nil {
			item.Incompatible = true
			item.Error = r.Err.Error()
		} else {
			item.UpdateAvailable = r.UpdateAvailable
			item.UpdateString = r.UpdateString
		}
		out.Mods = append(out.Mods, item)
	}

	return out, nil
}
