package releases

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/foundry/fvm/internal/install"
	"github.com/foundry/fvm/internal/paths"
)

// ErrNotInstalled is returned when a requested version is not found locally.
var ErrNotInstalled = errors.New("version not installed")

// InstalledVersion describes a locally installed, complete version.
type InstalledVersion struct {
	Version string
	Layout  install.Layout
}

// ListInstalled scans versionsRoot and returns all complete installed versions.
// Directories lacking a .complete marker are silently skipped.
// Returns nil (not an error) when versionsRoot does not exist yet.
func ListInstalled(versionsRoot string) ([]InstalledVersion, error) {
	entries, err := os.ReadDir(versionsRoot)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var result []InstalledVersion
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		version := e.Name()
		l := install.NewLayout(filepath.Join(versionsRoot, version))
		if !l.IsComplete() {
			continue
		}
		result = append(result, InstalledVersion{
			Version: version,
			Layout:  l,
		})
	}
	return result, nil
}

// DefaultVersionsRoot returns the standard versions directory path.
func DefaultVersionsRoot() string {
	return paths.VersionsPath()
}

// Provider is the interface for listing and fetching remote Foundry versions.
// Implementations may talk to an API, a registry file, or a stub.
type Provider interface {
	ListRemote() ([]string, error)
	Install(version, destDir string) error
}

// StubProvider is a no-op Provider used during scaffolding and testing.
type StubProvider struct {
	Versions []string
}

// ListRemote returns the pre-configured stub version list.
func (s *StubProvider) ListRemote() ([]string, error) {
	return s.Versions, nil
}

// Install always returns an error — the stub does not perform real downloads.
func (s *StubProvider) Install(version, destDir string) error {
	return errors.New("install: real provider not yet implemented")
}

// InstallRecord manages the lifecycle of a single version install.
type InstallRecord struct {
	Version string
	Layout  install.Layout
	TmpDir  string
}

// NewInstallRecord prepares an InstallRecord for version under versionsRoot.
// A temporary working directory is created under tmpBase.
func NewInstallRecord(version, versionsRoot, tmpBase string) (InstallRecord, error) {
	versionDir := filepath.Join(versionsRoot, version)
	tmp := filepath.Join(tmpBase, "install-"+version)
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return InstallRecord{}, err
	}
	return InstallRecord{
		Version: version,
		Layout:  install.NewLayout(versionDir),
		TmpDir:  tmp,
	}, nil
}

// Finalize writes the manifest and marks the install complete.
// Must be called only after all files have been placed under Layout.Root.
func (r *InstallRecord) Finalize(m install.Manifest) error {
	if err := os.MkdirAll(r.Layout.Root, 0o755); err != nil {
		return err
	}
	if err := install.WriteManifest(r.Layout.Root, m); err != nil {
		return err
	}
	return r.Layout.MarkComplete()
}

// Cleanup removes the temporary working directory.
func (r *InstallRecord) Cleanup() error {
	return os.RemoveAll(r.TmpDir)
}

// IsVersionString returns true if s looks like a Foundry version number
// (digits and dots only, non-empty).
func IsVersionString(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && c != '.' {
			return false
		}
	}
	return true
}
