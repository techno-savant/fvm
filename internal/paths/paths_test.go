package paths_test

import (
	"strings"
	"testing"

	"github.com/foundry/fvm/internal/paths"
)

func TestRoot_default(t *testing.T) {
	t.Setenv("FVM_ROOT", "")
	r := paths.Root()
	if !strings.HasSuffix(r, "/.fvm") {
		t.Fatalf("expected root to end in /.fvm, got %q", r)
	}
}

func TestRoot_override(t *testing.T) {
	t.Setenv("FVM_ROOT", "/tmp/test-fvm")
	r := paths.Root()
	if r != "/tmp/test-fvm" {
		t.Fatalf("expected /tmp/test-fvm, got %q", r)
	}
}

func TestConfigPath(t *testing.T) {
	t.Setenv("FVM_ROOT", "/tmp/test-fvm")
	if p := paths.ConfigPath(); p != "/tmp/test-fvm/config.yaml" {
		t.Fatalf("unexpected config path: %q", p)
	}
}

func TestGlobalVersionPath(t *testing.T) {
	t.Setenv("FVM_ROOT", "/tmp/test-fvm")
	if p := paths.GlobalVersionPath(); p != "/tmp/test-fvm/version" {
		t.Fatalf("unexpected global version path: %q", p)
	}
}

func TestShimsPath(t *testing.T) {
	t.Setenv("FVM_ROOT", "/tmp/test-fvm")
	if p := paths.ShimsPath(); p != "/tmp/test-fvm/shims" {
		t.Fatalf("unexpected shims path: %q", p)
	}
}

func TestVersionsPath(t *testing.T) {
	t.Setenv("FVM_ROOT", "/tmp/test-fvm")
	if p := paths.VersionsPath(); p != "/tmp/test-fvm/versions" {
		t.Fatalf("unexpected versions path: %q", p)
	}
}

func TestVersionDir(t *testing.T) {
	t.Setenv("FVM_ROOT", "/tmp/test-fvm")
	if p := paths.VersionDir("13.346"); p != "/tmp/test-fvm/versions/13.346" {
		t.Fatalf("unexpected version dir: %q", p)
	}
}

func TestLocalVersionFilename(t *testing.T) {
	if paths.LocalVersion != ".fvm-version" {
		t.Fatalf("expected .fvm-version, got %q", paths.LocalVersion)
	}
}
