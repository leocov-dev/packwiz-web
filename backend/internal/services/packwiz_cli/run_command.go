package packwiz_cli

import (
	"os/exec"
	"packwiz-web/internal/config"
	"packwiz-web/internal/logger"
	"path/filepath"
	"sync"
)

var execMutex sync.Mutex

// runCommand run a packwiz_cli command (has different context based on active dir)
func runCommand(modpack string, args ...string) error {
	// HTTP requests should not modify the packwiz_cli data files on disk at the
	// same time. Some operations may take a longer time than others, we need to
	// evaluate if the lock is being held for too long and what can be done.
	execMutex.Lock()
	defer execMutex.Unlock()

	// always run in non-interactive mode
	args = append([]string{"--yes"}, args...)

	cmd := exec.Command("packwiz", args...)

	logger.Debug("Execute:", cmd.Path, cmd.Args)

	// setting Dir here changes the execution context to be the mod's pack folder
	cmd.Dir = filepath.Join(config.C.PackwizDir, modpack)
	output, err := cmd.CombinedOutput()

	logger.Debug("Output: ", output)
	if err != nil {
		logger.Debug("Error: ", err)
		return err
	}

	return nil
}
