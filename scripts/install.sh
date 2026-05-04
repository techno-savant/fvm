#!/usr/bin/env sh
# fvm installer
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/foundry/fvm/main/scripts/install.sh | sh
#
# Environment variables:
#   VERSION     - release tag to install (default: latest)
#   INSTALL_DIR - directory to install the binary (default: auto-detected)
#
# Usage:
#   curl -fsSL https://example.com/fvm/install.sh | sh
#   curl -fsSL https://example.com/fvm/install.sh | VERSION=v0.2.0 sh
#   curl -fsSL https://example.com/fvm/install.sh -o install.sh; VERSION=v0.2.0 sh install.sh
set -eu

OWNER="techno-savant"
REPO="fvm"
BINARY_NAME="fvm"
VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-}"

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "error: required command not found: $1" >&2
    exit 1
  fi
}

extract_tag_name() {
  sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1
}

make_temp_dir() {
  if tmp="$(mktemp -d 2>/dev/null)"; then
    printf '%s\n' "$tmp"
    return 0
  fi
  if tmp="$(mktemp -d -t fvm 2>/dev/null)"; then
    printf '%s\n' "$tmp"
    return 0
  fi
  echo "error: could not create temporary directory" >&2
  exit 1
}

# ---------------------------------------------------------------------------
# Detect OS
# ---------------------------------------------------------------------------
OS="$(uname -s)"
case "$OS" in
  Linux)  OS="linux"  ;;
  Darwin) OS="darwin" ;;
  *)
    echo "error: unsupported operating system: $OS" >&2
    exit 1
    ;;
esac

# ---------------------------------------------------------------------------
# Detect architecture
# ---------------------------------------------------------------------------
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)   ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *)
    echo "error: unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

# ---------------------------------------------------------------------------
# Resolve install directory
# ---------------------------------------------------------------------------
if [ -z "$INSTALL_DIR" ]; then
  INSTALL_DIR="/usr/local/bin"
  if [ -d "$INSTALL_DIR" ]; then
    if [ ! -w "$INSTALL_DIR" ]; then
      INSTALL_DIR="${HOME}/.local/bin"
    fi
  elif mkdir -p "$INSTALL_DIR" 2>/dev/null; then
    :
  else
    INSTALL_DIR="${HOME}/.local/bin"
  fi
fi

mkdir -p "$INSTALL_DIR"
if [ ! -w "$INSTALL_DIR" ]; then
  echo "error: install directory is not writable: $INSTALL_DIR" >&2
  echo "set INSTALL_DIR to a writable path or rerun with appropriate permissions" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Resolve version tag
# ---------------------------------------------------------------------------
if [ "$VERSION" = "latest" ]; then
  need_cmd curl
  printf 'Resolving latest release... '
  VERSION="$(curl -fsSL \
    "https://api.github.com/repos/${OWNER}/${REPO}/releases/latest" \
    | extract_tag_name)"
  if [ -z "$VERSION" ]; then
    echo ""
    echo "error: could not determine latest release version" >&2
    exit 1
  fi
  echo "$VERSION"
fi

# Ensure version tag starts with 'v'
case "$VERSION" in
  v*) VERSION_TAG="$VERSION" ;;
  *)  VERSION_TAG="v${VERSION}" ;;
esac

# ---------------------------------------------------------------------------
# Construct URLs
# ---------------------------------------------------------------------------
ARCHIVE_NAME="${BINARY_NAME}_${VERSION_TAG}_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/${OWNER}/${REPO}/releases/download/${VERSION_TAG}"
DOWNLOAD_URL="${BASE_URL}/${ARCHIVE_NAME}"
CHECKSUM_URL="${BASE_URL}/checksums.sha256"

echo "Installing ${BINARY_NAME} ${VERSION_TAG} (${OS}/${ARCH})"
echo "  source:  ${DOWNLOAD_URL}"
echo "  dest:    ${INSTALL_DIR}/${BINARY_NAME}"

# ---------------------------------------------------------------------------
# Download to a temp directory (cleaned up on exit)
# ---------------------------------------------------------------------------
TMP_DIR="$(make_temp_dir)"
trap 'rm -rf "$TMP_DIR"' EXIT INT TERM

need_cmd tar
need_cmd install
need_cmd grep
need_cmd awk
need_cmd curl

echo ""
printf 'Downloading archive... '
curl -fsSL "$DOWNLOAD_URL" -o "${TMP_DIR}/${ARCHIVE_NAME}"
echo "done"

# ---------------------------------------------------------------------------
# Verify checksum (best-effort: skip if neither sha256sum nor shasum present)
# ---------------------------------------------------------------------------
CHECKSUM_CMD=""
if command -v sha256sum >/dev/null 2>&1; then
  CHECKSUM_CMD="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  CHECKSUM_CMD="shasum -a 256"
else
  echo "error: sha256sum or shasum is required for checksum verification" >&2
  exit 1
fi

printf 'Verifying checksum... '
curl -fsSL "$CHECKSUM_URL" -o "${TMP_DIR}/checksums.sha256"
EXPECTED="$(grep -F "  ${ARCHIVE_NAME}" "${TMP_DIR}/checksums.sha256" | awk 'NR==1 {print $1}')"
if [ -z "$EXPECTED" ]; then
  EXPECTED="$(grep -F " ${ARCHIVE_NAME}" "${TMP_DIR}/checksums.sha256" | awk 'NR==1 {print $1}')"
fi
if [ -z "$EXPECTED" ]; then
  echo ""
  echo "error: ${ARCHIVE_NAME} not found in checksums file" >&2
  exit 1
fi
ACTUAL="$($CHECKSUM_CMD "${TMP_DIR}/${ARCHIVE_NAME}" | awk '{print $1}')"
if [ "$EXPECTED" != "$ACTUAL" ]; then
  echo ""
  echo "error: checksum mismatch" >&2
  echo "  expected: ${EXPECTED}" >&2
  echo "  actual:   ${ACTUAL}" >&2
  exit 1
fi
echo "ok"

# ---------------------------------------------------------------------------
# Extract and install
# ---------------------------------------------------------------------------
printf 'Extracting... '
tar -xzf "${TMP_DIR}/${ARCHIVE_NAME}" -C "$TMP_DIR"
echo "done"

if [ ! -f "${TMP_DIR}/${BINARY_NAME}" ]; then
  echo "error: extracted archive did not contain ${BINARY_NAME} at archive root" >&2
  exit 1
fi

install -m 0755 "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"

# ---------------------------------------------------------------------------
# PATH guidance
# ---------------------------------------------------------------------------
echo ""
echo "  ${BINARY_NAME} ${VERSION_TAG} installed to ${INSTALL_DIR}/${BINARY_NAME}"
echo ""

case ":${PATH}:" in
  *":${INSTALL_DIR}:"*)
    ;;
  *)
    echo "NOTE: ${INSTALL_DIR} is not in your PATH."
    echo "Add the following line to your shell profile"
    echo "(~/.bashrc, ~/.zshrc, ~/.profile, etc.) and restart your shell:"
    echo ""
    echo "    export PATH=\"\$PATH:${INSTALL_DIR}\""
    echo ""
    ;;
esac

echo "Run '${BINARY_NAME} help' to get started."
