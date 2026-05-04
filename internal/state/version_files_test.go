package state_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/foundry/fvm/internal/paths"
	"github.com/foundry/fvm/internal/state"
)

func TestWriteReadLocalVersion(t *testing.T) {
	dir := t.TempDir()

	if err := state.WriteLocalVersion(dir, "13.346"); err != nil {
		t.Fatalf("WriteLocalVersion: %v", err)
	}

	v, err := state.ReadLocalVersion(dir)
	if err != nil {
		t.Fatalf("ReadLocalVersion: %v", err)
	}
	if v != "13.346" {
		t.Fatalf("expected 13.346, got %q", v)
	}
}

func TestWriteLocalVersion_file_is_named_correctly(t *testing.T) {
	dir := t.TempDir()
	if err := state.WriteLocalVersion(dir, "1.0"); err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(dir, paths.LocalVersion)
	if _, err := os.Stat(expected); err != nil {
		t.Fatalf("expected %s to exist: %v", expected, err)
	}
}

func TestWriteLocalVersion_trims_whitespace(t *testing.T) {
	dir := t.TempDir()
	if err := state.WriteLocalVersion(dir, "  13.346  \n"); err != nil {
		t.Fatal(err)
	}
	v, err := state.ReadLocalVersion(dir)
	if err != nil {
		t.Fatal(err)
	}
	if v != "13.346" {
		t.Fatalf("expected 13.346, got %q", v)
	}
}

func TestWriteReadGlobalVersion(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "version")

	if err := state.WriteGlobalVersion(globalPath, "12.345"); err != nil {
		t.Fatalf("WriteGlobalVersion: %v", err)
	}

	v, err := state.ReadGlobalVersion(globalPath)
	if err != nil {
		t.Fatalf("ReadGlobalVersion: %v", err)
	}
	if v != "12.345" {
		t.Fatalf("expected 12.345, got %q", v)
	}
}

func TestWriteGlobalVersion_creates_parent_dirs(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "nested", "dirs", "version")

	if err := state.WriteGlobalVersion(globalPath, "1.0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(globalPath); err != nil {
		t.Fatalf("expected file at %s: %v", globalPath, err)
	}
}

func TestWriteGlobalVersion_trims_whitespace(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "version")

	if err := state.WriteGlobalVersion(globalPath, "\t 12.345 \n\n"); err != nil {
		t.Fatal(err)
	}
	v, err := state.ReadGlobalVersion(globalPath)
	if err != nil {
		t.Fatal(err)
	}
	if v != "12.345" {
		t.Fatalf("expected 12.345, got %q", v)
	}
}
