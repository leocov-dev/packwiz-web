package packwiz_cli

import (
	"errors"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"os"
	"packwiz-web/internal/config"
	"packwiz-web/internal/utils"
	"path/filepath"
	"strings"
)

func parsePackFile(data []byte) (packFile PackFile, err error) {
	err = toml.Unmarshal(data, &packFile)
	if err != nil {
		return packFile, err
	}

	// TODO: do we need more validation?
	if packFile.Name == "" {
		return packFile, errors.New("pack.toml is invalid")
	}
	return packFile, nil
}

func parseIndexFile(data []byte) (indexFile IndexFile, err error) {
	err = toml.Unmarshal(data, &indexFile)
	if err != nil {
		return indexFile, err
	}
	return indexFile, nil
}

func parseModFile(data []byte) (modFile ModFile, err error) {
	err = toml.Unmarshal(data, &modFile)
	if err != nil {
		return modFile, err
	}
	return modFile, nil
}

func readFile(modpack string, paths ...string) ([]byte, error) {
	pathParts := []string{
		config.C.PackwizDir,
		modpack,
	}
	pathParts = append(pathParts, paths...)

	expectedFile := filepath.Join(pathParts...)

	return os.ReadFile(expectedFile)
}

func PackExists(modpack string) bool {
	return utils.FileExists(filepath.Join(config.C.PackwizDir, modpack, "pack.toml"))
}

func GetPackFile(modpack string) (PackFile, error) {
	packRaw, readErr := readFile(modpack, "pack.toml")
	if readErr != nil {
		return PackFile{}, readErr
	}

	return parsePackFile(packRaw)
}

func GetIndexFile(modpack string) (IndexFile, error) {
	indexRaw, readErr := readFile(modpack, "index.toml")
	if readErr != nil {
		return IndexFile{}, readErr
	}

	return parseIndexFile(indexRaw)
}

func GetModFile(modpack string, modFilePath string) (ModFile, error) {
	modRaw, readErr := readFile(modpack, modFilePath)
	if readErr != nil {
		return ModFile{}, readErr
	}

	return parseModFile(modRaw)
}

func ModExists(modpack string, modName string) error {
	indexFile, err := GetIndexFile(modpack)
	if err != nil {
		return err
	}

	for _, mod := range indexFile.Files {
		parsedName := strings.TrimSuffix(filepath.Base(mod.File), ".pw.toml")
		if parsedName == modName {
			return nil
		}
	}

	return errors.New(fmt.Sprintf("pack: %s mod: %s not found", modpack, modName))
}

func GetAllModpackNames() (packNames []string, err error) {
	entries, err := os.ReadDir(config.C.PackwizDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", config.C.PackwizDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		packTomlPath := filepath.Join(config.C.PackwizDir, entry.Name(), "pack.toml")
		if _, err := os.Stat(packTomlPath); err == nil {
			packNames = append(packNames, entry.Name())
		}
	}

	return packNames, nil
}
