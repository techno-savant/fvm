package main

import (
	"fmt"
	"os"

	"github.com/foundry/fvm/internal/cli"
)

func main() {
	app, err := cli.New(os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
