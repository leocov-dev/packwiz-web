package packwiz_cli

import (
	"errors"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"os"
	"packwiz-web/internal/config"
	"packwiz-web/internal/types"
	"path/filepath"
	"strings"
)

func writeFile(modpack string, paths []string, data []byte) error {
	pathParts := []string{
		config.C.PackwizDir,
		modpack,
	}
	pathParts = append(pathParts, paths...)

	expectedFile := filepath.Join(pathParts...)

	return os.WriteFile(expectedFile, data, 0755)
}

func ChangeModSide(modpack, modName string, side types.ModSide) error {
	index, err := GetIndexFile(modpack)
	if err != nil {
		return err
	}

	var relativePath string

	for _, mod := range index.Files {
		parsedName := strings.TrimSuffix(filepath.Base(mod.File), ".pw.toml")
		if parsedName == modName {
			relativePath = mod.File
			break
		}
	}

	if relativePath == "" {
		return errors.New(fmt.Sprintf("pack: %s mod: %s not found", modpack, modName))
	}

	rawToml, err := readFile(modpack, relativePath)
	if err != nil {
		return err
	}

	var tomlData map[string]interface{}

	// Parse the TOML file into a map for manipulation
	err = toml.Unmarshal(rawToml, &tomlData)
	if err != nil {
		return fmt.Errorf("failed to parse TOML file: %w", err)
	}
	tomlData["side"] = side

	updatedContent, err := toml.Marshal(tomlData)
	if err != nil {
		return fmt.Errorf("failed to marshal TOML data: %w", err)
	}

	return writeFile(modpack, []string{relativePath}, updatedContent)
}
