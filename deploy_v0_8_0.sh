#!/bin/bash
# Deployment script for Zabbix Agent 2 APT Updates Plugin v0.8.0
# This script automates the installation/upgrade process

set -e

echo "=========================================="
echo "Zabbix Agent 2 APT Updates Plugin v0.8.0"
echo "Deployment Script"
echo "=========================================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root (or with sudo)"
    exit 1
fi

# Detect system architecture
ARCH=$(uname -m)
BINARY_NAME=""
case "$ARCH" in
    x86_64|amd64)
        BINARY_NAME="zabbix-agent2-plugin-apt-updates-linux-amd64"
        ;;
    aarch64|arm64)
        BINARY_NAME="zabbix-agent2-plugin-apt-updates-linux-arm64"
        ;;
    armv7*|armv6*)
        BINARY_NAME="zabbix-agent2-plugin-apt-updates-linux-armv7"
        ;;
    *)
        echo "Error: Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "Detected architecture: $ARCH"
echo "Using binary: $BINARY_NAME"

# Download URLs (from Gitea release)
BASE_URL="http://192.168.0.23:3000/zbx/zabbix_agent2-apt-updates/releases/download/v0.8.0"
BINARY_URL="$BASE_URL/$BINARY_NAME"
CONFIG_URL="http://192.168.0.23:3000/zbx/zabbix_agent2-apt-updates/raw/branch/master/apt-updates.conf"
TEMPLATE_URL="http://192.168.0.23:3000/zbx/zabbix_agent2-apt-updates/raw/branch/master/templates/7.4/apt_updates_zabbix_agent2.yaml"

# Installation directories
BIN_DIR="/usr/libexec/zabbix/"
CONFIG_DIR="/etc/zabbix/zabbix_agent2.d/"
TMP_DIR="/tmp/zabbix-apt-updates-deploy/"

# Create temporary directory
mkdir -p "$TMP_DIR"
echo "Using temporary directory: $TMP_DIR"

# Download binary
echo "Downloading plugin binary..."
if ! wget -qO "$TMP_DIR/$BINARY_NAME" "$BINARY_URL"; then
    echo "Error: Failed to download plugin binary"
    echo "Please check your internet connection and the release URL"
    exit 1
fi
chmod +x "$TMP_DIR/$BINARY_NAME"

# Download configuration
echo "Downloading configuration file..."
if ! wget -qO "$TMP_DIR/apt-updates.conf" "$CONFIG_URL"; then
    echo "Error: Failed to download configuration file"
    exit 1
fi

# Create installation directories
mkdir -p "$BIN_DIR"
mkdir -p "$CONFIG_DIR"
echo "Created installation directories"

# Install binary
echo "Installing plugin binary..."
cp -f "$TMP_DIR/$BINARY_NAME" "$BIN_DIR/zabbix-agent2-plugin-apt-updates"
chown root:root "$BIN_DIR/zabbix-agent2-plugin-apt-updates"
echo "Binary installed to $BIN_DIR/zabbix-agent2-plugin-apt-updates"

# Install configuration
echo "Installing configuration file..."
cp -f "$TMP_DIR/apt-updates.conf" "$CONFIG_DIR/apt-updates.conf"
chown root:root "$CONFIG_DIR/apt-updates.conf"
echo "Configuration installed to $CONFIG_DIR/apt-updates.conf"

# Verify installation
echo "Verifying installation..."
if [ -f "$BIN_DIR/zabbix-agent2-plugin-apt-updates" ] && [ -f "$CONFIG_DIR/apt-updates.conf" ]; then
    echo "✓ Installation successful!"
else
    echo "Error: Installation verification failed"
    exit 1
fi

# Check if Zabbix Agent 2 is installed
if systemctl list-units --type=service | grep -q zabbix-agent2; then
    echo "Zabbix Agent 2 service detected"
    # Restart Zabbix Agent 2
    echo "Restarting Zabbix Agent 2..."
    systemctl restart zabbix-agent2
    if [ $? -eq 0 ]; then
        echo "✓ Zabbix Agent 2 restarted successfully"
    else
        echo "Warning: Failed to restart Zabbix Agent 2"
    fi
else
    echo "Note: Zabbix Agent 2 service not detected or not running"
    echo "You may need to start it manually after configuration"
fi

# Cleanup
echo "Cleaning up temporary files..."
rm -rf "$TMP_DIR"

# Display completion message
echo ""
echo "=========================================="
echo "Deployment Complete!"
echo "=========================================="
echo "Version: 0.8.0"
echo "Plugin binary: $BIN_DIR/zabbix-agent2-plugin-apt-updates"
echo "Configuration: $CONFIG_DIR/apt-updates.conf"
echo ""
echo "To use the new phased updates features:"
echo "1. Create a Zabbix item with key: updates.get"
echo "2. Add JSONPath preprocessing to extract specific fields:"
echo "   - Phased count: $.phased_updates_count"
echo "   - Phased list: $.phased_updates_list"
echo "   - Phased details: $.phased_updates_details"
echo "3. Import the updated template from templates/7.4/"
echo ""
echo "For more information, visit:"
echo "http://192.168.0.23:3000/zbx/zabbix_agent2-apt-updates"
