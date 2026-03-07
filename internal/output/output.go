package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolvePath determines the final output file path, creating directories as
// needed and appending numeric suffixes to avoid overwriting existing files.
func ResolvePath(dir, filename string) (string, error) {
	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getting current working directory: %w", err)
		}
		dir = cwd
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating output directory: %w", err)
	}

	ext := filepath.Ext(filename)
	if ext == "" {
		filename += ".png"
	} else if !strings.EqualFold(ext, ".png") {
		return "", fmt.Errorf("output file must have .png extension")
	}

	path := filepath.Join(dir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path, nil
	}

	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	for i := 2; ; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s%d.png", base, i))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		}
	}
}

// WriteFile writes data to path with 0644 permissions.
func WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}
