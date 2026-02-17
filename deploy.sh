#!/bin/bash
set -e

echo "=== Zabbix Agent 2 APT Updates Plugin Deployment ==="
echo ""

# Parse command-line arguments for Ansible integration
while [[ $# -gt 0 ]]; do
    case "$1" in
        --plugin-dir)
            PLUGIN_DIR="$2"
            shift 2
            ;;
        --config-dir)
            CONFIG_DIR="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Check if running as root (only when not called by Ansible)
if [ "$EUID" -ne 0 ] && [ -z "${PLUGIN_DIR}" ]; then
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
BASE_URL="http://192.168.0.23:3000/zbx/zabbix_agent2-apt-updates/releases/download/v1.0.0"
BINARY_URL="${BASE_URL}/${BINARY_NAME}"
CONFIG_URL="http://192.168.0.23:3000/zbx/zabbix_agent2-apt-updates/raw/branch/master/apt-updates.conf"

# Installation paths
INSTALL_DIR=${PLUGIN_DIR:-/etc/zabbix}
PLUGIN_PATH="${INSTALL_DIR}/${BINARY_NAME}"
CONFIG_DIR=${CONFIG_DIR:-/etc/zabbix/zabbix_agent2.d}
CONFIG_FILE="${CONFIG_DIR}/apt-updates.conf"

# Create directories
echo "Creating installation directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"

# Try to remove existing binary forcefully before download
if [ -f "$PLUGIN_PATH" ]; then
    echo "Removing existing plugin binary..."
    rm -f "$PLUGIN_PATH" || {
        # If regular remove fails, try with fuser
        if command -v fuser &> /dev/null; then
            fuser -k "$PLUGIN_PATH" 2>/dev/null || true
            sleep 1
            rm -f "$PLUGIN_PATH" || true
        fi
    }
fi

# Stop Zabbix Agent 2 service if running
if systemctl list-unit-files | grep -q zabbix-agent2; then
    if systemctl is-active --quiet zabbix-agent2 2>/dev/null; then
        echo "Zabbix Agent 2 service is running. Stopping to update plugin..."
        systemctl stop zabbix-agent2 || {
            echo "Warning: Could not stop Zabbix Agent 2, but continuing anyway"
        }
    fi
fi

# Download binary
echo "Downloading plugin binary for $ARCH..."
# Create a temporary installation directory to avoid file busy issues completely
TMP_INSTALL_DIR="/tmp/zabbix-plugin-install-$$"
mkdir -p "$TMP_INSTALL_DIR"
TEMP_BINARY="${TMP_INSTALL_DIR}/${BINARY_NAME}"

if command -v wget &> /dev/null; then
    wget -q "$BINARY_URL" -O "$TEMP_BINARY"
elif command -v curl &> /dev/null; then
    curl -sL "$BINARY_URL" -o "$TEMP_BINARY"
else
    echo "Error: Neither wget nor curl is available."
    exit 1
fi

# Set executable permissions on the temporary binary
echo "Setting permissions on downloaded plugin..."
chmod +x "$TEMP_BINARY"

# Move temporary installation to final location after all downloads complete
echo "Moving plugin to final installation directory..."
mv -f "$TMP_INSTALL_DIR/${BINARY_NAME}" "$PLUGIN_PATH" 2>/dev/null || mv -f "$TEMP_BINARY" "$PLUGIN_PATH"
rm -rf "$TMP_INSTALL_DIR"

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

# Restart Zabbix Agent 2 service if it was stopped during update
if systemctl list-unit-files | grep -q zabbix-agent2; then
    if ! systemctl is-active --quiet zabbix-agent2 2>/dev/null; then
        echo "Starting Zabbix Agent 2 service after plugin update..."
        systemctl start zabbix-agent2 || {
            echo "Warning: Could not start Zabbix Agent 2"
        }
    else
        # If it was already running (not stopped by us), restart to load new plugin
        echo "Restarting Zabbix Agent 2 service to load updated plugin..."
        systemctl restart zabbix-agent2 || {
            echo "Warning: Could not restart Zabbix Agent 2"
        }
    fi
fi

echo "Deployment successful!"
