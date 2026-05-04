package app

import (
	"fmt"
	"io"
)

const Version = "0.1.0"

func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		printHelp(stdout)
		return nil
	}

	switch args[0] {
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "fvm %s\n", Version)
		return nil
	case "help", "--help", "-h":
		printHelp(stdout)
		return nil
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		printHelp(stderr)
		return fmt.Errorf("usage error")
	}
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "fvm - Foundry version manager")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  fvm <command>")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  version   Print the current fvm version")
	fmt.Fprintln(w, "  help      Show this help message")
}
