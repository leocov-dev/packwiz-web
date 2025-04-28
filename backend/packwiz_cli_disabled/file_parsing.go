package packwiz_cli_disabled

import (
	"errors"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"os"
	"packwiz-web/internal/config"
	"packwiz-web/internal/pwlib/core"
	"packwiz-web/internal/utils"
	"path/filepath"
)

func parsePackFile(data []byte) (packFile core.Pack, err error) {
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

func parseIndexFile(data []byte) (indexFile core.Index, err error) {
	err = toml.Unmarshal(data, &indexFile)
	if err != nil {
		return indexFile, err
	}
	return indexFile, nil
}

func parseModFile(data []byte) (modFile core.Mod, err error) {
	err = toml.Unmarshal(data, &modFile)
	if err != nil {
		return modFile, err
	}
	return modFile, nil
}

func buildFilePath(modpack string, paths ...string) string {
	pathParts := []string{
		config.C.PackwizDir,
		modpack,
	}
	pathParts = append(pathParts, paths...)

	return filepath.Join(pathParts...)
}

func readFile(modpack string, paths ...string) ([]byte, error) {
	return os.ReadFile(buildFilePath(modpack, paths...))
}

func PackExists(modpack string) bool {
	return utils.FileExists(buildFilePath(modpack, "pack.toml"))
}

func GetPackFile(modpack string) (core.Pack, error) {
	packRaw, readErr := readFile(modpack, "pack.toml")
	if readErr != nil {
		return core.Pack{}, readErr
	}

	return parsePackFile(packRaw)
}

func GetIndexFile(modpack string) (core.Index, error) {
	return core.LoadIndex(buildFilePath(modpack, "index.toml"))
}

func GetModFile(modpack string, modFilePath string) (core.Mod, error) {
	return core.LoadMod(buildFilePath(modpack, modFilePath))
}

//func ModExists(modpack string, modName string) error {
//	indexFile, err := GetIndexFile(modpack)
//	if err != nil {
//		return err
//	}
//
//	for _, mod := range indexFile.Files {
//		parsedName := strings.TrimSuffix(filepath.Base(mod.), ".pw.toml")
//		if parsedName == modName {
//			return nil
//		}
//	}
//
//	return errors.New(fmt.Sprintf("pack: %s mod: %s not found", modpack, modName))
//}

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
