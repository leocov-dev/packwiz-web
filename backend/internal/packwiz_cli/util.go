package packwiz_cli

import (
	"errors"
	"fmt"
	"os"
	"packwiz-web/internal/config"
	"packwiz-web/internal/types"
	"packwiz-web/internal/utils"
	"path/filepath"
)

func DeleteModpack(modpack string) error {
	modpackDir := filepath.Join(config.C.PackwizDir, modpack)

	if !utils.DirectoryExists(modpackDir) {
		return errors.New(fmt.Sprintf("Modpack '%s' does not exist", modpack))
	}

	err := os.RemoveAll(modpackDir)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to delete modpack '%s'", modpack))
	}

	return nil
}

func LoaderDataFromVersionsData(versions PackFileVersions) *types.LoaderData {
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
