package resolve

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/foundry/fvm/internal/paths"
)

// Source indicates where a resolved version was found.
type Source int

const (
	SourceShellOverride Source = iota
	SourceLocal
	SourceGlobal
)

func (s Source) String() string {
	switch s {
	case SourceShellOverride:
		return "shell override"
	case SourceLocal:
		return "local (.fvm-version)"
	case SourceGlobal:
		return "global (~/.fvm/version)"
	default:
		return "unknown"
	}
}

// Result is the outcome of a successful version resolution.
type Result struct {
	Version string
	Source  Source
	File    string // path of the file that supplied the version, if any
}

// ErrNoVersion is returned when no version can be resolved via any source.
var ErrNoVersion = errors.New("no Foundry version configured: set a version with 'fvm local <version>' or 'fvm global <version>'")

var errNoLocalVersion = errors.New("no .fvm-version found")

// Inputs holds all context needed to resolve a version.
// Fields can be overridden in tests without touching the real filesystem.
type Inputs struct {
	// ShellOverride is the value of the FVM_VERSION environment variable.
	ShellOverride string
	// Cwd is the directory to begin the nearest-.fvm-version walk from.
	Cwd string
	// GlobalVersionPath is the path to the global version file (e.g. ~/.fvm/version).
	GlobalVersionPath string
}

// DefaultInputs returns Inputs populated from the real environment.
func DefaultInputs() (Inputs, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Inputs{}, fmt.Errorf("resolve: cannot determine cwd: %w", err)
	}
	return Inputs{
		ShellOverride:     os.Getenv("FVM_VERSION"),
		Cwd:               cwd,
		GlobalVersionPath: paths.GlobalVersionPath(),
	}, nil
}

// Resolve applies the locked resolution order:
//  1. shell override (FVM_VERSION env var)
//  2. nearest .fvm-version walking up from Cwd
//  3. global ~/.fvm/version
//  4. ErrNoVersion
func Resolve(in Inputs) (Result, error) {
	if v := strings.TrimSpace(in.ShellOverride); v != "" {
		return Result{Version: v, Source: SourceShellOverride}, nil
	}

	if in.Cwd != "" {
		if v, file, err := findLocal(in.Cwd); err == nil {
			return Result{Version: v, Source: SourceLocal, File: file}, nil
		} else if !errors.Is(err, errNoLocalVersion) {
			return Result{}, err
		}
	}

	if in.GlobalVersionPath != "" {
		if v, err := readVersionFile(in.GlobalVersionPath); err == nil {
			return Result{Version: v, Source: SourceGlobal, File: in.GlobalVersionPath}, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return Result{}, err
		}
	}

	return Result{}, ErrNoVersion
}

// findLocal walks cwd upward looking for the nearest .fvm-version file.
func findLocal(cwd string) (version, file string, err error) {
	dir := cwd
	for {
		candidate := filepath.Join(dir, paths.LocalVersion)
		if _, statErr := os.Stat(candidate); statErr == nil {
			v, readErr := readVersionFile(candidate)
			if readErr != nil {
				return "", "", readErr
			}
			return v, candidate, nil
		} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return "", "", statErr
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", "", errNoLocalVersion
}

// readVersionFile reads and trims a version string from a single-line file.
func readVersionFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	v := strings.TrimSpace(string(data))
	if v == "" {
		return "", errors.New("empty version file")
	}
	return v, nil
}
