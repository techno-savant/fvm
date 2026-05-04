# Installing fvm

`fvm` is distributed as a self-contained binary for Linux and macOS on both
AMD64 and ARM64 architectures.

## Quick install (recommended)

Run the installer script directly from the repository:

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  | sh
```

The script will:

1. Detect your operating system and CPU architecture.
2. Fetch the latest release tag from the GitHub API.
3. Download the appropriate archive from GitHub Releases.
4. Verify the SHA-256 checksum.
5. Extract and install the `fvm` binary.
6. Print PATH guidance if the install directory is not already on your `PATH`.

## Controlling the install

### Choose a specific version

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  | VERSION=v0.2.0 sh
```

Or equivalently, download the script first and run it with the variable set:

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  -o install.sh
VERSION=v0.2.0 sh install.sh
```

Important: `VERSION=... curl ... | sh` does not work the way people think.
The variable must be applied to `sh`, not `curl`.

### Choose an install directory

By default the script installs to `/usr/local/bin` (if writable) or
`$HOME/.local/bin`.  Override with `INSTALL_DIR`:

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  | INSTALL_DIR=/opt/bin sh
```

Important: `INSTALL_DIR=... curl ... | sh` applies the variable to `curl`, not
to the installer. Put the variable on `sh` or run the downloaded script
explicitly.

## Manual install

Download a pre-built archive from the [Releases page][releases]:

| Platform        | Archive                                   |
|-----------------|-------------------------------------------|
| Linux AMD64     | `fvm_vX.Y.Z_linux_amd64.tar.gz`          |
| Linux ARM64     | `fvm_vX.Y.Z_linux_arm64.tar.gz`          |
| macOS AMD64     | `fvm_vX.Y.Z_darwin_amd64.tar.gz`         |
| macOS ARM64     | `fvm_vX.Y.Z_darwin_arm64.tar.gz`         |

Verify the checksum against `checksums.sha256` (also on the Releases page):

```sh
sha256sum --check --ignore-missing checksums.sha256
```

Extract and install:

```sh
tar -xzf fvm_vX.Y.Z_<os>_<arch>.tar.gz
install -m 0755 fvm /usr/local/bin/fvm
```

## Build from source

Requirements: Go 1.22 or newer.

```sh
git clone https://github.com/foundry/fvm.git
cd fvm
go build -o fvm ./cmd/fvm
install -m 0755 fvm /usr/local/bin/fvm
```

## Verify the installation

```sh
fvm version
```

Expected output: `fvm vX.Y.Z`

## PATH configuration

If `fvm` is installed to a directory not already on your `PATH`, add it to
your shell profile.  For `$HOME/.local/bin`:

**bash** (`~/.bashrc` or `~/.bash_profile`):
```sh
export PATH="$PATH:$HOME/.local/bin"
```

**zsh** (`~/.zshrc`):
```sh
export PATH="$PATH:$HOME/.local/bin"
```

After editing your profile, restart your shell or run:

```sh
source ~/.bashrc   # bash
source ~/.zshrc    # zsh
```

[releases]: https://github.com/foundry/fvm/releases
