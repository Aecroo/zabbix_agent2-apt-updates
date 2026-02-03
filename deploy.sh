#!/bin/bash
set -e

echo "=== Zabbix Agent 2 APT Updates Plugin Deployment ==="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root or with sudo."
    exit 1
fi

# Detect architecture
ARCH=$(uname -m)
BINARY_URL=""
BINARY_NAME=""

echo "Detecting system architecture..."
case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        BINARY_NAME="zabbix-agent2-plugin-apt-updates-linux-amd64"
        ;;
    armv7*|armv6*)
        ARCH="armv7"
        BINARY_NAME="zabbix-agent2-plugin-apt-updates-linux-armv7"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        BINARY_NAME="zabbix-agent2-plugin-apt-updates-linux-arm64"
        ;;
    *)
        echo "Error: Unsupported architecture: $ARCH"
        echo "Supported architectures: x86_64/amd64, armv7, arm64/aarch64"
        exit 1
        ;;
esac

echo "Detected architecture: $ARCH"

# Base URL for releases (update this to your actual release URL)
BASE_URL="http://192.168.0.23:3000/zbx/zabbix_agent2-apt-updates/releases/download/v0.7.0"
BINARY_URL="${BASE_URL}/${BINARY_NAME}"
CONFIG_URL="http://192.168.0.23:3000/zbx/zabbix_agent2-apt-updates/raw/branch/master/apt-updates.conf"

# Installation paths
INSTALL_DIR="/etc/zabbix"
PLUGIN_PATH="${INSTALL_DIR}/${BINARY_NAME}"
CONFIG_DIR="/etc/zabbix/zabbix_agent2.d"
CONFIG_FILE="${CONFIG_DIR}/apt-updates.conf"

# Create directories
echo "Creating installation directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"

# Download binary
echo "Downloading plugin binary for $ARCH..."
if command -v wget &> /dev/null; then
    wget -q "$BINARY_URL" -O "$PLUGIN_PATH"
elif command -v curl &> /dev/null; then
    curl -sL "$BINARY_URL" -o "$PLUGIN_PATH"
else
    echo "Error: Neither wget nor curl is available."
    exit 1
fi

# Set executable permissions
echo "Setting permissions..."
chmod +x "$PLUGIN_PATH"

# Download configuration file
echo "Downloading configuration file..."
if command -v wget &> /dev/null; then
    wget -q "$CONFIG_URL" -O "$CONFIG_FILE"
elif command -v curl &> /dev/null; then
    curl -sL "$CONFIG_URL" -o "$CONFIG_FILE"
else
    # If neither wget nor curl available for config, create a basic one
    cat > "$CONFIG_FILE" <<EOF
Plugins.APTUpdates.System.Path=$PLUGIN_PATH
EOF
fi

# Update configuration with correct path
echo "Updating configuration file..."
cat > "$CONFIG_FILE" <<EOF
### Option: Plugins.APTUpdates.System.Path
#	Path to APT updates plugin executable.
#
# Mandatory: yes
# Default:
# Plugins.APTUpdates.System.Path=

Plugins.APTUpdates.System.Path=$PLUGIN_PATH

### Option: Plugins.APTUpdates.Timeout
#	Specifies the wait time (in seconds) for apt commands to respond.
# Note: With Zabbix version 7.0 and later, you can configure timeout directly in the item configuration
# with a range of 1-600 seconds (10 minutes), which overrides this plugin-level setting.
# This plugin-level timeout is used as a fallback for older versions or when not specified at the item level.
#
# Mandatory: no
# Default:
# Plugins.APTUpdates.Timeout=<Global timeout>
EOF

# Set proper ownership (zabbix user if exists)
if id "zabbix" &> /dev/null; then
    echo "Setting ownership to zabbix user..."
    chown zabbix:zabbix "$PLUGIN_PATH"
    chown zabbix:zabbix "$CONFIG_FILE"
fi

echo ""
echo "=== Deployment Complete ==="
echo ""
echo "Installed files:"
echo "  Binary: $PLUGIN_PATH"
echo "  Config: $CONFIG_FILE"
echo ""
echo "Next steps:"
echo "  1. Restart Zabbix Agent 2: sudo systemctl restart zabbix-agent2"
echo "  2. Verify installation: sudo -u zabbix $PLUGIN_PATH --version"
echo ""

# Check if zabbix-agent2 service exists and suggest restart
if systemctl list-unit-files | grep -q zabbix-agent2; then
    echo "Note: Zabbix Agent 2 service detected. Consider running:"
    echo "  sudo systemctl restart zabbix-agent2"
fi

echo "Deployment successful!"
