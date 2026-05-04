package cli

import "fmt"

func (a *App) runWhere(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: fvm where <version>")
	}
	fmt.Fprintln(a.Stdout, a.Registry.VersionDir(args[0]))
	return nil
}
