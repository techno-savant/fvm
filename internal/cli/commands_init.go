package cli

import (
	"fmt"
	"strings"
)

func (a *App) runInit(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: fvm init <shell>\n\nSupported shells: bash, zsh, fish")
	}

	switch strings.ToLower(args[0]) {
	case "bash":
		fmt.Fprint(a.Stdout, bashInit)
	case "zsh":
		fmt.Fprint(a.Stdout, zshInit)
	case "fish":
		fmt.Fprint(a.Stdout, fishInit)
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", args[0])
	}
	return nil
}

const bashInit = `# fvm shell integration for bash
# Add this to your ~/.bashrc or ~/.bash_profile:
#   eval "$(fvm init bash)"

export FVM_DIR="${FVM_DIR:-$HOME/.fvm}"
export PATH="$FVM_DIR/shims:$PATH"

# fvm_use activates a version for the current shell session only (ephemeral).
fvm_use() {
    export FOUNDRY_VERSION="$1"
    echo "Using Foundry $FOUNDRY_VERSION (session only — not persisted)"
}
`

const zshInit = `# fvm shell integration for zsh
# Add this to your ~/.zshrc:
#   eval "$(fvm init zsh)"

export FVM_DIR="${FVM_DIR:-$HOME/.fvm}"
export PATH="$FVM_DIR/shims:$PATH"

# fvm_use activates a version for the current shell session only (ephemeral).
fvm_use() {
    export FOUNDRY_VERSION="$1"
    echo "Using Foundry $FOUNDRY_VERSION (session only — not persisted)"
}
`

const fishInit = `# fvm shell integration for fish
# Add this to your ~/.config/fish/config.fish:
#   fvm init fish | source

set -gx FVM_DIR "$HOME/.fvm"
fish_add_path "$FVM_DIR/shims"

# fvm_use activates a version for the current shell session only (ephemeral).
function fvm_use
    set -gx FOUNDRY_VERSION $argv[1]
    echo "Using Foundry $FOUNDRY_VERSION (session only — not persisted)"
end
`
