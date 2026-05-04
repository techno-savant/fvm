package cli

import "fmt"

func (a *App) runLocal(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: fvm local <version>")
	}
	version := args[0]

	cwd, err := a.cwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}

	if err := a.State.WriteLocal(cwd, version); err != nil {
		return fmt.Errorf("failed to write .fvm-version: %w", err)
	}

	fmt.Fprintf(a.Stdout, "pinned %s in %s/.fvm-version\n", version, cwd)
	return nil
}
