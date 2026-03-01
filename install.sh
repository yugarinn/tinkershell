#!/bin/bash
set -e

REPO="yugarinn/tinkershell"
BINARY_NAME="tinkershell"

echo "=> Checking for latest $BINARY_NAME version..."

UNAME_S=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$UNAME_S" in
    linux*)   OS="linux";  EXT="";     INSTALL_DIR="/usr/local/bin" ;;
    darwin*)  OS="darwin"; EXT="";     INSTALL_DIR="/usr/local/bin" ;;
    msys*|mingw*) OS="windows"; EXT=".exe"; INSTALL_DIR="/usr/bin" ;;
    *) echo "=> Unsupported OS: $UNAME_S"; exit 1 ;;
esac

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

FILENAME="${BINARY_NAME}-${LATEST_RELEASE}-${OS}-${ARCH}${EXT}"
URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/$FILENAME"

echo "=> Downloading $BINARY_NAME $LATEST_RELEASE for $OS/$ARCH..."

TMP_BIN="./${BINARY_NAME}_tmp${EXT}"
curl -L -o "$TMP_BIN" "$URL"
chmod +x "$TMP_BIN"

FINAL_DEST="$INSTALL_DIR/$BINARY_NAME$EXT"

IS_WINDOWS=false
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OS" == "Windows_NT" ]]; then
    IS_WINDOWS=true
fi

if [ "$IS_WINDOWS" = true ]; then
    if command -v cygpath >/dev/null 2>&1; then
        WIN_INSTALL_DIR=$(cygpath -w "$LOCALAPPDATA/tinkershell/bin")
    else
        WIN_INSTALL_DIR=$(echo "$LOCALAPPDATA/tinkershell/bin" | sed 's/\//\\/g' | sed 's/^\\c/C:/')
    fi

    INSTALL_DIR="${LOCALAPPDATA:-$HOME/AppData/Local}/tinkershell/bin"
    FINAL_DEST="$INSTALL_DIR/tinkershell.exe"

    echo "=> Installing to: $WIN_INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"
    
    rm -f "$FINAL_DEST" 
    mv "$TMP_BIN" "$FINAL_DEST"
    chmod +x "$FINAL_DEST"

    if ! echo "$PATH" | grep -iq "$INSTALL_DIR"; then
        echo "=> Adding to User PATH..."
        powershell.exe -Command "[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', 'User') + ';$WIN_INSTALL_DIR', 'User')"
    fi
    
    echo "=> Installed successfully."
    echo "=> IMPORTANT: Restart your terminal or 'source ~/.bashrc' if using Git Bash."
else
    if command -v sudo >/dev/null 2>&1; then
        echo "=> Moving binary to $INSTALL_DIR (requires sudo)..."
        sudo mv "$TMP_BIN" "$FINAL_DEST"
    else
        echo "=> Moving binary to $INSTALL_DIR..."
        mv "$TMP_BIN" "$FINAL_DEST"
    fi
fi

if [ "$IS_WINDOWS" = true ]; then
    CONF_DIR="$(cygpath -u "$APPDATA")/tinkershell"
else
    CONF_DIR="$HOME/.config/tinkershell"
fi

if [ ! -d "$CONF_DIR" ]; then
    echo "=> Creating config directory at $CONF_DIR..."
    mkdir -p "$CONF_DIR"
    
    if [ ! -f "$CONF_DIR/tinkershell.toml" ]; then
        touch "$CONF_DIR/tinkershell.toml"
        echo "=> Created empty config file"
    fi
fi

echo "=> Successfully installed $BINARY_NAME version $LATEST_RELEASE"
"$BINARY_NAME" --version
