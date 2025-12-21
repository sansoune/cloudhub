#!/bin/bash

echo "Cloudhub Quick Install"
echo "========================="
echo ""

echo "Downloading latest release..."
RELEASE_URL="https://github.com/sansoune/cloudhub/releases/latest/download/cloudhub"

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



