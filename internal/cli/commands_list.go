package cli

import "fmt"

func (a *App) runList(_ []string) error {
	versions, err := a.Registry.InstalledVersions()
	if err != nil {
		return fmt.Errorf("cannot list installed versions: %w", err)
	}

	if len(versions) == 0 {
		fmt.Fprintln(a.Stdout, "no versions installed — run: fvm install <version>")
		return nil
	}

	// Mark the active version if resolvable.
	var active string
	if cwd, err := a.cwd(); err == nil {
		if rv, err := a.Resolver.Current(cwd); err == nil {
			active = rv.Version
		}
	}

	for _, v := range versions {
		if v == active {
			fmt.Fprintf(a.Stdout, "* %s\n", v)
		} else {
			fmt.Fprintf(a.Stdout, "  %s\n", v)
		}
	}
	return nil
}

func (a *App) runListRemote(_ []string) error {
	versions, err := a.Remote.RemoteVersions()
	if err != nil {
		return fmt.Errorf("cannot list remote versions: %w", err)
	}

	if len(versions) == 0 {
		fmt.Fprintln(a.Stdout, "no remote versions found")
		return nil
	}

	for _, v := range versions {
		fmt.Fprintln(a.Stdout, v)
	}
	return nil
}
