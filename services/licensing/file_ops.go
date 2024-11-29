package licensing

import (
	"os"
	"path/filepath"
)

// ensureDir creates a directory if it doesn't exist
func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// writeFile writes data to a file, creating parent directories if needed
func writeFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := ensureDir(dir); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// deleteFile removes a file
func deleteFile(path string) error {
	return os.Remove(path)
}

// getFileInfo returns file information
func getFileInfo(path string) (FileInfo, error) {
	return os.Stat(path)
}
