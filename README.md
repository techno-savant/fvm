# fvm

`fvm` is a Foundry version manager — a CLI tool for managing Foundry
installations across projects.

## Install

### Quick install

```sh
curl -fsSL \
  https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh \
  | sh
```

The installer detects your OS and architecture, downloads the matching release
artifact from GitHub, verifies the SHA-256 checksum, and places the binary in
`/usr/local/bin` (or `$HOME/.local/bin` if `/usr/local/bin` is not writable).

For full installation options — specific versions, custom install directories,
manual downloads, and PATH setup — see [docs/install.md](docs/install.md).

### Build from source

Requirements: Go 1.22+

```sh
git clone https://github.com/foundry/fvm.git
cd fvm
go build -o fvm ./cmd/fvm
```

## Usage

```
fvm <command>

Commands:
  version   Print the current fvm version
  help      Show this help message
```

Examples:

```sh
fvm version
fvm help
```

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

### CI

The repository runs a GitHub Actions workflow on every push and pull request
that checks formatting, runs tests with `-race`, and builds the binary.  A
separate release workflow fires on `v*` tags and produces cross-compiled
archives plus a `checksums.sha256` file attached to the GitHub Release.

## Releasing

Push a version tag to trigger the release workflow:

```sh
git tag v0.2.0
git push origin v0.2.0
```

The workflow builds binaries for all supported platforms, packages them as
`.tar.gz` archives, generates SHA-256 checksums, and creates a GitHub Release
with all artifacts attached.

## Supported platforms

| OS     | Architecture |
|--------|--------------|
| Linux  | amd64        |
| Linux  | arm64        |
| macOS  | amd64        |
| macOS  | arm64        |

## License

See [LICENSE](LICENSE) (if present) for details.
