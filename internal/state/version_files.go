package state

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/foundry/fvm/internal/paths"
)

// WriteLocalVersion writes version to .fvm-version in dir.
// The value is trimmed of surrounding whitespace before writing.
func WriteLocalVersion(dir, version string) error {
	p := filepath.Join(dir, paths.LocalVersion)
	return os.WriteFile(p, []byte(normalize(version)), 0o644)
}

// ReadLocalVersion reads .fvm-version from dir and returns the trimmed version string.
func ReadLocalVersion(dir string) (string, error) {
	p := filepath.Join(dir, paths.LocalVersion)
	return readVersionFile(p)
}

// WriteGlobalVersion writes version to the global version file at globalVersionPath.
// Parent directories are created as needed.
func WriteGlobalVersion(globalVersionPath, version string) error {
	if err := os.MkdirAll(filepath.Dir(globalVersionPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(globalVersionPath, []byte(normalize(version)), 0o644)
}

// ReadGlobalVersion reads and returns the trimmed version string from globalVersionPath.
func ReadGlobalVersion(globalVersionPath string) (string, error) {
	return readVersionFile(globalVersionPath)
}

// normalize trims surrounding whitespace and appends exactly one newline.
func normalize(v string) string {
	return strings.TrimSpace(v) + "\n"
}

func readVersionFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
