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

func loaderDataFromVersionsData(versions packwiz_cli.PackFileVersions) *types.LoaderData {
	var loaderType string
	var loaderVersion string

	if versions.Forge != "" {
		loaderType = "forge"
		loaderVersion = versions.Forge
	} else if versions.Fabric != "" {
		loaderType = "fabric"
		loaderVersion = versions.Fabric
	} else if versions.LiteLoader != "" {
		loaderType = "liteloader"
		loaderVersion = versions.LiteLoader
	} else if versions.Quilt != "" {
		loaderType = "quilt"
		loaderVersion = versions.Quilt
	} else if versions.NeoForge != "" {
		loaderType = "neoforge"
		loaderVersion = versions.NeoForge
	} else {
		return nil
	}

	return &types.LoaderData{
		Type:    loaderType,
		Version: loaderVersion,
	}
}

func getModpackData(modpack string) (*types.PackData, error) {
	packFile, err := packwiz_cli.GetPackFile(modpack)
	if err != nil {
		return nil, err
	}

	packData := &types.PackData{
		Name:       packFile.Name,
		Version:    packFile.Version,
		PackFormat: packFile.PackFormat,
		Versions: types.VersionsData{
			Minecraft: packFile.Versions.Minecraft,
			Loader:    loaderDataFromVersionsData(packFile.Versions),
		},
		Options: types.OptionsData{
			AcceptableGameVersions: packFile.Options.AcceptableGameVersions,
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
