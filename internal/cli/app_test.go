package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── mock implementations ──────────────────────────────────────────────────────

type mockResolver struct {
	result *ResolvedVersion
	err    error
}

func (m *mockResolver) Current(_ string) (*ResolvedVersion, error) { return m.result, m.err }

type mockState struct {
	localVersion       string
	localErr           error
	globalVersion      string
	globalErr          error
	writeLocalCalled   bool
	writeLocalVersion  string
	writeGlobalCalled  bool
	writeGlobalVersion string
}

func (m *mockState) ReadLocal(_ string) (string, error) { return m.localVersion, m.localErr }
func (m *mockState) ReadGlobal() (string, error)        { return m.globalVersion, m.globalErr }
func (m *mockState) WriteLocal(_, version string) error {
	m.writeLocalCalled = true
	m.writeLocalVersion = version
	return nil
}
func (m *mockState) WriteGlobal(version string) error {
	m.writeGlobalCalled = true
	m.writeGlobalVersion = version
	return nil
}

type mockRegistry struct {
	versions  []string
	installed map[string]bool
	dirs      map[string]string
	exePaths  map[string]string
}

func (m *mockRegistry) InstalledVersions() ([]string, error) { return m.versions, nil }
func (m *mockRegistry) IsInstalled(v string) bool            { return m.installed[v] }
func (m *mockRegistry) VersionDir(v string) string {
	if d := m.dirs[v]; d != "" {
		return d
	}
	return "/fake/.fvm/versions/" + v
}
func (m *mockRegistry) ExecutablePath(v string) (string, error) {
	if p := m.exePaths[v]; p != "" {
		return p, nil
	}
	return "", errors.New("executable not found")
}

type mockInstaller struct{ err error }

func (m *mockInstaller) Install(_ string) error { return m.err }

type mockShims struct {
	dir string
	err error
}

func (m *mockShims) ShimDir() string   { return m.dir }
func (m *mockShims) Regenerate() error { return m.err }

type mockRemote struct {
	versions []string
	err      error
}

func (m *mockRemote) RemoteVersions() ([]string, error) { return m.versions, m.err }

type mockDoctor struct{ results []DoctorResult }

func (m *mockDoctor) Check() []DoctorResult { return m.results }

// ── test helper ──────────────────────────────────────────────────────────────

func makeApp() (*App, *bytes.Buffer, *bytes.Buffer) {
	var out, errOut bytes.Buffer
	app := &App{
		Stdout:    &out,
		Stderr:    &errOut,
		Cwd:       func() (string, error) { return "/fake/project", nil },
		Resolver:  &mockResolver{err: errors.New("no version")},
		State:     &mockState{},
		Registry:  &mockRegistry{installed: map[string]bool{}, dirs: map[string]string{}, exePaths: map[string]string{}},
		Installer: &mockInstaller{},
		Shims:     &mockShims{dir: "/fake/.fvm/shims"},
		Remote:    &mockRemote{},
		Doctor:    &mockDoctor{},
	}
	return app, &out, &errOut
}

// ── meta commands ─────────────────────────────────────────────────────────────

func TestRunHelp(t *testing.T) {
	app, out, _ := makeApp()
	if err := app.Run(nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("expected help output, got: %q", out.String())
	}
}

func TestRunVersion(t *testing.T) {
	app, out, _ := makeApp()
	if err := app.Run([]string{"version"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), Version) {
		t.Fatalf("expected version in output, got: %q", out.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	app, _, errOut := makeApp()
	if err := app.Run([]string{"bogus"}); err == nil {
		t.Fatal("expected error for unknown command")
	}
	if !strings.Contains(errOut.String(), "unknown command") {
		t.Fatalf("expected 'unknown command' in stderr, got: %q", errOut.String())
	}
}

// ── current ───────────────────────────────────────────────────────────────────

func TestRunCurrentLocal(t *testing.T) {
	app, out, _ := makeApp()
	app.Resolver = &mockResolver{result: &ResolvedVersion{Version: "13.346", Source: "local"}}

	if err := app.Run([]string{"current"}); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "13.346") || !strings.Contains(got, ".fvm-version") {
		t.Fatalf("unexpected current output: %q", got)
	}
}

func TestRunCurrentGlobal(t *testing.T) {
	app, out, _ := makeApp()
	app.Resolver = &mockResolver{result: &ResolvedVersion{Version: "13.300", Source: "global"}}

	if err := app.Run([]string{"current"}); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "13.300") || !strings.Contains(got, "global") {
		t.Fatalf("unexpected current output: %q", got)
	}
}

