package packwiz_svc

import (
	"context"
	"database/sql"
	"errors"
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
	"packwiz-web/internal/services/authz_svc"
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
	authz       *authz_svc.Service
}

func NewPackwizService(db *gorm.DB, riverClient *river.Client[*sql.Tx]) *PackwizService {
	return &PackwizService{
		db:          db,
		riverClient: riverClient,
		authz:       authz_svc.NewService(db),
	}
}

func (ps *PackwizService) GetPacksWithPerms(
	request dto.AllPacksQuery,
	user tables.User,
	system authz_svc.Set,
) ([]dto.PackResponse, response.ServerError) {
	if len(request.Status) == 0 && !request.Archived {
		request.Status = []types.PackStatus{types.PackStatusDraft, types.PackStatusPublished}
	}

	var results []dto.PackResponse

	visibleIds, all, err := ps.authz.VisiblePackIDs(user, system)
	if err != nil {
		return nil, response.Wrap(err)
	}

	query := ps.db.Model(
		&tables.Pack{},
	).Select(
		"packs.*",
	).Preload(
		"User",
	).Order("packs.slug asc")

	if !all {
		// only packs where the user holds pack.view
		query = query.Where("packs.id IN ?", visibleIds)
	}

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

	if err := ps.attachPermissions(results, user, system); err != nil {
		return nil, err
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

func (ps *PackwizService) NewPack(request dto.NewPackRequest, author tables.User) (uint, response.ServerError) {

	if ps.PackExistsBySlug(request.Slug, true) {
		return 0, response.New(http.StatusBadRequest, "pack already exists")
	}

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

	if err := ps.db.Transaction(func(tx *gorm.DB) error {
		return CreatePackWithOwner(tx, newPack, author)
	}); err != nil {
		return 0, response.Wrap(err)
	}

	return newPack.ID, nil
}

// CreatePackWithOwner inserts pack and grants author the owner role on it.
func CreatePackWithOwner(tx *gorm.DB, pack *tables.Pack, author tables.User) error {
	var owner tables.Role
	if err := tx.Where("name = ? AND scope = ?", "owner", tables.RoleScopePack).First(&owner).Error; err != nil {
		return fmt.Errorf("find owner role: %w", err)
	}

	if err := tx.Create(pack).Error; err != nil {
		return err
	}

	return tx.Create(&tables.PackUsers{
		PackID: pack.ID,
		UserID: author.ID,
		RoleID: owner.ID,
	}).Error
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
	).Preload(
		// the owner's username is written to pack.toml as the author
		"Author",
	).Where(
		&tables.Pack{Slug: slug},
	)

	if err := query.Unscoped().First(&result).Error; err != nil {
		return result, response.New(http.StatusNotFound, fmt.Sprintf("pack '%s' not found", slug))
	}

	result.Mods = filterValidDependencyIds(result.Mods)

	return result, nil
}

func (ps *PackwizService) GetPackWithPerms(packId uint, user tables.User, system authz_svc.Set) (dto.PackResponse, response.ServerError) {
	var result dto.PackResponse

	query := ps.db.Model(
		&tables.Pack{},
	).Preload(
		"Mods",
	).Where(
		"packs.id = ?", packId,
	)

	if err := query.Unscoped().First(&result).Error; err != nil {
		return result, response.New(http.StatusNotFound, fmt.Sprintf("pack '%d' not found", packId))
	}

	result.Mods = filterValidDependencyIds(result.Mods)

	one := []dto.PackResponse{result}
	if err := ps.attachPermissions(one, user, system); err != nil {
		return result, err
	}

	return one[0], nil
}

// attachPermissions fills each pack's effective permissions and display role.
func (ps *PackwizService) attachPermissions(packs []dto.PackResponse, user tables.User, system authz_svc.Set) response.ServerError {
	archived := make(map[uint]bool, len(packs))
	for _, p := range packs {
		archived[p.ID] = p.DeletedAt.Valid
	}

	perms, roles, err := ps.authz.EffectivePackPermissions(user, system, archived)
	if err != nil {
		return response.Wrap(err)
	}

	for i := range packs {
		packs[i].Permissions = perms[packs[i].ID]
		packs[i].CurrentUserRole = roles[packs[i].ID]
	}

	return nil
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

	detail := map[string]any{"slug": newMod.Slug, "name": newMod.Name}
	if err := ps.withPackHistory(packId, user.ID, tables.SnapshotModAdd, detail, func(tx *gorm.DB) error {

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
		Version:  mod.Version,
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
//
// Publishing records a history snapshot when the content differs from the
// pack's head (the first publish, or edits made while it was a draft). Going
// back to draft records nothing and stops history until the next publish.
func (ps *PackwizService) SetPackStatus(packId uint, status types.PackStatus, user tables.User) response.ServerError {
	if err := ps.db.Transaction(func(tx *gorm.DB) error {
		if _, err := lockPackRow(tx, packId); err != nil {
			return err
		}
		if err := tx.Model(&tables.Pack{ID: packId}).Update("status", status).Error; err != nil {
			return err
		}
		if status == types.PackStatusPublished {
			return recordPackHistory(tx, packId, user.ID, tables.SnapshotPublish, nil)
		}
		return nil
	}); err != nil {
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

// UpdateAll
// update all the mods in a pack with partial success: pinned mods are skipped,
// per-mod failures are collected and do not abort the run. Mods are checked in
// one pass (CheckAllMods, per-mod errors) and each available update is applied
// straight from that check result (no second network check). Updates are
// sequential (the nxt updaters are not verified concurrency-safe). Only mods
// that actually changed are written to the DB, so unchanged rows keep their
// updated_at.
//
// The work runs detached from ctx's cancellation: a client disconnect or
// timeout must not leave a half-applied update answered with an error, so the
// run always completes and its result is simply dropped if nobody listens.
// Only one update per pack may run at a time; a concurrent call gets 409.
func (ps *PackwizService) UpdateAll(ctx context.Context, packId uint, user tables.User) (dto.UpdateAllResponse, response.ServerError) {
	unlock, lockErr := lockPackUpdate(packId)
	if lockErr != nil {
		return dto.UpdateAllResponse{}, lockErr
	}
	defer unlock()

	ctx = context.WithoutCancel(ctx)
	db := ps.db.WithContext(ctx)

	dbPack, err := ps.GetPackById(packId)
	if err != nil {
		return dto.UpdateAllResponse{}, err
	}

	// the per-mod writes below are not one transaction, so capture the state
	// they start from first; the run's own snapshot is recorded after the loop
	if err := db.Transaction(func(tx *gorm.DB) error {
		return ensureHistoryBaseline(tx, packId, user.ID)
	}); err != nil {
		return dto.UpdateAllResponse{}, response.Wrap(err)
	}

	pack := dbPack.AsMeta()

	dbModsBySlug := make(map[string]tables.Mod, len(dbPack.Mods))
	for _, dbMod := range dbPack.Mods {
		dbModsBySlug[dbMod.Slug] = dbMod
	}

	checks, checkErr := core.CheckAllMods(nil, pack)
	if checkErr != nil {
		return dto.UpdateAllResponse{}, response.Wrap(checkErr)
	}

	summary := &updateAllSummary{}
	checked := make(map[string]bool, len(checks))

	for _, check := range checks {
		dbMod, ok := dbModsBySlug[check.Mod.Slug]
		if !ok {
			continue
		}
		checked[check.Mod.Slug] = true
		item := dto.UpdateAllItem{ModId: dbMod.ID, Slug: dbMod.Slug, Name: dbMod.Name}

		switch {
		case check.Err != nil && dbMod.Pinned:
			// pinned mods are never updated; the failed check is not actionable,
			// but we did not verify the mod so it is not "up to date" either
			log.Debug(fmt.Errorf("update check failed for pinned mod %s: %w", dbMod.Slug, check.Err))
			summary.addNotChecked()
		case check.Err != nil:
			summary.addFailed(item, check.Err)
		case !check.UpdateAvailable:
			summary.addUpToDate()
		case dbMod.Pinned:
			summary.addSkipped(item, SkipReasonPinned)
		default:
			applyCheckedUpdate(db, check, item, user, summary)
		}
	}

	// mods without a registered updater (e.g. manual URL sources) are never checked
	for slug := range dbModsBySlug {
		if !checked[slug] {
			summary.addNotChecked()
		}
	}

	// stored check results no longer describe the pack
	invalidatePackChecks(db, packId)

	if len(summary.updated) > 0 {
		// The updates above are already committed (partial success is allowed),
		// so a failure here must not turn the run into an error. Snapshots hold
		// the full state, so the next recorded change captures these updates too.
		detail := map[string]any{"updated": len(summary.updated), "failed": len(summary.failed)}
		if err := db.Transaction(func(tx *gorm.DB) error {
			return recordPackHistory(tx, packId, user.ID, tables.SnapshotUpdateAll, detail)
		}); err != nil {
			log.Error(fmt.Sprintf("failed to record history for update-all of pack %d:", packId), err)
		}
	}

	return summary.response(), nil
}

// applyCheckedUpdate applies the update found by CheckAllMods for a single mod
// using the check's cached state, recording the outcome in summary. Failures
// are recorded, not returned.
func applyCheckedUpdate(db *gorm.DB, check core.UpdateCheckResult, item dto.UpdateAllItem, user tables.User, summary *updateAllSummary) {
	updater, ok := core.GetUpdater(check.Source)
	if !ok {
		summary.addFailed(item, fmt.Errorf("no updater registered for source: %s", check.Source))
		return
	}

	mod := check.Mod
	before, snapErr := takeModSnapshot(mod)
	if snapErr != nil {
		summary.addFailed(item, snapErr)
		return
	}

	if updateErr := updater.DoUpdate([]*core.Mod{mod}, []any{check.CachedState}); updateErr != nil {
		summary.addFailed(item, updateErr)
		return
	}

	// DoUpdate may resolve to the identical file; don't write (and bump updated_at) then
	changed, cmpErr := before.changedSince(mod)
	if cmpErr != nil {
		summary.addFailed(item, cmpErr)
		return
	}
	if !changed {
		persistModVersion(db, item.ModId, mod.Version)
		summary.addUpToDate()
		return
	}

	if dbErr := applyModUpdate(db, item.ModId, mod, user); dbErr != nil {
		summary.addFailed(item, fmt.Errorf("failed to save update: %w", dbErr))
		return
	}

	item.FileName = mod.FileName
	summary.addUpdated(item)
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

	if txErr := ps.withPackHistory(packId, user.ID, tables.SnapshotRehash, map[string]any{"format": format}, func(tx *gorm.DB) error {
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

	var existing tables.Mod
	if err := tx.Select("id", "file_name", "download").
		First(&existing, dbModID).Error; err != nil {
		return err
	}
	fileChanged := existing.FileName != updated.FileName ||
		existing.Download.Hash != updated.Download.Hash

	nextVersion, writeVersion := nextStoredVersion(existing.Version, updated.Version, fileChanged)

	columns := []string{"FileName", "Download", "Source", "Update", "UpdatedBy"}
	if writeVersion && nextVersion != "" {
		columns = append(columns, "Version")
	}
	clearVersion := writeVersion && nextVersion == ""

	if err := tx.Model(&tables.Mod{ID: dbModID}).Select(columns).Updates(tables.Mod{
		FileName: updated.FileName,
		Version:  nextVersion,
		Download: tables.DownloadInfo{
			URL:        updated.Download.URL,
			Mode:       updated.Download.Mode,
			Hash:       updated.Download.Hash,
			HashFormat: updated.Download.HashFormat,
		},
		Source:    source,
		Update:    update,
		UpdatedBy: user.ID,
	}).Error; err != nil {
		return err
	}

	if clearVersion {
		// stale version must not outlive the file it described; NULL matches legacy rows
		return tx.Model(&tables.Mod{}).Where("id = ?", dbModID).
			UpdateColumn("version", nil).Error
	}
	return nil
}

// nextStoredVersion decides what to persist for mods.version after an update.
// A non-empty new version always wins. With no new version: keep the stored one
// if the file is unchanged, otherwise clear it (write=true, "") so the UI falls
// back to the file name instead of showing a stale version.
func nextStoredVersion(oldVersion, newVersion string, fileChanged bool) (version string, write bool) {
	if newVersion != "" {
		return newVersion, true
	}
	if fileChanged && oldVersion != "" {
		return "", true
	}
	return oldVersion, false
}

// persistModVersion stores a freshly resolved version without touching
// updated_at, for updates that resolved to the identical file (so the mod is
// not reported as changed). No-op for an empty version. Best effort: the
// version is display-only, so failure is logged, not returned.
func persistModVersion(db *gorm.DB, dbModID uint, version string) {
	if version == "" {
		return
	}
	if err := db.Model(&tables.Mod{}).Where("id = ?", dbModID).
		UpdateColumn("version", version).Error; err != nil {
		log.Error("failed to store mod version:", err)
	}
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
func (ps *PackwizService) RemoveModById(packId, modId uint, user tables.User) response.ServerError {
	detail := map[string]any{}

	if err := ps.withPackHistory(packId, user.ID, tables.SnapshotModRemove, detail, func(tx *gorm.DB) error {
		var mod tables.Mod
		if err := tx.Select("id", "slug", "name").
			Where("id = ? AND pack_id = ?", modId, packId).
			First(&mod).Error; err != nil {
			return err
		}
		detail["slug"] = mod.Slug
		detail["name"] = mod.Name

		return tx.Where("id = ? AND pack_id = ?", modId, packId).Delete(&tables.Mod{}).Error
	}); err != nil {
		return modWriteError(err, packId, modId)
	}

	return nil
}

// modWriteError maps a failed single-mod write to a response: a mod that is not
// in the pack is a 404, anything else is an internal error.
func modWriteError(err error, packId, modId uint) response.ServerError {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return response.New(http.StatusNotFound, fmt.Sprintf("pack %d with mod %d not found", packId, modId))
	}
	return response.Wrap(err)
}

// ModExistsInPack reports whether the mod belongs to the pack.
func (ps *PackwizService) ModExistsInPack(packId, modId uint) bool {
	var count int64
	if err := ps.db.Model(&tables.Mod{}).
		Where("id = ? AND pack_id = ?", modId, packId).
		Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}

// UpdateMod
// update a given mod from a given pack. Reports whether anything changed; the
// DB row is only written when it did.
func (ps *PackwizService) UpdateMod(modId uint, user tables.User) (dto.UpdateModResponse, response.ServerError) {
	modInfo, err := ps.GetMod(modId)
	if err != nil {
		return dto.UpdateModResponse{}, err
	}

	if modInfo.Pinned {
		return dto.UpdateModResponse{}, response.New(http.StatusBadRequest, "cannot update pinned mod")
	}

	unlock, lockErr := lockPackUpdate(modInfo.PackID)
	if lockErr != nil {
		return dto.UpdateModResponse{}, lockErr
	}
	defer unlock()

	dbPack, err := ps.GetPackById(modInfo.PackID)
	if err != nil {
		return dto.UpdateModResponse{}, err
	}

	pack := dbPack.AsMeta()

	mod, ok := pack.Mods[modInfo.Slug]
	if !ok {
		return dto.UpdateModResponse{}, response.New(http.StatusNotFound, fmt.Sprintf("mod '%s' not found in pack", modInfo.Slug))
	}

	before, snapErr := takeModSnapshot(mod)
	if snapErr != nil {
		return dto.UpdateModResponse{}, response.Wrap(snapErr)
	}

	if updateErr := core.UpdateSingleMod(nil, pack, mod); updateErr != nil {
		return dto.UpdateModResponse{}, response.Wrap(updateErr)
	}

	changed, cmpErr := before.changedSince(mod)
	if cmpErr != nil {
		return dto.UpdateModResponse{}, response.Wrap(cmpErr)
	}
	if !changed {
		persistModVersion(ps.db, modInfo.ID, mod.Version)
		invalidateModChecks(ps.db, modInfo.ID)
		return dto.UpdateModResponse{Updated: false}, nil
	}

	detail := map[string]any{"slug": modInfo.Slug, "name": modInfo.Name, "fileName": mod.FileName}
	if txErr := ps.withPackHistory(modInfo.PackID, user.ID, tables.SnapshotModUpdate, detail, func(tx *gorm.DB) error {
		return applyModUpdate(tx, modInfo.ID, mod, user)
	}); txErr != nil {
		return dto.UpdateModResponse{}, response.Wrap(txErr)
	}
	invalidateModChecks(ps.db, modInfo.ID)

	return dto.UpdateModResponse{Updated: true}, nil
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

// updateModConfig writes a single-mod configuration change (side, option, pin)
// and records the history snapshot for it. columns uses snake_case column
// names; updated_by is added here.
func (ps *PackwizService) updateModConfig(
	packId, modId uint,
	user tables.User,
	reason tables.SnapshotReason,
	detail map[string]any,
	columns map[string]any,
) response.ServerError {
	if err := ps.withPackHistory(packId, user.ID, reason, detail, func(tx *gorm.DB) error {
		var mod tables.Mod
		if err := tx.Select("id", "slug", "name").
			Where("id = ? AND pack_id = ?", modId, packId).
			First(&mod).Error; err != nil {
			return err
		}
		detail["slug"] = mod.Slug
		detail["name"] = mod.Name

		columns["updated_by"] = user.ID
		return tx.Model(&tables.Mod{}).
			Where("id = ? AND pack_id = ?", modId, packId).
			Updates(columns).Error
	}); err != nil {
		return modWriteError(err, packId, modId)
	}

	return nil
}

func (ps *PackwizService) ChangeModSide(packId, modId uint, side core.ModSide, user tables.User) response.ServerError {
	return ps.updateModConfig(packId, modId, user, tables.SnapshotModSide,
		map[string]any{"side": string(side)},
		map[string]any{"side": side},
	)
}

func (ps *PackwizService) ChangeModOption(packId, modId uint, req dto.ChangeModOptionRequest, user tables.User) response.ServerError {
	option := tables.OptionInfo{
		Optional:    req.Optional,
		Description: req.Description,
		Default:     req.Default,
	}

	return ps.updateModConfig(packId, modId, user, tables.SnapshotModOption,
		map[string]any{"optional": req.Optional, "default": req.Default},
		map[string]any{"option": option},
	)
}

func (ps *PackwizService) SetModPinnedValue(packId, modId uint, value bool, user tables.User) response.ServerError {
	return ps.updateModConfig(packId, modId, user, tables.SnapshotModPin,
		map[string]any{"pinned": value},
		map[string]any{"pinned": value},
	)
}

func (ps *PackwizService) GetPersonalLink(
	user tables.User,
	packId uint,
	scheme string,
	host string,
) (url.URL, response.ServerError) {

	var key string
	if ps.IsPackPublicById(packId) {
		key = publicLinkKey
	} else {
		key = user.LinkToken
	}

	pack, err := ps.GetPackById(packId)
	if err != nil {
		return url.URL{}, err
	}

	link, parseErr := url.Parse(packTomlURL(scheme, host, key, pack.Slug))
	if parseErr != nil {
		return url.URL{}, response.New(http.StatusInternalServerError, "failed to build link url")
	}

	return *link, nil
}

// PackUserInfo
// a user's access information for a given pack
type PackUserInfo struct {
	UserID    uint      `json:"userId"`
	Username  string    `json:"username"`
	FullName  string    `json:"fullName"`
	Email     string    `json:"email"`
	RoleID    uint      `json:"roleId"`
	RoleName  string    `json:"roleName"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

// ListPackUsers
// list all users with access to a pack
func (ps *PackwizService) ListPackUsers(packId uint) ([]PackUserInfo, response.ServerError) {
	var results []PackUserInfo

	if err := ps.db.Model(&tables.PackUsers{}).
		Select(
			"pack_users.user_id, users.username, users.full_name, users.email, pack_users.role_id, roles.name as role_name, users.is_active, pack_users.created_at",
		).
		Joins("JOIN users ON users.id = pack_users.user_id").
		Joins("JOIN roles ON roles.id = pack_users.role_id").
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
// search for active users by username/full name/email who do not already have access to a pack
func (ps *PackwizService) SearchPackUsers(packId uint, query string) ([]PackUserSearchResult, response.ServerError) {
	var results []PackUserSearchResult

	like := "%" + query + "%"
	if err := ps.db.Model(&tables.User{}).
		Select("users.id as user_id, users.username, users.full_name, users.email").
		Where("users.username ILIKE ? OR users.full_name ILIKE ? OR users.email ILIKE ?", like, like, like).
		Where("users.is_active = ?", true).
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
func (ps *PackwizService) GrantPackUser(packId, userId, roleId uint) response.ServerError {
	if err := ps.checkAssignablePackRole(roleId); err != nil {
		return err
	}

	if _, err := ps.GetPackById(packId); err != nil {
		return err
	}

	var target tables.User
	if err := ps.db.Select("id", "is_active").
		Where("id = ?", userId).
		Limit(1).
		Find(&target).Error; err != nil {
		return response.New(http.StatusInternalServerError, "failed to query db for user")
	}
	if target.ID == 0 {
		return response.New(http.StatusNotFound, fmt.Sprintf("user '%d' not found", userId))
	}
	if !target.IsActive {
		return response.New(http.StatusConflict, "user is deactivated and cannot be added to a pack")
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
			PackID: packId,
			UserID: userId,
			RoleID: roleId,
		}).Error
	}); err != nil {
		return response.Wrap(err)
	}

	return nil
}

// PackRoleInfo
// a pack-scope role that can be granted to collaborators
type PackRoleInfo struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// ListAssignablePackRoles
// list the pack roles that can be granted through the collaborator API
func (ps *PackwizService) ListAssignablePackRoles() ([]PackRoleInfo, response.ServerError) {
	var roles []tables.Role
	if err := ps.db.Preload("Permissions").
		Where("scope = ? AND assignable", tables.RoleScopePack).
		Order("id asc").
		Find(&roles).Error; err != nil {
		return nil, response.New(http.StatusInternalServerError, "failed to query db for roles")
	}

	out := make([]PackRoleInfo, 0, len(roles))
	for _, r := range roles {
		names := make([]string, 0, len(r.Permissions))
		for _, p := range r.Permissions {
			names = append(names, p.Name)
		}
		out = append(out, PackRoleInfo{ID: r.ID, Name: r.Name, Description: r.Description, Permissions: names})
	}

	return out, nil
}

// checkAssignablePackRole
// 400 unless the role exists, is pack-scope and may be granted manually
func (ps *PackwizService) checkAssignablePackRole(roleId uint) response.ServerError {
	var role tables.Role
	if err := ps.db.Where("id = ?", roleId).First(&role).Error; err != nil {
		return response.New(http.StatusBadRequest, fmt.Sprintf("role '%d' not found", roleId))
	}
	if role.Scope != tables.RoleScopePack {
		return response.New(http.StatusBadRequest, "not a pack role")
	}
	if !role.Assignable {
		return response.New(http.StatusBadRequest, fmt.Sprintf("%s role cannot be assigned", role.Name))
	}
	return nil
}

// abortIfPackOwner
// returns a forbidden error with msg if the user created the pack. Owner
// identity is packs.created_by, not a role check.
func (ps *PackwizService) abortIfPackOwner(packId, userId uint, msg string) response.ServerError {
	var isOwner bool
	if err := ps.db.Unscoped().Model(&tables.Pack{}).
		Select("1").
		Where("id = ? AND created_by = ?", packId, userId).
		Limit(1).
		Find(&isOwner).Error; err != nil {
		return response.New(http.StatusInternalServerError, "failed to query db for pack user")
	}
	if isOwner {
		return response.New(http.StatusForbidden, msg)
	}

	return nil
}

// RevokePackUser
// revoke a user's access to a pack
func (ps *PackwizService) RevokePackUser(packId, userId uint) response.ServerError {
	if err := ps.abortIfPackOwner(packId, userId, "the pack owner cannot be removed"); err != nil {
		return err
	}

	result := ps.db.Where("pack_id = ? AND user_id = ?", packId, userId).Delete(&tables.PackUsers{})
	if result.Error != nil {
		return response.Wrap(result.Error)
	}
	if result.RowsAffected == 0 {
		return response.New(http.StatusNotFound, "user does not have access to this pack")
	}

	return nil
}

// ChangePackUserRole
// change a user's role on a pack
func (ps *PackwizService) ChangePackUserRole(packId, userId, roleId uint) response.ServerError {
	if err := ps.checkAssignablePackRole(roleId); err != nil {
		return err
	}
	if err := ps.abortIfPackOwner(packId, userId, "the pack owner's role cannot be changed"); err != nil {
		return err
	}

	result := ps.db.Model(&tables.PackUsers{}).
		Where("pack_id = ? AND user_id = ?", packId, userId).
		Update("role_id", roleId)
	if result.Error != nil {
		return response.Wrap(result.Error)
	}
	if result.RowsAffected == 0 {
		return response.New(http.StatusNotFound, "user does not have access to this pack")
	}

	return nil
}

func (ps *PackwizService) EditPack(packId uint, request dto.EditPackRequest, user tables.User) response.ServerError {

	pack, err := ps.GetPackById(packId)
	if err != nil {
		return response.Wrap(err)
	}

	before := struct {
		mc, loader, loaderVersion string
		acceptable                []string
	}{pack.MCVersion, pack.Loader, pack.LoaderVersion, append([]string(nil), pack.AcceptableGameVersions...)}

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

	targetChanged := pack.MCVersion != before.mc ||
		pack.Loader != before.loader ||
		pack.LoaderVersion != before.loaderVersion ||
		!slices.Equal([]string(pack.AcceptableGameVersions), before.acceptable)

	// Write only the editable columns. A Save(pack) would also upsert the
	// preloaded Mods (re-inserting a mod removed meanwhile, and blanking the
	// DB-only mods.version) and write back a stale status/is_public/deleted_at.
	detail := map[string]any{}
	if err := ps.withPackHistory(packId, user.ID, tables.SnapshotPackEdit, detail, func(tx *gorm.DB) error {
		return tx.Model(&tables.Pack{}).
			Where("id = ?", packId).
			Select(
				"Name", "Description", "Version", "MCVersion", "Loader", "LoaderVersion",
				"AcceptableGameVersions", "UpdatedBy",
			).
			Updates(&tables.Pack{
				Name:                   pack.Name,
				Description:            pack.Description,
				Version:                pack.Version,
				MCVersion:              pack.MCVersion,
				Loader:                 pack.Loader,
				LoaderVersion:          pack.LoaderVersion,
				AcceptableGameVersions: pack.AcceptableGameVersions,
				UpdatedBy:              user.ID,
			}).Error
	}); err != nil {
		return response.Wrap(err)
	}

	if targetChanged {
		// update availability depends on the MC version/loader/acceptable versions
		invalidatePackChecks(ps.db, packId)
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

	if active, activeErr := ps.migrateJobActive(packId); activeErr != nil {
		return dto.MigrateResponse{}, response.Wrap(activeErr)
	} else if active {
		return dto.MigrateResponse{}, response.New(http.StatusConflict, "a migration is already running for this pack")
	}

	detail := map[string]any{
		"mcVersion":     target.MCVersion,
		"loader":        target.LoaderName,
		"loaderVersion": target.LoaderVersion,
	}
	if err := ps.withPackHistory(packId, user.ID, tables.SnapshotMigrate, detail, func(tx *gorm.DB) error {
		return tx.Model(&tables.Pack{ID: packId}).Select(
			"MCVersion", "Loader", "LoaderVersion", "AcceptableGameVersions", "UpdatedBy",
		).Updates(tables.Pack{
			MCVersion:              target.MCVersion,
			Loader:                 target.LoaderName,
			LoaderVersion:          target.LoaderVersion,
			AcceptableGameVersions: datatypes.JSONSlice[string](acceptableVersions),
			UpdatedBy:              user.ID,
		}).Error
	}); err != nil {
		return dto.MigrateResponse{}, response.Wrap(err)
	}

	// stored update checks were made against the old target
	invalidatePackChecks(ps.db, packId)

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

	updated := 0
	txErr := ps.db.Transaction(func(tx *gorm.DB) error {
		// pack row first, like every content transaction, so this cannot
		// deadlock with a revert or an edit
		if _, err := lockPackRow(tx, args.PackID); err != nil {
			return err
		}

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
				updated++
			}

			if err := tx.Create(&resultRow).Error; err != nil {
				return err
			}
		}

		// skipped for drafts and archived packs, and when nothing changed
		return recordPackHistory(tx, args.PackID, args.UserID, tables.SnapshotMigrateMods,
			map[string]any{"jobId": jobId, "updated": updated})
	})
	if txErr == nil {
		// mods were updated; stored update checks no longer describe the pack
		invalidatePackChecks(ps.db, args.PackID)
	}
	return txErr
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
