#!/bin/bash
set -e

REPO="yugarinn/tinkershell"
BINARY_NAME="tinkershell"
INSTALL_DIR="/usr/local/bin"

echo "=> Checking for latest tinkershell version..."

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "=> Unsupported architecture: $ARCH"; exit 1 ;;
esac

LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo "=> Error: Could not find latest release for $REPO"
    exit 1
fi

FILENAME="${BINARY_NAME}-${LATEST_RELEASE}-${OS}-${ARCH}"
URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/$FILENAME"

echo "=> Downloading $BINARY_NAME $LATEST_RELEASE for $OS/$ARCH..."

TMP_BIN="/tmp/$BINARY_NAME"
curl -L -o "$TMP_BIN" "$URL"
chmod +x "$TMP_BIN"

echo "=> Moving binary to $INSTALL_DIR (requires sudo)..."
sudo mv "$TMP_BIN" "$INSTALL_DIR/$BINARY_NAME"

echo "=> Successfully installed $BINARY_NAME version $LATEST_RELEASE!"
tinkershell --version
