package import_svc

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"time"

	"github.com/leocov-dev/packwiz-nxt/core"
	"github.com/leocov-dev/packwiz-nxt/fileio"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"packwiz-web/internal/services/packwiz_svc"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

// knownLoaders are the pack.toml [versions] keys that name a mod loader.
var knownLoaders = []string{"fabric", "forge", "liteloader", "quilt", "neoforge"}

// RemoteLoader reads a live packwiz pack from its pack.toml url.
type RemoteLoader func(ctx context.Context, packUrl string) (*fileio.RemotePack, error)

type ImportService struct {
	db   *gorm.DB
	load RemoteLoader
	now  func() time.Time
}

func NewImportService(db *gorm.DB) *ImportService {
	return &ImportService{
		db: db,
		load: func(ctx context.Context, packUrl string) (*fileio.RemotePack, error) {
			return fileio.LoadRemotePack(ctx, packUrl, fileio.RemoteLoadOptions{})
		},
		now: time.Now,
	}
}

// ImportPack reads the pack at request.Url and stores it as a new draft pack
// owned by author. Slug and name clashes with existing packs are resolved by
// appending text; mods without a known update source and non-mod files are
// reported but not imported.
func (s *ImportService) ImportPack(ctx context.Context, request dto.ImportPackRequest, author tables.User) (dto.ImportPackResponse, response.ServerError) {
	remote, err := s.load(ctx, request.Url)
	if err != nil {
		return dto.ImportPackResponse{}, response.New(http.StatusBadRequest, fmt.Sprintf("import failed: %s", err))
	}

	pack, mods, skippedMods, sErr := s.buildRows(remote.Pack, request.Url, author)
	if sErr != nil {
		return dto.ImportPackResponse{}, sErr
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		pack.Slug = uniqueSlug(pack.Slug, func(v string) bool { return rowExists(tx, "slug", v) })
		pack.Name = uniqueName(pack.Name, func(v string) bool { return rowExists(tx, "name", v) })

		if err := packwiz_svc.CreatePackWithOwner(tx, &pack, author); err != nil {
			return err
		}
		for i := range mods {
			mods[i].PackID = pack.ID
		}
		if len(mods) == 0 {
			return nil
		}
		return tx.Create(&mods).Error
	}); err != nil {
		return dto.ImportPackResponse{}, response.Wrap(fmt.Errorf("store imported pack: %w", err))
	}

	return dto.ImportPackResponse{
		PackId:       pack.ID,
		Slug:         pack.Slug,
		Name:         pack.Name,
		ModsImported: len(mods),
		SkippedMods:  nonNil(skippedMods),
		SkippedFiles: nonNil(remote.SkippedFiles),
		Warnings:     nonNil(remote.Warnings),
	}, nil
}

// rowExists reports whether any pack, deleted or not, has column = value.
func rowExists(tx *gorm.DB, column, value string) bool {
	var count int64
	if err := tx.Unscoped().Model(&tables.Pack{}).Where(column+" = ?", value).Count(&count).Error; err != nil {
		// treat a failed lookup as a clash; the insert would fail on it anyway
		return true
	}
	return count > 0
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// buildRows maps a loaded pack onto unsaved table rows. Slug and name are the
// pack's own, not yet made unique.
func (s *ImportService) buildRows(src *core.Pack, packUrl string, author tables.User) (tables.Pack, []tables.Mod, []string, response.ServerError) {
	mcVersion := src.Versions["minecraft"]
	if mcVersion == "" {
		return tables.Pack{}, nil, nil, response.New(http.StatusBadRequest, "pack has no minecraft version")
	}
	loader, loaderVersion := pickLoader(src.Versions)
	if loader == "" {
		return tables.Pack{}, nil, nil, response.New(http.StatusBadRequest, "pack has no supported mod loader")
	}

	name := src.Name
	if name == "" {
		name = slugFromName("", packUrl)
	}
	version := src.Version
	if version == "" {
		version = "1.0.0"
	}

	pack := tables.Pack{
		Slug:                   slugFromName(src.Name, packUrl),
		Name:                   name,
		Description:            importedDescription(packUrl, s.now(), src.Description),
		CreatedBy:              author.ID,
		UpdatedBy:              author.ID,
		Status:                 types.PackStatusDraft,
		MCVersion:              mcVersion,
		Loader:                 loader,
		LoaderVersion:          loaderVersion,
		AcceptableGameVersions: acceptableVersions(src.Options),
		Version:                version,
		PackFormat:             core.CurrentPackFormat,
	}

	slugs := make([]string, 0, len(src.Mods))
	for slug := range src.Mods {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	var rows []tables.Mod
	var skipped []string
	for _, slug := range slugs {
		mod := src.Mods[slug]
		source, update := tables.ExtractModSource(mod)
		if source == "" {
			skipped = append(skipped, fmt.Sprintf("%s/%s: no modrinth, curseforge or github update source", mod.ModType, slug))
			continue
		}
		rows = append(rows, modRow(slug, mod, source, update, author))
	}
	return pack, rows, skipped, nil
}

func modRow(slug string, mod *core.Mod, source string, update map[string]interface{}, author tables.User) tables.Mod {
	row := tables.Mod{
		Slug:     slug,
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
		HashFormat: mod.HashFormat,
		Alias:      mod.Alias,
		Preserve:   mod.Preserve,
		Source:     source,
		Update:     update,
		CreatedBy:  author.ID,
		UpdatedBy:  author.ID,
	}
	if row.Type == "" {
		row.Type = "mods"
	}
	if row.HashFormat == "" {
		row.HashFormat = core.DefaultHashFormat
	}
	if mod.Option != nil {
		row.Option = tables.OptionInfo{
			Optional:    mod.Option.Optional,
			Description: mod.Option.Description,
			Default:     mod.Option.Default,
		}
	}
	return row
}

func pickLoader(versions map[string]string) (name, version string) {
	for _, l := range knownLoaders {
		if v, ok := versions[l]; ok {
			return l, v
		}
	}
	return "", ""
}

func acceptableVersions(options map[string]interface{}) datatypes.JSONSlice[string] {
	out := datatypes.JSONSlice[string]{}
	switch v := options["acceptable-game-versions"].(type) {
	case []string:
		out = slices.Clone(v)
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				out = append(out, str)
			}
		}
	}
	return out
}
