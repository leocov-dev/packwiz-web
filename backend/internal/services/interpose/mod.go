package interpose

import (
	"fmt"
	"packwiz-web/internal/packwiz_cli"
	"packwiz-web/internal/types"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/packwiz_schema"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ModInfo struct {
	Slug               string
	Name               string
	Type               string
	Filename           string
	Side               types.ModSide
	Pinned             bool
	DownloadUrl        string
	DownloadMode       string
	DownloadHash       string
	DownloadHashFormat string
	Source             string
	ModKey             string
	VersionKey         string
	Hash               string
	HashFormat         string
}

func AddModToPack(slug string, request dto.AddModRequest) (ModInfo, error) {
	var name string
	var err error

	if request.Modrinth.IsSet() {
		data := request.Modrinth
		name, err = packwiz_cli.AddModrinthMod(slug, data.Url, "", "", "")
		if err != nil {
			return ModInfo{}, err
		}
	} else if request.Curseforge.IsSet() {
		data := request.Curseforge
		name, err = packwiz_cli.AddCurseforgeMod(slug, data.Url, "", "", "", "")
		if err != nil {
			return ModInfo{}, err
		}
	}

	time.Sleep(3 * time.Second)

	return findModData(slug, name)
}

func getModData(modpack string, modFilePath string) (ModInfo, error) {
	indexFile, err := packwiz_cli.GetIndexFile(modpack)
	if err != nil {
		return ModInfo{}, err
	}

	modFile, err := packwiz_cli.GetModFile(modpack, modFilePath)
	if err != nil {
		return ModInfo{}, err
	}

	var hash string
	for _, mod := range indexFile.Files {
		if modFilePath == mod.File {
			hash = mod.Hash
		}
	}
	if hash == "" {
		return ModInfo{}, fmt.Errorf("mod not found in index file")
	}

	source, modKey, versionKey := getModSource(modFile)

	return ModInfo{
		Slug:               strings.TrimSuffix(filepath.Base(modFilePath), ".pw.toml"),
		Name:               modFile.Name,
		Type:               filepath.Dir(modFilePath),
		Filename:           modFile.Filename,
		Side:               modFile.Side,
		Pinned:             modFile.Pin,
		DownloadUrl:        modFile.Download.Url,
		DownloadMode:       modFile.Download.Mode,
		DownloadHash:       modFile.Download.Hash,
		DownloadHashFormat: modFile.Download.HashFormat,
		Source:             source,
		ModKey:             modKey,
		VersionKey:         versionKey,
		Hash:               hash,
		HashFormat:         indexFile.HashFormat,
	}, nil
}

func getModSource(modFile packwiz_schema.ModFile) (string, string, string) {
	if modFile.Update.Modrinth.ModId != "" {
		return "modrinth", modFile.Update.Modrinth.ModId, modFile.Update.Modrinth.Version

	} else {
		return "curseforge", strconv.Itoa(modFile.Update.Curseforge.FileId), strconv.Itoa(modFile.Update.Curseforge.ProjectId)

	}
}

func findModData(modpack string, modName string) (modData ModInfo, err error) {
	indexFile, err := packwiz_cli.GetIndexFile(modpack)
	if err != nil {
		return modData, err
	}

	for _, mod := range indexFile.Files {
		modFile, err := packwiz_cli.GetModFile(modpack, mod.File)
		if err != nil {
			return modData, err
		}
		if modFile.Name == modName {
			return getModData(modpack, mod.File)
		}
	}

	return modData, nil
}
