# Installing fvm

`fvm` is distributed as a self-contained binary for Linux and macOS on AMD64 and ARM64.

This page focuses on the install experience for normal users, especially people who do not spend all day in shell configs.

## What this installs

The installer installs the `fvm` CLI itself.

It does not automatically install a Foundry version. After installing `fvm`, you still need to run commands like:

```sh
fvm list-remote
fvm install 13.346
fvm global 13.346
```

## Recommended install

Run the installer script directly from GitHub:

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  | sh
```

The script will:

1. detect your operating system and CPU architecture
2. fetch the latest release tag from the GitHub API
3. download the correct release archive from GitHub Releases
4. verify the SHA-256 checksum
5. extract and install the `fvm` binary
6. tell you if you need to update your `PATH`

## After installation

Run these commands first:

```sh
fvm version
fvm help
```

If those work, `fvm` is installed correctly.

Then set up shell integration:

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

Then install and select a Foundry version:

```sh
fvm list-remote
fvm install 13.346
fvm global 13.346
fvm current
```

If you skip those later steps, `fvm` will be installed but not doing anything useful yet.

## Choose a specific `fvm` release

If you want to install a specific `fvm` release instead of the latest one:

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  | VERSION=v0.2.0 sh
```

Or download the script first:

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  -o install.sh
VERSION=v0.2.0 sh install.sh
```

Important: this does not install Foundry `v0.2.0`. It installs `fvm` release `v0.2.0`.

Also important: this does not work the way many people expect:

```sh
VERSION=v0.2.0 curl ... | sh
```

That sets the variable on `curl`, not on the installer shell. Put the variable on `sh` instead.

## Choose an install directory

By default, the script installs to `/usr/local/bin` if it can write there. If not, it falls back to `$HOME/.local/bin`.

To choose your own install directory:

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  | INSTALL_DIR=/opt/bin sh
```

Or:

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  -o install.sh
INSTALL_DIR=/opt/bin sh install.sh
```

Again, this does not work the way people often assume:

```sh
INSTALL_DIR=/opt/bin curl ... | sh
```

That applies the variable to `curl`, not to the installer.

## PATH setup

If the installer says the binary directory is not on your `PATH`, add it to your shell profile.

The most common fallback install directory is `$HOME/.local/bin`.

### bash

Add this to `~/.bashrc` or `~/.bash_profile`:

```sh
export PATH="$PATH:$HOME/.local/bin"
```

Then reload your shell:

```sh
source ~/.bashrc
```

If your system uses `~/.bash_profile` instead, reload that file instead.

### zsh

Add this to `~/.zshrc`:

```sh
export PATH="$PATH:$HOME/.local/bin"
```

Then reload your shell:

```sh
source ~/.zshrc
```

### fish

Use fish's universal path support or add the directory to your fish config.

A common approach is:

```fish
fish_add_path $HOME/.local/bin
```

If you want it in your config file, add it to `~/.config/fish/config.fish`.

## Shell integration setup

Installing the `fvm` binary is not the same thing as enabling version switching for `foundry`.

For switching to work smoothly, you usually want the shell integration too.

### What shell integration does

It mainly does two things:

1. puts `~/.fvm/shims` on your `PATH`
2. gives you the `fvm_use` helper for temporary shell-only switching

### Temporary setup for this shell only

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

### Permanent setup

Add the same command to your shell startup file.

bash:

```sh
eval "$(fvm init bash)"
```

zsh:

```sh
eval "$(fvm init zsh)"
```

fish:

```fish
fvm init fish | source
```

Put them in:

- `~/.bashrc` or `~/.bash_profile` for bash
- `~/.zshrc` for zsh
- `~/.config/fish/config.fish` for fish

## Manual install

If you do not want to use the installer script, download a release archive from the [Releases page][releases].

### Pick the correct archive

| Platform    | Archive                          |
|-------------|----------------------------------|
| Linux AMD64 | `fvm_vX.Y.Z_linux_amd64.tar.gz`  |
| Linux ARM64 | `fvm_vX.Y.Z_linux_arm64.tar.gz`  |
| macOS AMD64 | `fvm_vX.Y.Z_darwin_amd64.tar.gz` |
| macOS ARM64 | `fvm_vX.Y.Z_darwin_arm64.tar.gz` |

### Verify the checksum

Download `checksums.sha256` from the same release and run:

```sh
sha256sum --check --ignore-missing checksums.sha256
```

If your system does not have `sha256sum`, use the equivalent checksum tool available on your platform.

### Extract and install the binary

```sh
tar -xzf fvm_vX.Y.Z_<os>_<arch>.tar.gz
install -m 0755 fvm /usr/local/bin/fvm
```

If `/usr/local/bin` is not writable, install to a directory you control and add that directory to your `PATH`.

## Build from source

Requirements: Go 1.22 or newer.

```sh
git clone https://github.com/foundry/fvm.git
cd fvm
go build -o fvm ./cmd/fvm
install -m 0755 fvm /usr/local/bin/fvm
```

## Verify everything worked

Use this checklist:

```sh
fvm version
fvm help
fvm init bash
```

If shell integration is enabled, also try:

```sh
fvm current
fvm doctor
```

If you have already installed a Foundry version, also try:

```sh
fvm which
```

## Common install problems

### `fvm: command not found`

The binary installed successfully, but its directory is not on your `PATH`.

Fix that first, then open a new shell or reload your shell config.

### The installer ran, but `foundry` still does not switch versions

That usually means one of these:

- shell integration was never enabled
- `~/.fvm/shims` is not on `PATH`
- you have not installed a Foundry version yet
- you have not selected a version yet with `global`, `local`, or `fvm_use`

### `fvm current` reports no configured version

That means `fvm` is installed, but you have not selected a version yet.

Run something like:

```sh
fvm install 13.346
fvm global 13.346
```

### `fvm which` says the version is not installed

You selected a version, but the actual Foundry files for that version are not present yet.

Install it with:

```sh
fvm install <version>
```

## Next step

After installation, go back to the project root README and follow the Quick start there.

[releases]: https://github.com/foundry/fvm/releases
