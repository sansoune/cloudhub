#!/bin/bash
set -e

echo "Cloudhub Quick Install"
echo "========================="
echo ""

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    armv7l)
        ARCH="armv7"
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

if [ "$OS" != "linux" ]; then
    echo "Unsupported OS: $OS"
    echo "Cloudhub currently and will support Linux only"
    exit 1
fi

echo "✅ System: $OS $ARCH"
echo ""

echo "Downloading latest release..."
RELEASE_URL="https://github.com/sansoune/cloudhub/releases/latest/download/cloudhub-${OS}-${ARCH}"

curl -L -o cloudhub "$RELEASE_URL"

if [ ! -f "cloudhub" ]; then
    echo "Download failed"
    exit 1
fi 

echo "✅ Downloaded successfully"
echo ""

echo "Installing..."
chmod +x cloudhub
sudo mv cloudhub /usr/local/bin/cloudhub

echo "✅ Installed to /usr/local/bin/cloudhub"
echo ""

echo "Initializing configuration..."
cloudhub config init
echo ""

echo "Installing daemon..."
cloudhub daemon install

echo ""
echo "✅ Installation complete!"
echo ""
echo "Next steps:"
echo "   1. Edit config: nano ~/.config/cloudhub/config.yaml"
echo "   2. Start daemon: cloudhub daemon start"
echo ""
echo "Documentation: https://github.com/sansoune/cloudhub"