func TestRunCurrentShell(t *testing.T) {
	app, out, _ := makeApp()
	app.Resolver = &mockResolver{result: &ResolvedVersion{Version: "13.400", Source: "shell"}}

	if err := app.Run([]string{"current"}); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "13.400") || !strings.Contains(got, "FOUNDRY_VERSION") {
		t.Fatalf("unexpected current output: %q", got)
	}
}

func TestRunCurrentNoVersion(t *testing.T) {
	app, _, _ := makeApp()
	// default resolver returns error
	if err := app.Run([]string{"current"}); err == nil {
		t.Fatal("expected error when no version configured")
	}
}

// ── local / global ────────────────────────────────────────────────────────────

func TestRunLocal(t *testing.T) {
	app, _, _ := makeApp()
	state := &mockState{}
	app.State = state

	if err := app.Run([]string{"local", "13.346"}); err != nil {
		t.Fatal(err)
	}
	if !state.writeLocalCalled {
		t.Fatal("expected WriteLocal to be called")
	}
	if state.writeLocalVersion != "13.346" {
		t.Fatalf("expected version 13.346, got %s", state.writeLocalVersion)
	}
}

func TestRunLocalMissingArg(t *testing.T) {
	app, _, _ := makeApp()
	if err := app.Run([]string{"local"}); err == nil {
		t.Fatal("expected error when version arg missing")
	}
}

func TestRunGlobal(t *testing.T) {
	app, _, _ := makeApp()
	state := &mockState{}
	app.State = state

	if err := app.Run([]string{"global", "13.346"}); err != nil {
		t.Fatal(err)
	}
	if !state.writeGlobalCalled {
		t.Fatal("expected WriteGlobal to be called")
	}
	if state.writeGlobalVersion != "13.346" {
		t.Fatalf("expected version 13.346, got %s", state.writeGlobalVersion)
	}
}

func TestRunGlobalMissingArg(t *testing.T) {
	app, _, _ := makeApp()
	if err := app.Run([]string{"global"}); err == nil {
		t.Fatal("expected error when version arg missing")
	}
}

// ── which / where ─────────────────────────────────────────────────────────────

