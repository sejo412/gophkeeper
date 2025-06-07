package helpers

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// SaveRegularFile writes file with creating parent dir.
func SaveRegularFile(path string, content []byte, perms fs.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dir, err)
	}
	if err := os.WriteFile(path, content, perms); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}
	return nil
}
