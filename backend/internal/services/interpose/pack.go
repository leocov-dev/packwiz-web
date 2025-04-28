package interpose

import (
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
)

type CreatePackInfo struct {
	MCVersion          string
	LoaderType         string
	LoaderVersion      string
	AcceptableVersions []string
	Version            string
	PackFormat         string
	Hash               string
	HashFormat         string
}

func CreatePack(
	request dto.NewPackRequest,
	author tables.User,
) (CreatePackInfo, error) {
	// file operations
	if err := packwiz_cli.NewModpack(
		request.Slug,
		request.Name,
		author.Username,
		request.Version,
		request.MinecraftDef.AsCliType(),
		request.LoaderDef.AsCliType(),
	); err != nil {
		return CreatePackInfo{}, err
	}

	if request.AcceptableVersions != nil && len(request.AcceptableVersions) > 0 {
		if err := packwiz_cli.SetAcceptableVersions(
			request.Slug,
			request.AcceptableVersions...,
		); err != nil {
			return CreatePackInfo{}, err
		}
	}

	packFile, err := packwiz_cli.GetPackFile(request.Slug)
	if err != nil {
		return CreatePackInfo{}, err
	}

	loaderType, loaderVersion := getLoaderData(packFile.Versions)

	return CreatePackInfo{
		MCVersion:          packFile.Versions["minecraft"],
		LoaderType:         loaderType,
		LoaderVersion:      loaderVersion,
		AcceptableVersions: packFile.Options["acceptable-game-versions"].([]string),
		Version:            packFile.Version,
		PackFormat:         packFile.PackFormat,
		Hash:               packFile.Index.Hash,
		HashFormat:         packFile.Index.HashFormat,
	}, nil
}

func getLoaderData(versions map[string]string) (string, string) {
	var loaderKey string
	var loaderValue string
	for k, v := range versions {
		if k != "minecraft" {
			loaderKey = k
			loaderValue = v
			break
		}
	}
	return loaderKey, loaderValue
}
