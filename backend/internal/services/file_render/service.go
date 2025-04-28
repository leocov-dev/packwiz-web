package file_render

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"gorm.io/gorm"
	"io"
	"packwiz-web/internal/pwlib/core"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/packwiz_schema"
	"strconv"
)

type FileRenderService struct {
	db *gorm.DB
}

func NewFileRenderService(db *gorm.DB) *FileRenderService {
	return &FileRenderService{
		db: db,
	}
}

func (frs *FileRenderService) BuildPackFile(slug string) (core.Pack, error) {
	var pack tables.Pack
	err := frs.db.Where("slug = ?", slug).First(&pack).Error
	if err != nil {
		return core.Pack{}, err
	}

	var author tables.User
	if err := frs.db.Model(&author).First(&author, &tables.User{Id: pack.CreatedBy}).Error; err != nil {
		return core.Pack{}, err
	}

	return core.Pack{
		Name:       pack.Name,
		Author:     author.Username,
		Version:    pack.Version,
		PackFormat: pack.PackFormat,
		Index: struct {
			File       string `toml:"file"`
			HashFormat string `toml:"hash-format"`
			Hash       string `toml:"hash,omitempty"`
		}{
			File:       "index.toml",
			HashFormat: pack.HashFormat,
			Hash:       pack.Hash,
		},
		Versions: map[string]string{
			"minecraft": pack.MCVersion,
			pack.Loader: pack.LoaderVersion,
		},
		Options: map[string]interface{}{
			"acceptable-game-versions": pack.AcceptableGameVersions,
		},
	}, nil
}

func (frs *FileRenderService) BuildIndexFile(slug string) (core.Index, error) {

	var mods []tables.Mod

	if err := frs.db.Where(
		"pack_slug = ?", slug,
	).Find(
		&mods,
	).Order(
		"mod_slug",
	).Error; err != nil {
		return core.Index{}, err
	}

	var hashFormat string
	if len(mods) == 0 {
		hashFormat = "sha256"
	} else {
		hashFormat = mods[0].HashFormat
	}

	indexFile := core.Index{
		HashFormat: hashFormat,
		//Files:      make([]core.IndexFiles, 0),
	}

	//for _, mod := range mods {
	//	meta := packwiz_schema.IndexMeta{
	//		File:     fmt.Sprintf("%s/%s.pw.toml", mod.Type, mod.ModSlug),
	//		Hash:     mod.Hash,
	//		Metafile: mod.Metafile,
	//	}
	//	indexFile.Files = append(indexFile.Files, meta)
	//}

	return indexFile, nil
}

func (frs *FileRenderService) BuildModFile(slug, modSlug string) (packwiz_schema.ModFile, error) {
	var mod tables.Mod
	err := frs.db.Where("pack_slug = ? AND mod_slug = ?", slug, modSlug).First(&mod).Error
	if err != nil {
		return packwiz_schema.ModFile{}, err
	}

	var updateSourceMap packwiz_schema.UpdateSourceMap
	switch mod.Source {
	case "modrinth":
		updateSourceMap = packwiz_schema.UpdateSourceMap{
			Modrinth: packwiz_schema.ModrinthMeta{
				ModId:   mod.ModKey,
				Version: mod.VersionKey,
			},
		}
	case "curseforge":
		projectId, err := strconv.Atoi(mod.ModKey)
		if err != nil {
			return packwiz_schema.ModFile{}, err
		}
		fileId, err := strconv.Atoi(mod.VersionKey)
		if err != nil {
			return packwiz_schema.ModFile{}, err
		}
		updateSourceMap = packwiz_schema.UpdateSourceMap{
			Curseforge: packwiz_schema.CurseforgeMeta{
				ProjectId: projectId,
				FileId:    fileId,
			},
		}
	default:
		return packwiz_schema.ModFile{}, fmt.Errorf("unsupported mod source: %s", mod.Source)
	}

	return packwiz_schema.ModFile{
		Name:     mod.Name,
		Filename: mod.FileName,
		Side:     mod.Side,
		Pin:      mod.Pinned,
		Download: packwiz_schema.DownloadMeta{
			Url:        mod.DownloadUrl,
			Mode:       mod.DownloadMode,
			Hash:       mod.DownloadHash,
			HashFormat: mod.DownloadHashFormat,
		},
		Update: updateSourceMap,
	}, nil
}

func (frs *FileRenderService) UpdateHashes(slug string) error {
	if err := frs.updateAllModHashes(slug); err != nil {
		return err
	}

	if err := frs.updatePackHash(slug); err != nil {
		return err
	}

	return nil
}

func (frs *FileRenderService) updateAllModHashes(slug string) error {
	type ModData struct {
		ModSlug    string
		HashFormat string
		Hash       string
	}

	var modData []ModData
	if err := frs.db.Model(&tables.Mod{}).Where("pack_slug = ?", slug).Find(&modData).Error; err != nil {
		return err
	}

	for _, data := range modData {
		modFile, err := frs.BuildModFile(slug, data.ModSlug)
		if err != nil {
			return err
		}

		h, err := core.GetHashImpl(data.HashFormat)
		if err != nil {
			return err
		}

		f := bytes.NewBuffer(nil)

		w := io.MultiWriter(h, f)

		enc := toml.NewEncoder(w)
		// Disable indentation
		enc.Indent = ""

		if err = enc.Encode(modFile); err != nil {
			return err
		}

		modHash := h.HashToString(h.Sum(nil))

		if modHash != data.Hash {
			if err := frs.db.Model(
				&tables.Mod{},
			).Where(
				"pack_slug = ? AND mod_slug = ?", slug, data.ModSlug,
			).Update(
				"hash", modHash,
			).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func (frs *FileRenderService) updatePackHash(slug string) error {
	var pack tables.Pack
	if err := frs.db.Model(&pack).First(&pack, &tables.Pack{Slug: slug}).Error; err != nil {
		return err
	}

	indexFile, err := frs.BuildIndexFile(slug)
	if err != nil {
		return err
	}

	h, err := core.GetHashImpl(pack.HashFormat)
	if err != nil {
		return err
	}

	f := bytes.NewBuffer(nil)

	w := io.MultiWriter(h, f)

	enc := toml.NewEncoder(w)
	// Disable indentation
	enc.Indent = ""

	if err = enc.Encode(indexFile); err != nil {
		return err
	}

	indexHash := h.HashToString(h.Sum(nil))

	if err := frs.db.Model(
		&tables.Pack{},
	).Where(
		"slug = ?", slug,
	).Update(
		"hash", indexHash,
	).Error; err != nil {
		return err
	}

	return nil
}
