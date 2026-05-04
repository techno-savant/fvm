package install_test

import (
	"testing"
	"time"

	"github.com/foundry/fvm/internal/install"
)

func TestWriteReadManifest(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)

	m := install.Manifest{
		Version:        "13.346",
		SourceURL:      "https://example.com/foundry-13.346.tar.gz",
		InstalledAt:    now,
		Platform:       "darwin",
		Arch:           "arm64",
		Checksum:       "abc123",
		ExecutablePath: "foundry/resources/app",
	}

	if err := install.WriteManifest(dir, m); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}

	got, err := install.ReadManifest(dir)
	if err != nil {
		t.Fatalf("ReadManifest: %v", err)
	}

	if got.Version != m.Version {
		t.Errorf("Version: got %q, want %q", got.Version, m.Version)
	}
	if got.Platform != m.Platform {
		t.Errorf("Platform: got %q, want %q", got.Platform, m.Platform)
	}
	if got.Arch != m.Arch {
		t.Errorf("Arch: got %q, want %q", got.Arch, m.Arch)
	}
	if got.Checksum != m.Checksum {
		t.Errorf("Checksum: got %q, want %q", got.Checksum, m.Checksum)
	}
	if got.ExecutablePath != m.ExecutablePath {
		t.Errorf("ExecutablePath: got %q, want %q", got.ExecutablePath, m.ExecutablePath)
	}
	if !got.InstalledAt.Equal(m.InstalledAt) {
		t.Errorf("InstalledAt: got %v, want %v", got.InstalledAt, m.InstalledAt)
	}
}

func TestWriteManifest_creates_dir(t *testing.T) {
	dir := t.TempDir()
	target := dir + "/versions/13.346"

	if err := install.WriteManifest(target, install.Manifest{Version: "13.346"}); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}

	got, err := install.ReadManifest(target)
	if err != nil {
		t.Fatalf("ReadManifest: %v", err)
	}
	if got.Version != "13.346" {
		t.Fatalf("expected 13.346, got %q", got.Version)
	}
}
