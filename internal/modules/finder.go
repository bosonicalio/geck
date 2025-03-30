package modules

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindNearestGoModPath finds the nearest go.mod filepath in the parent directories.
func FindNearestGoModPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err = os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir { // reached the root of the filesystem
			return "", fmt.Errorf("go.mod not found in any parent directories")
		}
		dir = parent
	}
}
