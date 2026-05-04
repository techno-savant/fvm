package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ManifestFile is the filename of the per-version install manifest.
const ManifestFile = "manifest.json"

// Manifest holds per-version install metadata stored as manifest.json.
type Manifest struct {
	Version        string    `json:"version"`
	SourceURL      string    `json:"source_url"`
	InstalledAt    time.Time `json:"installed_at"`
	Platform       string    `json:"platform"`
	Arch           string    `json:"arch"`
	Checksum       string    `json:"checksum"`
	ExecutablePath string    `json:"executable_path"`
}

// WriteManifest marshals m to JSON and writes it to <dir>/manifest.json.
// The directory is created if it does not exist.
func WriteManifest(dir string, m Manifest) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ManifestFile), data, 0o644)
}

// ReadManifest reads and parses <dir>/manifest.json.
func ReadManifest(dir string) (Manifest, error) {
	data, err := os.ReadFile(filepath.Join(dir, ManifestFile))
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, err
	}
	return m, nil
}
