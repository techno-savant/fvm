package releases_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/foundry/fvm/internal/install"
	"github.com/foundry/fvm/internal/releases"
)

func TestListInstalled_empty_when_no_versions_dir(t *testing.T) {
	dir := t.TempDir()
	vs, err := releases.ListInstalled(filepath.Join(dir, "versions"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vs) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(vs))
	}
}

func TestListInstalled_skips_incomplete(t *testing.T) {
	dir := t.TempDir()
	vDir := filepath.Join(dir, "13.346")

	// Write manifest but no .complete marker.
	if err := install.WriteManifest(vDir, install.Manifest{Version: "13.346"}); err != nil {
		t.Fatal(err)
	}

	vs, err := releases.ListInstalled(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vs) != 0 {
		t.Fatalf("expected empty list (incomplete install), got %d", len(vs))
	}
}

func TestListInstalled_includes_complete(t *testing.T) {
	dir := t.TempDir()
	vDir := filepath.Join(dir, "13.346")
	l := install.NewLayout(vDir)

	if err := install.WriteManifest(vDir, install.Manifest{Version: "13.346"}); err != nil {
		t.Fatal(err)
	}
	if err := l.MarkComplete(); err != nil {
		t.Fatal(err)
	}

	vs, err := releases.ListInstalled(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vs) != 1 {
		t.Fatalf("expected 1 installed version, got %d", len(vs))
	}
	if vs[0].Version != "13.346" {
		t.Fatalf("expected 13.346, got %q", vs[0].Version)
	}
}

func TestListInstalled_multiple_versions(t *testing.T) {
	dir := t.TempDir()
	for _, ver := range []string{"12.345", "13.346"} {
		vDir := filepath.Join(dir, ver)
		l := install.NewLayout(vDir)
		if err := install.WriteManifest(vDir, install.Manifest{Version: ver}); err != nil {
			t.Fatal(err)
		}
		if err := l.MarkComplete(); err != nil {
			t.Fatal(err)
		}
	}

	vs, err := releases.ListInstalled(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vs) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(vs))
	}
}

func TestStubProvider_list_remote(t *testing.T) {
	p := &releases.StubProvider{Versions: []string{"13.346", "12.345"}}
	vs, err := p.ListRemote()
	if err != nil {
		t.Fatalf("ListRemote: %v", err)
	}
	if len(vs) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(vs))
	}
}

func TestStubProvider_install_returns_error(t *testing.T) {
	p := &releases.StubProvider{}
	if err := p.Install("13.346", t.TempDir()); err == nil {
		t.Fatal("expected error from stub installer")
	}
}

func TestNewInstallRecord_creates_tmp_dir(t *testing.T) {
	dir := t.TempDir()
	versionsRoot := filepath.Join(dir, "versions")
	tmpBase := filepath.Join(dir, "tmp")

	rec, err := releases.NewInstallRecord("13.346", versionsRoot, tmpBase)
	if err != nil {
		t.Fatalf("NewInstallRecord: %v", err)
	}
	if rec.TmpDir == "" {
		t.Fatal("expected non-empty TmpDir")
	}
}

func TestInstallRecord_finalize(t *testing.T) {
	dir := t.TempDir()
	versionsRoot := filepath.Join(dir, "versions")
	tmpBase := filepath.Join(dir, "tmp")

	rec, err := releases.NewInstallRecord("13.346", versionsRoot, tmpBase)
	if err != nil {
		t.Fatal(err)
	}

	m := install.Manifest{
		Version:     "13.346",
		Platform:    "linux",
		Arch:        "amd64",
		InstalledAt: time.Now().UTC(),
	}
	if err := rec.Finalize(m); err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if !rec.Layout.IsComplete() {
		t.Fatal("expected install to be complete after Finalize")
	}

	got, err := install.ReadManifest(rec.Layout.Root)
	if err != nil {
		t.Fatalf("ReadManifest: %v", err)
	}
	if got.Version != "13.346" {
		t.Fatalf("expected 13.346, got %q", got.Version)
	}
}

func TestInstallRecord_cleanup(t *testing.T) {
	dir := t.TempDir()
	rec, err := releases.NewInstallRecord("13.346", filepath.Join(dir, "versions"), filepath.Join(dir, "tmp"))
	if err != nil {
		t.Fatal(err)
	}
	if err := rec.Cleanup(); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
}

func TestIsVersionString(t *testing.T) {
	cases := []struct {
		input string
		valid bool
	}{
		{"13.346", true},
		{"12.345", true},
		{"1.0.0", true},
		{"11", true},
		{"", false},
		{"abc", false},
		{"13.346-beta", false},
		{"  ", false},
	}
	for _, c := range cases {
		if got := releases.IsVersionString(c.input); got != c.valid {
			t.Errorf("IsVersionString(%q) = %v, want %v", c.input, got, c.valid)
		}
	}
}
