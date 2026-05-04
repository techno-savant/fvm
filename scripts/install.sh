#!/usr/bin/env sh
set -eu

OWNER_REPO="foundry/fvm"
BINARY_NAME="fvm"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${VERSION:-latest}"

cat <<EOF
fvm installer stub

This script is a placeholder for curl-based installation.

Planned behavior:
- detect OS and architecture
- resolve a release artifact for ${OWNER_REPO}
- download the ${BINARY_NAME} archive
- install into ${INSTALL_DIR}

Current requested version: ${VERSION}
EOF
