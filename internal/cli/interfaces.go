package cli

// ResolvedVersion is the result of version resolution.
type ResolvedVersion struct {
	Version string // e.g. "13.346"
	Source  string // "shell", "local", "global"
	File    string // path to the config file; empty for "shell"
}

// Resolver resolves which Foundry version is currently active.
type Resolver interface {
	Current(cwd string) (*ResolvedVersion, error)
}

// VersionState reads and writes persistent version selections.
type VersionState interface {
	ReadLocal(cwd string) (string, error)
	WriteLocal(cwd, version string) error
	ReadGlobal() (string, error)
	WriteGlobal(version string) error
}

// Registry queries installed Foundry versions.
type Registry interface {
	InstalledVersions() ([]string, error)
	IsInstalled(version string) bool
	VersionDir(version string) string
	ExecutablePath(version string) (string, error)
}

// Installer installs a Foundry version.
type Installer interface {
	Install(version string) error
}

// ShimManager manages shim binaries.
type ShimManager interface {
	ShimDir() string
	Regenerate() error
}

// RemoteProvider lists versions available for download.
type RemoteProvider interface {
	RemoteVersions() ([]string, error)
}

// DoctorResult is a single health-check result.
type DoctorResult struct {
	Name    string
	OK      bool
	Message string
}

// DoctorChecker performs environment health checks.
type DoctorChecker interface {
	Check() []DoctorResult
}
