package packwiz_cli_disabled

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"packwiz-web/internal/config"
	"packwiz-web/internal/log"
	"packwiz-web/internal/utils"
	"path/filepath"
	"strings"
	"sync"
)

// TODO: this restricts the backend to a single instance since the mutex is in
//  memory. We won't deal with this now since the better solution is to convert
//  the packwiz cli codebase to be usable as a library that does not write any
//  files to disk.

// modpackMutexes maps modpack names to their respective mutexes
var modpackMutexes = make(map[string]*sync.Mutex)
var mutexMapLock sync.Mutex

func getModpackMutex(modpack string) *sync.Mutex {
	mutexMapLock.Lock()
	defer mutexMapLock.Unlock()

	mutex, exists := modpackMutexes[modpack]
	if !exists {
		mutex = &sync.Mutex{}
		modpackMutexes[modpack] = mutex
	}
	return mutex
}

// runCommand run a packwiz_cli command (has different context based on active dir)
func runCommand(modpack string, args ...string) (string, error) {
	// HTTP requests should not modify the packwiz_cli data files on disk at the
	// same time. Some operations may take a longer time than others, we need to
	// evaluate if the lock is being held for too long and what can be done.
	mutex := getModpackMutex(modpack)
	mutex.Lock()
	defer mutex.Unlock()

	// always run in non-interactive mode
	args = append([]string{"--yes"}, args...)

	exePath, _ := os.Executable()
	packwizPath := filepath.Join(filepath.Dir(exePath), "packwiz")
	if !utils.FileExists(packwizPath) {
		packwizPath = "packwiz"
	}

	cmd := exec.Command(packwizPath, args...)

	log.Debug("Execute:", cmd.Path, cmd.Args)

	// setting Dir here changes the execution context to be the mod's pack folder
	cmd.Dir = filepath.Join(config.C.PackwizDir, modpack)
	output, err := cmd.CombinedOutput()

	outputText := string(output)

	if err != nil {
		errorMsg := fmt.Sprintf("%s - %s", strings.Replace(outputText, "\n", ", ", -1), err)
		log.Error("Error: ", errorMsg)
		return outputText, errors.New(errorMsg)
	} else {
		log.Debug("Output: ", string(output))
	}

	return outputText, nil
}
