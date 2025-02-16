package utils

import "os"

func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		// If the error is "not exists", the directory doesn't exist
		if os.IsNotExist(err) {
			return false
		}
		// Other errors (e.g., permission issues) are treated as "does not exist"
		return false
	}
	// Check if the path is a directory
	return info.IsDir()
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		// If the error is "not exists", the file doesn't exist
		if os.IsNotExist(err) {
			return false
		}
		// Other errors (e.g., permission issues) are treated as "does not exist"
		return false
	}
	return true
}
