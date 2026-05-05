package cli

import "fmt"

func (a *App) runDoctor(_ []string) error {
	results := a.Doctor.Check()

	allOK := true
	for _, r := range results {
		mark := "[ok]"
		if !r.OK {
			mark = "[!!]"
			allOK = false
		}
		fmt.Fprintf(a.Stdout, "  %s  %s: %s\n", mark, r.Name, r.Message)
	}

	fmt.Fprintln(a.Stdout, "")
	if allOK {
		fmt.Fprintln(a.Stdout, "fvm environment looks good")
	} else {
		fmt.Fprintln(a.Stdout, "fvm environment has issues — see above")
	}

	// Add data path check
	cfg, err := config.LoadDefault()
	if err != nil {
		fmt.Fprintf(a.Stdout, "  [!!]  data-path: failed to load config\n")
	} else {
		dataRoot := cfg.DataPath
		if dataRoot == "" {
			dataRoot = filepath.Join(paths.Root(), "data")
		}
		fmt.Fprintf(a.Stdout, "  [ok]  data-path: %s\n", dataRoot)
	}

	return nil
}
