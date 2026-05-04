package resolve_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/foundry/fvm/internal/paths"
	"github.com/foundry/fvm/internal/resolve"
)

func TestShellOverrideBeatsAll(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, paths.LocalVersion), "local-version")
	globalFile := filepath.Join(dir, "global-version")
	writeFile(t, globalFile, "global-version")

	res, err := resolve.Resolve(resolve.Inputs{
		ShellOverride:     "shell-version",
		Cwd:               dir,
		GlobalVersionPath: globalFile,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Version != "shell-version" {
		t.Fatalf("expected shell-version, got %q", res.Version)
	}
	if res.Source != resolve.SourceShellOverride {
		t.Fatalf("expected SourceShellOverride, got %v", res.Source)
	}
}

func TestLocalBeatsGlobal(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, paths.LocalVersion), "local-version")
	globalFile := filepath.Join(dir, "global-version")
	writeFile(t, globalFile, "global-version")

	res, err := resolve.Resolve(resolve.Inputs{
		Cwd:               dir,
		GlobalVersionPath: globalFile,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Version != "local-version" {
		t.Fatalf("expected local-version, got %q", res.Version)
	}
	if res.Source != resolve.SourceLocal {
		t.Fatalf("expected SourceLocal, got %v", res.Source)
	}
}

func TestGlobalUsedWhenLocalAbsent(t *testing.T) {
	dir := t.TempDir()
	globalFile := filepath.Join(dir, "global-version")
	writeFile(t, globalFile, "global-version")

	res, err := resolve.Resolve(resolve.Inputs{
		Cwd:               dir,
		GlobalVersionPath: globalFile,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Version != "global-version" {
		t.Fatalf("expected global-version, got %q", res.Version)
	}
	if res.Source != resolve.SourceGlobal {
		t.Fatalf("expected SourceGlobal, got %v", res.Source)
	}
}

func TestErrorWhenNothingResolves(t *testing.T) {
	dir := t.TempDir()

	_, err := resolve.Resolve(resolve.Inputs{
		Cwd:               dir,
		GlobalVersionPath: filepath.Join(dir, "nonexistent"),
	})
	if err == nil {
		t.Fatal("expected error when nothing resolves")
	}
	if err != resolve.ErrNoVersion {
		t.Fatalf("expected ErrNoVersion, got %v", err)
	}
}

func TestLocalFoundInAncestorDirectory(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, paths.LocalVersion), "ancestor-version")

	res, err := resolve.Resolve(resolve.Inputs{
		Cwd:               sub,
		GlobalVersionPath: filepath.Join(root, "nonexistent"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Version != "ancestor-version" {
		t.Fatalf("expected ancestor-version, got %q", res.Version)
	}
	if res.Source != resolve.SourceLocal {
		t.Fatalf("expected SourceLocal, got %v", res.Source)
	}
}

func TestLocalFilePathRecorded(t *testing.T) {
	dir := t.TempDir()
	localFile := filepath.Join(dir, paths.LocalVersion)
	writeFile(t, localFile, "13.346")

	res, err := resolve.Resolve(resolve.Inputs{
		Cwd:               dir,
		GlobalVersionPath: filepath.Join(dir, "nonexistent"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.File != localFile {
		t.Fatalf("expected file %q, got %q", localFile, res.File)
	}
}

func TestSourceString(t *testing.T) {
	cases := []struct {
		s    resolve.Source
		want string
	}{
		{resolve.SourceShellOverride, "shell override"},
		{resolve.SourceLocal, "local (.fvm-version)"},
		{resolve.SourceGlobal, "global (~/.fvm/version)"},
	}
	for _, c := range cases {
		if got := c.s.String(); got != c.want {
			t.Errorf("Source(%d).String() = %q, want %q", c.s, got, c.want)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
