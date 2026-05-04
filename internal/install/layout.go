package install

import (
	"os"
	"path/filepath"
)

const (
	// FoundryDir is the subdirectory holding the Foundry application.
	FoundryDir = "foundry"
	// BinDir is the subdirectory holding version-local helper binaries.
	BinDir = "bin"
	// CompleteFile is the sentinel marker written last during a successful install.
	CompleteFile = ".complete"
)

// Layout describes the filesystem layout for one installed version.
// Root is typically ~/.fvm/versions/<version>.
type Layout struct {
	Root string
}

// NewLayout returns a Layout rooted at versionDir.
func NewLayout(versionDir string) Layout {
	return Layout{Root: versionDir}
}

// FoundryPath returns <root>/foundry.
func (l Layout) FoundryPath() string {
	return filepath.Join(l.Root, FoundryDir)
}

// BinPath returns <root>/bin.
func (l Layout) BinPath() string {
	return filepath.Join(l.Root, BinDir)
}

// ManifestPath returns <root>/manifest.json.
func (l Layout) ManifestPath() string {
	return filepath.Join(l.Root, ManifestFile)
}

// CompletePath returns <root>/.complete.
func (l Layout) CompletePath() string {
	return filepath.Join(l.Root, CompleteFile)
}

// IsComplete reports whether the install has a .complete sentinel marker.
func (l Layout) IsComplete() bool {
	_, err := os.Stat(l.CompletePath())
	return err == nil
}

// MarkComplete creates the .complete sentinel file, creating parent dirs as needed.
func (l Layout) MarkComplete() error {
	if err := os.MkdirAll(l.Root, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(l.CompletePath(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	return f.Close()
}

// RemoveComplete removes the .complete marker, making the install appear incomplete.
// Returns nil if the file did not exist.
func (l Layout) RemoveComplete() error {
	err := os.Remove(l.CompletePath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