func TestRunWhich(t *testing.T) {
	app, out, _ := makeApp()
	app.Resolver = &mockResolver{result: &ResolvedVersion{Version: "13.346", Source: "local"}}
	app.Registry = &mockRegistry{
		installed: map[string]bool{"13.346": true},
		dirs:      map[string]string{},
		exePaths:  map[string]string{"13.346": "/fake/.fvm/versions/13.346/bin/foundry"},
	}

	if err := app.Run([]string{"which"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "/fake/.fvm/versions/13.346/bin/foundry") {
		t.Fatalf("unexpected which output: %q", out.String())
	}
}

func TestRunWhichNotInstalled(t *testing.T) {
	app, _, _ := makeApp()
	app.Resolver = &mockResolver{result: &ResolvedVersion{Version: "13.346", Source: "local"}}
	// registry has no installed versions

	if err := app.Run([]string{"which"}); err == nil {
		t.Fatal("expected error when version not installed")
	}
}

func TestRegistryAdapterExecutablePath_macAppBundle(t *testing.T) {
	dir := t.TempDir()
	version := "13.346"
	root := filepath.Join(dir, ".fvm", "versions", version)
	appBinary := filepath.Join(root, "foundry", "Foundry Virtual Tabletop.app", "Contents", "MacOS", "Foundry Virtual Tabletop")
	if err := os.MkdirAll(filepath.Dir(appBinary), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(appBinary, []byte("binary"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	oldHome := os.Getenv("HOME")
	fakeHome := filepath.Join(dir)
	if err := os.Setenv("HOME", fakeHome); err != nil {
		t.Fatalf("Setenv: %v", err)
	}
	defer os.Setenv("HOME", oldHome)

	exePath, err := (&registryAdapter{}).ExecutablePath(version)
	if err != nil {
		t.Fatalf("ExecutablePath: %v", err)
	}
	if exePath != appBinary {
		t.Fatalf("expected %q, got %q", appBinary, exePath)
	}
}

func TestRunWhere(t *testing.T) {
	app, out, _ := makeApp()
	app.Registry = &mockRegistry{
		installed: map[string]bool{},
		dirs:      map[string]string{"13.346": "/fake/.fvm/versions/13.346"},
		exePaths:  map[string]string{},
	}

	if err := app.Run([]string{"where", "13.346"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "/fake/.fvm/versions/13.346") {
		t.Fatalf("unexpected where output: %q", out.String())
	}
}

func TestRunWhereMissingArg(t *testing.T) {
	app, _, _ := makeApp()
	if err := app.Run([]string{"where"}); err == nil {
		t.Fatal("expected error when version arg missing")
	}
}

// ── list ──────────────────────────────────────────────────────────────────────

func TestRunList(t *testing.T) {
	app, out, _ := makeApp()
	app.Registry = &mockRegistry{
		versions:  []string{"13.200", "13.346"},
		installed: map[string]bool{},
		dirs:      map[string]string{},
		exePaths:  map[string]string{},
	}
	app.Resolver = &mockResolver{result: &ResolvedVersion{Version: "13.346", Source: "local"}}

	if err := app.Run([]string{"list"}); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "13.200") || !strings.Contains(got, "13.346") {
		t.Fatalf("expected both versions in output, got: %q", got)
	}
	if !strings.Contains(got, "* 13.346") {
		t.Fatalf("expected active version to be marked with *, got: %q", got)
	}
}

func TestRunListEmpty(t *testing.T) {
	app, out, _ := makeApp()
	// registry returns no versions by default

	if err := app.Run([]string{"list"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "no versions installed") {
		t.Fatalf("unexpected list output: %q", out.String())
	}
}

func TestRunListRemote(t *testing.T) {
	app, out, _ := makeApp()
	app.Remote = &mockRemote{versions: []string{"13.300", "13.346"}}

	if err := app.Run([]string{"list-remote"}); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "13.300") || !strings.Contains(got, "13.346") {
		t.Fatalf("unexpected list-remote output: %q", got)
	}
}

func TestRunListRemoteError(t *testing.T) {
	app, _, _ := makeApp()
	app.Remote = &mockRemote{err: errors.New("network unavailable")}

	if err := app.Run([]string{"list-remote"}); err == nil {
		t.Fatal("expected error when remote provider fails")
	}
}

// ── install ───────────────────────────────────────────────────────────────────

func TestRunInstall(t *testing.T) {
	app, out, _ := makeApp()
	app.Installer = &mockInstaller{err: nil}

	if err := app.Run([]string{"install", "13.346"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "installed successfully") {
		t.Fatalf("unexpected install output: %q", out.String())
	}
}

func TestRunInstallAlreadyInstalled(t *testing.T) {
	app, out, _ := makeApp()
	app.Registry = &mockRegistry{
		installed: map[string]bool{"13.346": true},
		dirs:      map[string]string{},
		exePaths:  map[string]string{},
	}

	if err := app.Run([]string{"install", "13.346"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "already installed") {
		t.Fatalf("unexpected install output: %q", out.String())
	}
}

func TestRunInstallMissingArg(t *testing.T) {
	app, _, _ := makeApp()
	if err := app.Run([]string{"install"}); err == nil {
		t.Fatal("expected error when version arg missing")
	}
}

// ── shim ──────────────────────────────────────────────────────────────────────

func TestRunShimRegenerate(t *testing.T) {
	app, out, _ := makeApp()
	app.Shims = &mockShims{dir: "/fake/.fvm/shims", err: nil}

	if err := app.Run([]string{"shim", "regenerate"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "regenerated") {
		t.Fatalf("unexpected shim output: %q", out.String())
	}
}

func TestRunShimRegenerateError(t *testing.T) {
	app, _, _ := makeApp()
	app.Shims = &mockShims{dir: "/fake/.fvm/shims", err: errors.New("write error")}

	if err := app.Run([]string{"shim", "regenerate"}); err == nil {
		t.Fatal("expected error when regenerate fails")
	}
}

func TestRunShimNoSubcommand(t *testing.T) {
	app, _, _ := makeApp()
	if err := app.Run([]string{"shim"}); err == nil {
		t.Fatal("expected error when no subcommand given")
	}
}

func TestRunShimUnknownSubcommand(t *testing.T) {
	app, _, _ := makeApp()
	if err := app.Run([]string{"shim", "frobnicate"}); err == nil {
		t.Fatal("expected error for unknown shim subcommand")
	}
}

// ── doctor ────────────────────────────────────────────────────────────────────

func TestRunDoctorAllOK(t *testing.T) {
	app, out, _ := makeApp()
	app.Doctor = &mockDoctor{results: []DoctorResult{
		{Name: "fvm root", OK: true, Message: "/fake/.fvm"},
		{Name: "shims on PATH", OK: true, Message: "/fake/.fvm/shims"},
	}}

	if err := app.Run([]string{"doctor"}); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "looks good") {
		t.Fatalf("expected 'looks good' in doctor output, got: %q", got)
	}
}

func TestRunDoctorHasIssues(t *testing.T) {
	app, out, _ := makeApp()
	app.Doctor = &mockDoctor{results: []DoctorResult{
		{Name: "fvm root", OK: true, Message: "/fake/.fvm"},
		{Name: "shims on PATH", OK: false, Message: "not on PATH"},
	}}

	if err := app.Run([]string{"doctor"}); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "[!!]") {
		t.Fatalf("expected failure marker in doctor output, got: %q", got)
	}
	if !strings.Contains(got, "has issues") {
		t.Fatalf("expected 'has issues' in doctor output, got: %q", got)
	}
}

// ── init ──────────────────────────────────────────────────────────────────────

func TestRunInitBash(t *testing.T) {
	app, out, _ := makeApp()
	if err := app.Run([]string{"init", "bash"}); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "FVM_DIR") || !strings.Contains(got, "PATH") {
		t.Fatalf("unexpected bash init output: %q", got)
	}
}

func TestRunInitZsh(t *testing.T) {
	app, out, _ := makeApp()
	if err := app.Run([]string{"init", "zsh"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "FVM_DIR") {
		t.Fatalf("unexpected zsh init output: %q", out.String())
	}
}

func TestRunInitFish(t *testing.T) {
	app, out, _ := makeApp()
	if err := app.Run([]string{"init", "fish"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "FVM_DIR") {
		t.Fatalf("unexpected fish init output: %q", out.String())
	}
}

func TestRunInitCaseInsensitive(t *testing.T) {
	app, out, _ := makeApp()
	if err := app.Run([]string{"init", "BASH"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "FVM_DIR") {
		t.Fatalf("expected bash init for uppercase BASH: %q", out.String())
	}
}

func TestRunInitUnknownShell(t *testing.T) {
	app, _, _ := makeApp()
	if err := app.Run([]string{"init", "powershell"}); err == nil {
		t.Fatal("expected error for unsupported shell")
	}
}

func TestRunInitMissingArg(t *testing.T) {
	app, _, _ := makeApp()
	if err := app.Run([]string{"init"}); err == nil {
		t.Fatal("expected error when shell arg missing")
	}
}
