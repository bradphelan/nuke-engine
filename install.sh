#!/bin/sh
# nuke-build installer for Linux and macOS
# Usage: curl -sSfL https://raw.githubusercontent.com/bradphelan/nuke-engine/main/install.sh | sh

set -e

REPO="bradphelan/nuke-engine"
BINARY="nuke-build"
INSTALL_DIR="${HOME}/.local/bin"

echo "Installing ${BINARY}..."

# Detect OS
OS=$(uname -s)
case "${OS}" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)
    echo "Unsupported OS: ${OS}" >&2
    exit 1
    ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "${ARCH}" in
  x86_64)         ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: ${ARCH}" >&2
    exit 1
    ;;
esac

ASSET="${BINARY}-${OS}-${ARCH}"

# Fetch latest release version
VERSION=$(curl -sSfL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

if [ -z "${VERSION}" ]; then
  echo "Failed to determine latest release version" >&2
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

mkdir -p "${INSTALL_DIR}"
echo "Downloading ${ASSET} (${VERSION})..."
curl -sSfL "${URL}" -o "${INSTALL_DIR}/${BINARY}"
chmod +x "${INSTALL_DIR}/${BINARY}"

echo "Installed to ${INSTALL_DIR}/${BINARY}"

# Remind about PATH if needed
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo ""
    echo "NOTE: Add ${INSTALL_DIR} to your PATH:"
    echo '  export PATH="${HOME}/.local/bin:${PATH}"'
    ;;
esac
