package cli

import "fmt"

func (a *App) runWhich(args []string) error {
	cwd, err := a.cwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}

	rv, err := a.Resolver.Current(cwd)
	if err != nil {
		return err
	}

	if !a.Registry.IsInstalled(rv.Version) {
		return fmt.Errorf("version %s is not installed — run: fvm install %s", rv.Version, rv.Version)
	}

	exePath, err := a.Registry.ExecutablePath(rv.Version)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		switch args[0] {
		case "foundry", "foundryvtt":
			// Both supported shim names currently resolve to the Foundry executable.
		default:
			return fmt.Errorf("unknown executable: %s", args[0])
		}
	}

	fmt.Fprintln(a.Stdout, exePath)
	return nil
}
