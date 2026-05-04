package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/foundry/fvm/internal/config"
	"github.com/foundry/fvm/internal/install"
	"github.com/foundry/fvm/internal/paths"
	"github.com/foundry/fvm/internal/releases"
	"github.com/foundry/fvm/internal/resolve"
	"github.com/foundry/fvm/internal/shim"
	"github.com/foundry/fvm/internal/state"
)

type resolverAdapter struct{}

func (r *resolverAdapter) Current(cwd string) (*ResolvedVersion, error) {
	result, err := resolve.Resolve(resolve.Inputs{
		ShellOverride:     shellOverride(),
		Cwd:               cwd,
		GlobalVersionPath: paths.GlobalVersionPath(),
	})
	if err != nil {
		if err == resolve.ErrNoVersion {
			return nil, fmt.Errorf(
				"no Foundry version configured\n" +
					"  Set FVM_VERSION, create .fvm-version, or run: fvm global <version>",
			)
		}
		return nil, err
	}

	return &ResolvedVersion{
		Version: result.Version,
		Source:  adaptSource(result.Source),
		File:    result.File,
	}, nil
}

type stateAdapter struct{}

func (s *stateAdapter) ReadLocal(cwd string) (string, error) {
	return state.ReadLocalVersion(cwd)
}

func (s *stateAdapter) WriteLocal(cwd, version string) error {
	return state.WriteLocalVersion(cwd, version)
}

func (s *stateAdapter) ReadGlobal() (string, error) {
	return state.ReadGlobalVersion(paths.GlobalVersionPath())
}

func (s *stateAdapter) WriteGlobal(version string) error {
	return state.WriteGlobalVersion(paths.GlobalVersionPath(), version)
}

type registryAdapter struct{}

func (r *registryAdapter) InstalledVersions() ([]string, error) {
	installed, err := releases.ListInstalled(paths.VersionsPath())
	if err != nil {
		return nil, err
	}
	versions := make([]string, 0, len(installed))
	for _, item := range installed {
		versions = append(versions, item.Version)
	}
	return versions, nil
}

func (r *registryAdapter) IsInstalled(version string) bool {
	layout := install.NewLayout(paths.VersionDir(version))
	return layout.IsComplete()
}

func (r *registryAdapter) VersionDir(version string) string {
	return paths.VersionDir(version)
}

func (r *registryAdapter) ExecutablePath(version string) (string, error) {
	manifest, err := install.ReadManifest(paths.VersionDir(version))
	if err == nil && manifest.ExecutablePath != "" {
		return manifest.ExecutablePath, nil
	}

	layout := install.NewLayout(paths.VersionDir(version))
	for _, candidate := range []string{
		filepath.Join(layout.BinPath(), "foundry"),
		filepath.Join(layout.FoundryPath(), "foundry"),
		filepath.Join(layout.Root, "foundry"),
	} {
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("foundry executable not found in %s", layout.Root)
}

type installerAdapter struct {
	provider releases.Provider
}

func (i *installerAdapter) Install(version string) error {
	if i.provider == nil {
		return fmt.Errorf("install: no release provider configured")
	}
	if !releases.IsVersionString(version) {
		return fmt.Errorf("invalid version: %s", version)
	}

	record, err := releases.NewInstallRecord(version, paths.VersionsPath(), paths.TmpPath())
	if err != nil {
		return err
	}
	defer func() { _ = record.Cleanup() }()

	if err := os.MkdirAll(record.Layout.Root, 0o755); err != nil {
		return err
	}
	if err := i.provider.Install(version, record.Layout.Root); err != nil {
		return err
	}

	exePath, err := (&registryAdapter{}).ExecutablePath(version)
	if err != nil {
		exePath = ""
	}

	manifest := install.Manifest{
		Version:        version,
		Platform:       runtime.GOOS,
		Arch:           runtime.GOARCH,
		ExecutablePath: exePath,
	}
	if err := record.Finalize(manifest); err != nil {
		return err
	}

	cfg, _ := config.LoadDefault()
	if cfg.AutoRegenerateShims {
		manager := &shimAdapter{shimNames: cfg.ShimNames}
		if err := manager.Regenerate(); err != nil {
			return err
		}
	}

	return nil
}

type shimAdapter struct {
	shimNames []string
}

func (s *shimAdapter) ShimDir() string { return paths.ShimsPath() }

func (s *shimAdapter) Regenerate() error {
	names := s.shimNames
	if len(names) == 0 {
		names = shim.DefaultNames()
	}
	return shim.Generate(paths.ShimsPath(), names)
}

type remoteAdapter struct {
	provider releases.Provider
}

func (r *remoteAdapter) RemoteVersions() ([]string, error) {
	if r.provider == nil {
		return nil, fmt.Errorf("remote provider not configured")
	}
	return r.provider.ListRemote()
}

type defaultDoctorChecker struct {
	root     string
	registry Registry
}

func (d *defaultDoctorChecker) Check() []DoctorResult {
	var results []DoctorResult

	_, statErr := os.Stat(d.root)
	results = append(results, DoctorResult{
		Name: "fvm root",
		OK:   statErr == nil,
		Message: func() string {
			if statErr == nil {
				return d.root
			}
			return fmt.Sprintf("missing — run: mkdir -p %s", d.root)
		}(),
	})

	shimDir := paths.ShimsPath()
	onPath := false
	for _, p := range filepath.SplitList(os.Getenv("PATH")) {
		if p == shimDir {
			onPath = true
			break
		}
	}
	results = append(results, DoctorResult{
		Name: "shims on PATH",
		OK:   onPath,
		Message: func() string {
			if onPath {
				return shimDir
			}
			return fmt.Sprintf("add %s to PATH or run: eval \"$(fvm init <shell>)\"", shimDir)
		}(),
	})

	versions, _ := d.registry.InstalledVersions()
	results = append(results, DoctorResult{
		Name: "installed versions",
		OK:   len(versions) > 0,
		Message: func() string {
			if len(versions) > 0 {
				return fmt.Sprintf("%d installed", len(versions))
			}
			return "none — run: fvm install <version>"
		}(),
	})

	return results
}

func New(stdout, stderr io.Writer) (*App, error) {
	cfg, err := config.LoadDefault()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	provider := &releases.StubProvider{}
	registry := &registryAdapter{}
	shims := &shimAdapter{shimNames: cfg.ShimNames}

	return &App{
		Stdout:    stdout,
		Stderr:    stderr,
		Resolver:  &resolverAdapter{},
		State:     &stateAdapter{},
		Registry:  registry,
		Installer: &installerAdapter{provider: provider},
		Shims:     shims,
		Remote:    &remoteAdapter{provider: provider},
		Doctor:    &defaultDoctorChecker{root: paths.Root(), registry: registry},
	}, nil
}

func adaptSource(source resolve.Source) string {
	switch source {
	case resolve.SourceShellOverride:
		return "shell"
	case resolve.SourceLocal:
		return "local"
	case resolve.SourceGlobal:
		return "global"
	default:
		return "unknown"
	}
}

func shellOverride() string {
	if v := os.Getenv("FVM_VERSION"); v != "" {
		return v
	}
	return os.Getenv("FOUNDRY_VERSION")
}
