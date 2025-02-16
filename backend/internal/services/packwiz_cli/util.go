package packwiz_cli

import (
	"errors"
	"fmt"
	"os"
	"packwiz-web/internal/config"
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
