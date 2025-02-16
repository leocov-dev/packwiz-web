package packwiz_svc

import (
	"packwiz-web/internal/services/packwiz_cli"
	"packwiz-web/internal/types"
	"path/filepath"
	"strings"
)

func getModSource(modFile packwiz_cli.ModFile) types.ModSource {
	if modFile.Update.Modrinth.ModrinthId != "" {
		return types.ModSource{
			Type:       "modrinth",
			ModrinthId: modFile.Update.Modrinth.ModrinthId,
			Version:    modFile.Update.Modrinth.Version,
		}
	} else {
		return types.ModSource{
			Type:      "curseforge",
			FileId:    modFile.Update.Curseforge.FileId,
			ProjectId: modFile.Update.Curseforge.ProjectId,
		}
	}
}

func getModData(modpack string, modFilePath string) (types.ModData, error) {
	modFile, err := packwiz_cli.GetModFile(modpack, modFilePath)
	if err != nil {
		return types.ModData{}, err
	}

	return types.ModData{
		Name:        strings.TrimSuffix(filepath.Base(modFilePath), ".pw.toml"),
		DisplayName: modFile.Name,
		Type:        strings.TrimSuffix(strings.Split(modFilePath, "/")[0], "s"),
		Filename:    modFile.Filename,
		Side:        modFile.Side,
		Pinned:      modFile.Pin,
		Source:      getModSource(modFile),
	}, nil
}

func loaderDataFromVersionsData(versions packwiz_cli.PackFileVersions) types.LoaderData {
	var loader_type string
	var loader_version string

	if versions.Forge != "" {
		loader_type = "forge"
		loader_version = versions.Forge
	} else if versions.Fabric != "" {
		loader_type = "fabric"
		loader_version = versions.Fabric
	} else if versions.LiteLoader != "" {
		loader_type = "liteloader"
		loader_version = versions.LiteLoader
	} else if versions.Quilt != "" {
		loader_type = "quilt"
		loader_version = versions.Quilt
	} else if versions.NeoForge != "" {
		loader_type = "neoforge"
		loader_version = versions.NeoForge
	}

	return types.LoaderData{
		Type:    loader_type,
		Version: loader_version,
	}
}

func getModpackData(modpack string) (packData types.PackData, err error) {
	packFile, err := packwiz_cli.GetPackFile(modpack)
	if err != nil {
		return packData, err
	}

	packData = types.PackData{
		Name:       packFile.Name,
		Version:    packFile.Version,
		PackFormat: packFile.PackFormat,
		Versions: types.VersionsData{
			Minecraft: packFile.Versions.Minecraft,
			Loader:    loaderDataFromVersionsData(packFile.Versions),
		},
	}

	return packData, nil
}

func getModpackMods(modpack string) ([]types.ModData, error) {
	indexFile, err := packwiz_cli.GetIndexFile(modpack)
	if err != nil {
		return nil, err
	}

	modDataList := make([]types.ModData, len(indexFile.Files))

	for i, mod := range indexFile.Files {
		modData, err := getModData(modpack, mod.File)
		if err != nil {
			return nil, err
		}

		modDataList[i] = modData
	}

	return modDataList, nil
}

func findModData(modpack string, modName string) (modData types.ModData, err error) {
	indexFile, err := packwiz_cli.GetIndexFile(modpack)
	if err != nil {
		return modData, err
	}

	for _, mod := range indexFile.Files {
		parsedName := strings.TrimSuffix(filepath.Base(mod.File), ".pw.toml")
		if parsedName == modName {
			return getModData(modpack, mod.File)
		}
	}

	return modData, nil
}
