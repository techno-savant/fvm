package cli

import "fmt"

func (a *App) runGlobal(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: fvm global <version>")
	}
	version := args[0]

	if err := a.State.WriteGlobal(version); err != nil {
		return fmt.Errorf("failed to write global version: %w", err)
	}

	fmt.Fprintf(a.Stdout, "set global Foundry version to %s\n", version)
	return nil
}
