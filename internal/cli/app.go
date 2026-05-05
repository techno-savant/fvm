package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/foundry/fvm/internal/config"
	"github.com/foundry/fvm/internal/paths"
	"gopkg.in/yaml.v3"
)

// override it via -ldflags.
var Version = "0.1.0"

// App is the fvm CLI application. All service fields are interfaces so tests
// can inject mocks and Stream B can swap in real implementations later.
type App struct {
	Stdout   io.Writer
	Stderr   io.Writer
	Cwd      func() (string, error) // nil defaults to os.Getwd
	Resolver Resolver
	State    VersionState
	Registry Registry
	Installer Installer
	Shims    ShimManager
	Remote   RemoteProvider
	Doctor   DoctorChecker
}

// Run dispatches to the command handler identified by args[0].
func (a *App) Run(args []string) error {
	if len(args) == 0 {
		a.printHelp(a.Stdout)
		return nil
	}

	switch args[0] {
	case "config":
		return a.runConfig(args[1:])
	case "current":
		return a.runCurrent(args[1:])
	case "which":
		return a.runWhich(args[1:])
	case "where":
		return a.runWhere(args[1:])
	case "local":
		return a.runLocal(args[1:])
	case "global":
		return a.runGlobal(args[1:])
	case "list":
		return a.runList(args[1:])
	case "list-remote":
		return a.runListRemote(args[1:])
	case "install":
		return a.runInstall(args[1:])
	case "run":
		return a.runRun(args[1:])
	case "shim":
		return a.runShim(args[1:])
	case "doctor":
		return a.runDoctor(args[1:])
	case "init":
		return a.runInit(args[1:])
	case "version", "--version", "-v":
		fmt.Fprintf(a.Stdout, "fvm %s\n", Version)
		return nil
	case "help", "--help", "-h":
		a.printHelp(a.Stdout)
		return nil
	default:
		fmt.Fprintf(a.Stderr, "unknown command: %s\n\n", args[0])
		a.printHelp(a.Stderr)
		return fmt.Errorf("usage error")
	}
}

// cwd returns the current working directory, using a.Cwd if set.
func (a *App) cwd() (string, error) {
	if a.Cwd != nil {
		return a.Cwd()
	}
	return os.Getwd()
}

func (a *App) printHelp(w io.Writer) {
	fmt.Fprintf(w, "fvm %s — Foundry version manager\n", Version)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  fvm <command> [args]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Version selection:")
	fmt.Fprintln(w, "  current            Show the active Foundry version and its source")
	fmt.Fprintln(w, "  which              Print the path to the active foundry executable")
	fmt.Fprintln(w, "  where <version>    Show the install directory for a version")
	fmt.Fprintln(w, "  local <version>    Pin a version for this project (.fvm-version)")
	fmt.Fprintln(w, "  global <version>   Set the user-global default (~/.fvm/version)")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Installation:")
	fmt.Fprintln(w, "  list               List installed versions")
	fmt.Fprintln(w, "  list-remote        List versions available for download")
	fmt.Fprintln(w, "  install <version>  Download and install a version")
	fmt.Fprintln(w, "  run [version] [-- args...]  Launch Foundry with version-specific data path")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Maintenance:")
	fmt.Fprintln(w, "  config set <key> <value>  Update fvm config values")
	fmt.Fprintln(w, "  shim regenerate    Rebuild shim binaries in ~/.fvm/shims")
	fmt.Fprintln(w, "  doctor             Check your fvm environment")
	fmt.Fprintln(w, "  init <shell>       Print shell integration (bash, zsh, fish)")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Meta:")
	fmt.Fprintln(w, "  version            Print fvm version")
	fmt.Fprintln(w, "  help               Show this message")
}

func (a *App) runConfig(args []string) error {
	if len(args) == 0 {
		fmt.Fprintf(a.Stderr, "config command requires a subcommand\n")
		return fmt.Errorf("usage error")
	}

	switch args[0] {
	case "set":
		return a.runConfigSet(args[1:])
	default:
		fmt.Fprintf(a.Stderr, "unknown config subcommand: %s\n", args[0])
		return fmt.Errorf("usage error")
	}
}

func (a *App) runConfigSet(args []string) error {
	if len(args) < 2 {
		fmt.Fprintf(a.Stderr, "config set requires a key and value\n")
		return fmt.Errorf("usage error")
	}

	key := args[0]
	value := strings.Join(args[1:], " ")

	cfg, err := config.LoadDefault()
	if err != nil {
		return err
	}

	switch key {
	case "data-path":
		cfg.DataPath = value
	default:
		fmt.Fprintf(a.Stderr, "unknown config key: %s\n", key)
		return fmt.Errorf("usage error")
	}

	configPath := paths.ConfigPath()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return err
	}

	fmt.Fprintf(a.Stdout, "Config updated: %s=%s\n", key, value)
	return nil
}
