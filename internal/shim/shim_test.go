package shim_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/foundry/fvm/internal/shim"
)

func TestGenerate_creates_executable_shims(t *testing.T) {
	dir := t.TempDir()
	names := []string{"foundry", "foundryvtt"}

	if err := shim.Generate(dir, names); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	for _, name := range names {
		p := filepath.Join(dir, name)
		info, err := os.Stat(p)
		if err != nil {
			t.Fatalf("expected shim %s to exist: %v", name, err)
		}
		if info.Mode()&0o100 == 0 {
			t.Fatalf("expected shim %s to be executable", name)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(data), "#!/bin/sh") {
			t.Fatalf("expected shim %s to start with #!/bin/sh", name)
		}
		if !strings.Contains(string(data), name) {
			t.Fatalf("expected shim %s content to reference %q", name, name)
		}
	}
}

func TestGenerate_overwrites_existing(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "foundry")
	if err := os.WriteFile(p, []byte("old content"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := shim.Generate(dir, []string{"foundry"}); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "old content") {
		t.Fatal("expected old content to be overwritten")
	}
}

func TestGenerate_creates_shim_dir(t *testing.T) {
	dir := t.TempDir()
	shimDir := filepath.Join(dir, "shims")

	if err := shim.Generate(shimDir, []string{"foundry"}); err != nil {
		t.Fatalf("Generate with missing shim dir: %v", err)
	}

	if _, err := os.Stat(shimDir); err != nil {
		t.Fatalf("expected shim dir to be created: %v", err)
	}
}

func TestDefaultNames(t *testing.T) {
	names := shim.DefaultNames()
	if len(names) == 0 {
		t.Fatal("expected non-empty default names")
	}
	found := false
	for _, n := range names {
		if n == "foundry" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected 'foundry' in default names")
	}
}
