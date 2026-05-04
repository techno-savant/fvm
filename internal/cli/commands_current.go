package cli

import "fmt"

func (a *App) runCurrent(_ []string) error {
	cwd, err := a.cwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}

	rv, err := a.Resolver.Current(cwd)
	if err != nil {
		return err
	}

	switch rv.Source {
	case "shell":
		fmt.Fprintf(a.Stdout, "%s (FOUNDRY_VERSION)\n", rv.Version)
	case "local":
		fmt.Fprintf(a.Stdout, "%s (.fvm-version)\n", rv.Version)
	case "global":
		fmt.Fprintf(a.Stdout, "%s (global)\n", rv.Version)
	default:
		fmt.Fprintf(a.Stdout, "%s\n", rv.Version)
	}
	return nil
}
