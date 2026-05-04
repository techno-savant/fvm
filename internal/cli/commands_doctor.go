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
	return nil
}
