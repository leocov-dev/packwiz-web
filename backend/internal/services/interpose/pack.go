package interpose

import (
	"packwiz-web/internal/packwiz_cli"
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

	loader := packwiz_cli.LoaderDataFromVersionsData(packFile.Versions)

	return CreatePackInfo{
		MCVersion:          packFile.Versions.Minecraft,
		LoaderType:         loader.Type,
		LoaderVersion:      loader.Version,
		AcceptableVersions: packFile.Options.AcceptableGameVersions,
		Version:            packFile.Version,
		PackFormat:         packFile.PackFormat,
		Hash:               packFile.Index.Hash,
		HashFormat:         packFile.Index.HashFormat,
	}, nil
}
