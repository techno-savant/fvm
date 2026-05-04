package install_test

import (
	"path/filepath"
	"testing"

	"github.com/foundry/fvm/internal/install"
)

func TestLayout_paths(t *testing.T) {
	dir := t.TempDir()
	vDir := filepath.Join(dir, "versions", "13.346")
	l := install.NewLayout(vDir)

	if got := l.FoundryPath(); got != filepath.Join(vDir, "foundry") {
		t.Errorf("FoundryPath: got %q", got)
	}
	if got := l.BinPath(); got != filepath.Join(vDir, "bin") {
		t.Errorf("BinPath: got %q", got)
	}
	if got := l.ManifestPath(); got != filepath.Join(vDir, "manifest.json") {
		t.Errorf("ManifestPath: got %q", got)
	}
	if got := l.CompletePath(); got != filepath.Join(vDir, ".complete") {
		t.Errorf("CompletePath: got %q", got)
	}
}

func TestLayout_complete_marker_lifecycle(t *testing.T) {
	dir := t.TempDir()
	l := install.NewLayout(dir)

	if l.IsComplete() {
		t.Fatal("expected incomplete before marking")
	}

	if err := l.MarkComplete(); err != nil {
		t.Fatalf("MarkComplete: %v", err)
	}
	if !l.IsComplete() {
		t.Fatal("expected complete after MarkComplete")
	}

	if err := l.RemoveComplete(); err != nil {
		t.Fatalf("RemoveComplete: %v", err)
	}
	if l.IsComplete() {
		t.Fatal("expected incomplete after RemoveComplete")
	}
}

func TestLayout_remove_complete_idempotent(t *testing.T) {
	dir := t.TempDir()
	l := install.NewLayout(dir)

	// Should not error even when .complete doesn't exist.
	if err := l.RemoveComplete(); err != nil {
		t.Fatalf("RemoveComplete on missing file: %v", err)
	}
}

func TestLayout_mark_complete_creates_dirs(t *testing.T) {
	dir := t.TempDir()
	l := install.NewLayout(filepath.Join(dir, "nested", "version"))

	if err := l.MarkComplete(); err != nil {
		t.Fatalf("MarkComplete with nested dir: %v", err)
	}
	if !l.IsComplete() {
		t.Fatal("expected complete after MarkComplete with nested dir")
	}
}
