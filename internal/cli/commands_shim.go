package cli

import "fmt"

func (a *App) runShim(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: fvm shim <subcommand>\n\nSubcommands:\n  regenerate   Rebuild shim binaries in ~/.fvm/shims")
	}

	switch args[0] {
	case "regenerate":
		fmt.Fprintln(a.Stdout, "regenerating shims...")
		if err := a.Shims.Regenerate(); err != nil {
			return fmt.Errorf("shim regeneration failed: %w", err)
		}
		fmt.Fprintf(a.Stdout, "shims regenerated in %s\n", a.Shims.ShimDir())
		return nil
	default:
		return fmt.Errorf("unknown shim subcommand: %s\n\nRun 'fvm shim' for usage", args[0])
	}
}
