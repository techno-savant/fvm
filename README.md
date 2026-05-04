# fvm

`fvm` is a minimal Go CLI scaffold for future Foundry version-management work.

## Current status

This repository currently provides a tiny CLI that supports:

- `fvm version`
- `fvm help`
- running `fvm` with no arguments to show help

## Development

Requirements:

- Go 1.22+

Common commands:

```sh
go fmt ./...
go test ./...
go build ./cmd/fvm
```

## Install via curl

A starter installer script is included at `scripts/install.sh`.

Example:

```sh
curl -fsSL https://example.com/fvm/install.sh | sh
```

The script is currently a stub intended to be adapted once release artifacts are published.
