package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/foundry/fvm/internal/config"
)

func TestDefaults(t *testing.T) {
	cfg := config.Defaults()
	if cfg.ReleaseChannel != "stable" {
		t.Fatalf("expected stable channel, got %q", cfg.ReleaseChannel)
	}
	if len(cfg.ShimNames) == 0 {
		t.Fatal("expected non-empty shim names")
	}
	if !cfg.AutoRegenerateShims {
		t.Fatal("expected auto_regenerate_shims to be true by default")
	}
	if !cfg.PreferOfficialBuilds {
		t.Fatal("expected prefer_official_builds to be true by default")
	}
	if cfg.FoundryPlatform != "" {
		t.Fatalf("expected empty foundry_platform config default, got %q", cfg.FoundryPlatform)
	}
}

func TestLoad_missing_file_returns_defaults(t *testing.T) {
	cfg, err := config.Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if cfg.ReleaseChannel != "stable" {
		t.Fatalf("expected stable, got %q", cfg.ReleaseChannel)
	}
}

func TestLoad_overrides_fields(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	content := "release_channel: nightly\ncache_ttl: 1h\nfoundry_cookie: sessionid=abc; csrftoken=def\nfoundry_username: savant\nfoundry_password: hunter2\nfoundry_platform: windows_portable\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ReleaseChannel != "nightly" {
		t.Fatalf("expected nightly, got %q", cfg.ReleaseChannel)
	}
	if cfg.CacheTTL != "1h" {
		t.Fatalf("expected 1h, got %q", cfg.CacheTTL)
	}
	if cfg.FoundryCookie != "sessionid=abc; csrftoken=def" {
		t.Fatalf("expected foundry_cookie override, got %q", cfg.FoundryCookie)
	}
	if cfg.FoundryUsername != "savant" {
		t.Fatalf("expected foundry_username override, got %q", cfg.FoundryUsername)
	}
	if cfg.FoundryPassword != "hunter2" {
		t.Fatalf("expected foundry_password override, got %q", cfg.FoundryPassword)
	}
	if cfg.FoundryPlatform != "windows_portable" {
		t.Fatalf("expected foundry_platform override, got %q", cfg.FoundryPlatform)
	}
	if !cfg.AutoRegenerateShims {
		t.Fatal("expected auto_regenerate_shims to keep default true")
	}
}

func TestLoad_shim_names_override(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	content := "shim_names:\n  - forge\n  - cast\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.ShimNames) != 2 {
		t.Fatalf("expected 2 shim names, got %d", len(cfg.ShimNames))
	}
	if cfg.ShimNames[0] != "forge" {
		t.Fatalf("expected forge, got %q", cfg.ShimNames[0])
	}
}
