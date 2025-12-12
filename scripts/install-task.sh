#!/bin/bash

# Install Task (taskfile)
# https://taskfile.dev/installation/

set -e

echo "Installing Task..."

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture names
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    arm64)
        ARCH="arm64"
        ;;
    aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Download URL
VERSION=$(curl -s https://api.github.com/repos/go-task/task/releases/latest | grep '"tag_name":' | sed -E 's/.*"tag_name": ?"v?([^"]+).*/\1/')
URL="https://github.com/go-task/task/releases/latest/download/task_${OS}_${ARCH}.tar.gz"

# Create temp directory
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR

# Download and extract
echo "Downloading Task from: $URL"
curl -L -o task.tar.gz "$URL"
tar -xzf task.tar.gz

# Install to /usr/local/bin
if [ -w /usr/local/bin ]; then
    sudo mv task /usr/local/bin/
else
    mkdir -p ~/.local/bin
    mv task ~/.local/bin/
    echo "Adding ~/.local/bin to PATH if not already present..."
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc 2>/dev/null || true
fi

# Cleanup
cd -
rm -rf $TEMP_DIR

echo "✅ Task installed successfully!"
echo "Run 'task --version' to verify installation."
echo "Run 'task' to see available tasks."