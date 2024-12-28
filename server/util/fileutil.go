package util

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands a given path by expanding "~" to the user's home directory
// and resolving "." and ".." to a full absolute path.
func ExpandPath(p string) (string, error) {
	// Expand "~" to the home directory
	if strings.HasPrefix(p, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		p = filepath.Join(homeDir, p[1:]) // Replace ~ with home directory
	}

	// Resolve "." and ".." and convert the path to absolute
	absPath, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}

	return absPath, nil
}

// CountFilesInDir counts the number of files in a given directory
func CountFilesInDir(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	// Count files (not directories)
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() { // Only count regular files
			count++
		}
	}

	return count, nil
}

// GetFilesInDir creates a slice of files that exist in a directory
func GetFilesInDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// Create a slice to store the file names
	var files []string

	// Loop through the directory entries
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}
