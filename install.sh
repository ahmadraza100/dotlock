#!/bin/sh
set -e

REPO="ahmadraza100/dotlock"
BINARY="dotlock"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "unsupported arch: $ARCH"; exit 1 ;;
esac

case "$OS" in
  darwin) OS="Darwin" ;;
  linux) OS="Linux" ;;
  *) echo "unsupported OS: $OS"; exit 1 ;;
esac

VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
URL="https://github.com/$REPO/releases/download/$VERSION/${BINARY}_${OS}_${ARCH}.tar.gz"

echo "installing dotlock $VERSION..."
curl -sL "$URL" | tar xz
sudo mv "$BINARY" /usr/local/bin/
echo "✓ done — run: dotlock --help"