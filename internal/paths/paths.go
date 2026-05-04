package paths

import (
	"os"
	"path/filepath"
)

const (
	// LocalVersion is the project-local version file name.
	LocalVersion = ".fvm-version"

	fvmDir       = ".fvm"
	configFile   = "config.yaml"
	globalVer    = "version"
	shimsDir     = "shims"
	versionsDir  = "versions"
	downloadsDir = "downloads"
	tmpDir       = "tmp"
	registryDir  = "registry"
	logDir       = "log"
)

// Root returns the fvm home directory.
// Defaults to ~/.fvm; overridable via FVM_ROOT for testing.
func Root() string {
	if r := os.Getenv("FVM_ROOT"); r != "" {
		return r
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, fvmDir)
}

// ConfigPath returns the path to ~/.fvm/config.yaml.
func ConfigPath() string {
	return filepath.Join(Root(), configFile)
}

// GlobalVersionPath returns the path to ~/.fvm/version.
func GlobalVersionPath() string {
	return filepath.Join(Root(), globalVer)
}

// ShimsPath returns the path to ~/.fvm/shims.
func ShimsPath() string {
	return filepath.Join(Root(), shimsDir)
}

// VersionsPath returns the path to ~/.fvm/versions.
func VersionsPath() string {
	return filepath.Join(Root(), versionsDir)
}

// VersionDir returns the path to ~/.fvm/versions/<version>.
func VersionDir(version string) string {
	return filepath.Join(VersionsPath(), version)
}

// DownloadsPath returns the path to ~/.fvm/downloads.
func DownloadsPath() string {
	return filepath.Join(Root(), downloadsDir)
}

// TmpPath returns the path to ~/.fvm/tmp.
func TmpPath() string {
	return filepath.Join(Root(), tmpDir)
}

// RegistryPath returns the path to ~/.fvm/registry.
func RegistryPath() string {
	return filepath.Join(Root(), registryDir)
}

// LogPath returns the path to ~/.fvm/log.
func LogPath() string {
	return filepath.Join(Root(), logDir)
}
