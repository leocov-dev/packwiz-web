package packwiz_cli

import (
	"errors"
	"fmt"
	"os"
	"packwiz-web/internal/config"
	"packwiz-web/internal/utils"
	"path/filepath"
)

func NewModpack(slug, name, author string, minecraft MinecraftDef, loader LoaderDef) error {
	args := []string{
		"init",
		"--name", name,
		"--author", author,
		"--version", "1.0.0",
	}

	args = append(args, minecraft.AsArgs()...)
	args = append(args, loader.AsArgs()...)

	modpackDir := filepath.Join(config.C.PackwizDir, slug)
	if utils.FileExists(filepath.Join(modpackDir, "pack.toml")) {
		return errors.New(fmt.Sprintf("Modpack '%s' already exists", name))
	}

	err := os.MkdirAll(modpackDir, 0755)
	if err != nil {
		return err
	}

	return runCommand(name, args...)
}
