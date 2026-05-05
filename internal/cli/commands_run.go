package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/foundry/fvm/internal/config"
	"github.com/foundry/fvm/internal/paths"
)

func (a *App) runRun(args []string) error {
	var version string
	var passthrough []string

	if len(args) > 0 {
		if args[0] == "--" {
			passthrough = args[1:]
		} else {
			version = args[0]
			if len(args) > 1 {
				if args[1] == "--" {
					passthrough = args[2:]
				} else {
					passthrough = args[1:]
				}
			}
		}
	}

	if version == "" {
		cwd, err := a.cwd()
		if err != nil {
			return err
		}
		resolved, err := a.Resolver.Current(cwd)
		if err != nil {
			return err
		}
		version = resolved.Version
	}

	exePath, err := a.Registry.ExecutablePath(version)
	if err != nil {
		return err
	}

	cfg, err := config.LoadDefault()
	if err != nil {
		return err
	}
	dataRoot := cfg.DataPath
	if dataRoot == "" {
		dataRoot = filepath.Join(paths.Root(), "data")
	}
	dataPath := filepath.Join(dataRoot, version)
	if err := os.MkdirAll(dataPath, 0o755); err != nil {
		return err
	}

	cmdArgs := append([]string{"--dataPath", dataPath}, passthrough...)
	cmd := exec.Command(exePath, cmdArgs...)
	cmd.Stdout = a.Stdout
	cmd.Stderr = a.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run foundry %s: %w", version, err)
	}
	return nil
}
