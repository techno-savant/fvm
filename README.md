# fvm

`fvm` is a Foundry version manager.

It helps you install multiple Foundry Virtual Tabletop versions on one machine and switch between them per project, per user, or for just the current shell session.

If you have used tools like `asdf`, `nvm`, or `pyenv`, this is the same basic idea for Foundry.

## Who this is for

`fvm` is meant for Foundry developers, module authors, system authors, and hobbyists who need to test against more than one Foundry release without manually renaming folders or juggling symlinks.

## What `fvm` does

With `fvm`, you can:

- install multiple Foundry versions side-by-side
- set a default version for your machine
- pin a version for one project with `.fvm-version`
- temporarily override the version for your current shell
- ask which version is active and where it lives on disk

## How version selection works

When `fvm` needs to decide which Foundry version to use, it checks in this order:

1. `FOUNDRY_VERSION` in your current shell
2. the nearest `.fvm-version` file in your current project
3. your global default in `~/.fvm/version`

If none of those exist, `fvm` will tell you no version is configured.

## Quick start

This is the shortest path from zero to working.

### 1. Install `fvm`

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  | sh
```

This installs the `fvm` binary itself. It does not install a Foundry version yet.

If you want more detail, custom install directories, or manual installation steps, see [docs/install.md](docs/install.md).

### 2. Make sure `fvm` works

```sh
fvm version
fvm help
```

You should see the CLI version and the available commands.

### 3. Enable shell integration

Pick your shell:

For bash:

```sh
eval "$(fvm init bash)"
```

For zsh:

```sh
eval "$(fvm init zsh)"
```

For fish:

```sh
fvm init fish | source
```

That only affects your current shell session. To make it permanent, add the same command to your shell config file:

- bash: `~/.bashrc` or `~/.bash_profile`
- zsh: `~/.zshrc`
- fish: `~/.config/fish/config.fish`

This step puts `~/.fvm/shims` on your `PATH`, which is how `fvm` intercepts `foundry` and routes it to the selected version.

### 4. Install a Foundry version

```sh
fvm list-remote
fvm install 13.346
```

Use `fvm list-remote` to see what versions are available, then install the one you want.

### 5. Set a default version

```sh
fvm global 13.346
```

This writes your default version to `~/.fvm/version`.

### 6. Verify what is active

```sh
fvm current
fvm which
```

Typical output might look like:

```text
13.346 (global)
/Users/you/.fvm/versions/13.346/foundry/foundry
```

## Common workflows

### Use one version for all projects by default

```sh
fvm install 13.346
fvm global 13.346
```

### Pin a specific project to a version

From inside the project directory:

```sh
fvm install 12.331
fvm local 12.331
```

This creates a `.fvm-version` file in the current directory.

Anyone opening that project later can run `fvm current` and see what it expects.

### Temporarily switch just for this terminal

After enabling shell integration, use:

```sh
fvm_use 12.331
```

That sets `FOUNDRY_VERSION` for the current shell only. Close the shell and the override is gone.

Use this when you want to test quickly without changing project files or your global default.

### Check where a version is installed

```sh
fvm where 13.346
```

### Check whether your environment is healthy

```sh
fvm doctor
```

Use this first if something feels off.

## Command reference

### `fvm current`

Shows the active version and where it came from.

Examples:

```sh
fvm current
```

Possible output:

```text
13.346 (.fvm-version)
13.346 (global)
13.346 (FOUNDRY_VERSION)
```

### `fvm which`

Prints the path to the active Foundry executable.

```sh
fvm which
```

If the selected version is not installed yet, `fvm` tells you which install command to run.

### `fvm where <version>`

Shows the install directory for a specific version.

```sh
fvm where 13.346
```

### `fvm local <version>`

Pins the current project to a version by writing `.fvm-version`.

```sh
fvm local 13.346
```

### `fvm global <version>`

Sets your machine-wide default version.

```sh
fvm global 13.346
```

### `fvm list`

Lists versions currently installed on your machine.

```sh
fvm list
```

### `fvm list-remote`

Lists versions available for download.

```sh
fvm list-remote
```

### `fvm install <version>`

Downloads and installs a Foundry version.

```sh
fvm install 13.346
```

### `fvm shim regenerate`

Rebuilds the shim executables in `~/.fvm/shims`.

```sh
fvm shim regenerate
```

Run this if your shell integration is set up but the shim directory seems stale.

### `fvm doctor`

Runs basic environment checks.

```sh
fvm doctor
```

### `fvm init <shell>`

Prints shell integration code for `bash`, `zsh`, or `fish`.

```sh
fvm init bash
fvm init zsh
fvm init fish
```

Use this to inspect what `fvm` wants you to add to your shell config.

## Files and directories `fvm` uses

By default, `fvm` stores data under `~/.fvm`.

Important paths:

- `~/.fvm/version` — your global default version
- `~/.fvm/shims/` — shim executables placed on your `PATH`
- `~/.fvm/versions/` — installed Foundry versions
- `.fvm-version` — per-project version file

## Troubleshooting

### `fvm: command not found`

The `fvm` binary is installed somewhere not on your `PATH`.

Read [docs/install.md](docs/install.md) and add the install directory to your shell profile.

### `foundry` does not switch versions

Usually one of these is true:

- you have not run `fvm init <shell>` yet
- you added it to your shell config but have not reloaded the shell
- `~/.fvm/shims` is not on your `PATH`
- the version you selected is not installed yet

Start here:

```sh
fvm doctor
fvm current
fvm which
```

### `fvm current` says no version is configured

That means none of the three selectors are set:

- no `FOUNDRY_VERSION` in your shell
- no `.fvm-version` in this project or a parent directory
- no global `~/.fvm/version`

Fix it with one of:

```sh
fvm global 13.346
fvm local 13.346
fvm_use 13.346
```

### I installed `fvm`, but I still cannot run `foundry`

Installing `fvm` only installs the version manager.

You still need to:

1. enable shell integration
2. install a Foundry version
3. choose a version

Use the Quick start above in order.

## Development

Common commands:

```sh
go fmt ./...
go test -race ./...
go build ./cmd/fvm
```

Syntax-check the installer script:

```sh
bash -n scripts/install.sh
```

## CI

GitHub Actions runs formatting, tests, and binary builds on pushes and pull requests.

A separate release workflow runs on `v*` tags and publishes cross-compiled archives plus a `checksums.sha256` file to GitHub Releases.

## Releasing

Push a version tag to trigger the release workflow:

```sh
git tag v0.2.0
git push origin v0.2.0
```

## Supported platforms

| OS    | Architecture |
|-------|--------------|
| Linux | amd64        |
| Linux | arm64        |
| macOS | amd64        |
| macOS | arm64        |

## License

See [LICENSE](LICENSE) if present.
