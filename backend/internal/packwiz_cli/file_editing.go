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

func writeFile(slug string, paths []string, data []byte) error {
	pathParts := []string{
		config.C.PackwizDir,
		slug,
	}
	pathParts = append(pathParts, paths...)

	expectedFile := filepath.Join(pathParts...)

	return os.WriteFile(expectedFile, data, 0755)
}

func RenamePack(slug, newName string) error {
	rawToml, err := readFile(slug, "pack.toml")
	if err != nil {
		return err
	}

	var tomlData map[string]interface{}

	// Parse the TOML file into a map for manipulation
	err = toml.Unmarshal(rawToml, &tomlData)
	if err != nil {
		return fmt.Errorf("failed to parse TOML file: %w", err)
	}
	if tomlData["name"].(string) == newName {
		return nil
	}

	tomlData["name"] = newName

	updatedContent, err := toml.Marshal(tomlData)
	if err != nil {
		return fmt.Errorf("failed to marshal TOML data: %w", err)
	}

	return writeFile(slug, []string{"pack.toml"}, updatedContent)
}

func ChangeModSide(slug, modName string, side types.ModSide) error {
	index, err := GetIndexFile(slug)
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
		return errors.New(fmt.Sprintf("pack: %s mod: %s not found", slug, modName))
	}

	rawToml, err := readFile(slug, relativePath)
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

	return writeFile(slug, []string{relativePath}, updatedContent)
}
