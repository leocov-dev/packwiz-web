package packwiz_cli

import (
	"os/exec"
	"packwiz-web/internal/config"
	"packwiz-web/internal/log"
	"path/filepath"
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
func runCommand(modpack string, args ...string) error {
	// HTTP requests should not modify the packwiz_cli data files on disk at the
	// same time. Some operations may take a longer time than others, we need to
	// evaluate if the lock is being held for too long and what can be done.
	mutex := getModpackMutex(modpack)
	mutex.Lock()
	defer mutex.Unlock()

	// always run in non-interactive mode
	args = append([]string{"--yes"}, args...)

	cmd := exec.Command("packwiz", args...)

	log.Debug("Execute:", cmd.Path, cmd.Args)

	// setting Dir here changes the execution context to be the mod's pack folder
	cmd.Dir = filepath.Join(config.C.PackwizDir, modpack)
	output, err := cmd.CombinedOutput()

	log.Debug("Output: ", output)
	if err != nil {
		log.Debug("Error: ", err)
		return err
	}

	return nil
}
