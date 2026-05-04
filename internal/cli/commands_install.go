package cli

import "fmt"

func (a *App) runInstall(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: fvm install <version>")
	}
	version := args[0]

	if a.Registry.IsInstalled(version) {
		fmt.Fprintf(a.Stdout, "Foundry %s is already installed\n", version)
		return nil
	}

	fmt.Fprintf(a.Stdout, "installing Foundry %s...\n", version)

	if err := a.Installer.Install(version); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	fmt.Fprintf(a.Stdout, "Foundry %s installed successfully\n", version)
	return nil
}
