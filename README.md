# Zabbix Agent 2 APT Updates Plugin

[![Docker](https://img.shields.io/badge/Docker-Supported-blue)](docker-compose.yml)
[![Version](https://img.shields.io/badge/version-1.0.0-blue)](CHANGELOG.md)

A monitoring plugin for Zabbix Agent 2 that checks available package updates on Debian/Ubuntu systems using APT.

## Quick Deployment

Deploy with a single command:

```bash
wget -qO- https://raw.githubusercontent.com/Aecroo/zabbix_agent2-apt-updates/master/deploy.sh | sudo bash
```

This will automatically detect your system architecture (amd64, armv7, or arm64), download the correct binary, install it to `/etc/zabbix/`, create the configuration file at `/etc/zabbix/zabbix_agent2.d/apt-updates.conf`, and set proper permissions.

After deployment, restart Zabbix Agent:
```bash
sudo systemctl restart zabbix-agent2
```

## Overview

This plugin detects available system updates by executing `apt list --upgradable` and returns the count of available updates in a format compatible with Zabbix Agent 2.

## 📋 Normal User Guide - Install on Ubuntu/Debian

This section provides step-by-step instructions for installing and using this plugin on a standard Ubuntu or Debian system.

### Prerequisites

- Ubuntu 20.04 LTS or later, or Debian 10 or later
- Zabbix Agent 2 installed and configured
- Basic command line knowledge

### Step 1: Download the Pre-built Binary

Download the latest release from our Git repository:

```bash
# Create a directory for the plugin
sudo mkdir -p /usr/libexec/zabbix/

# Download the binary (Ubuntu/Debian x86_64)
wget https://github.com/Aecroo/zabbix_agent2-apt-updates/releases/download/v1.0.0/zabbix-agent2-plugin-apt-updates-linux-amd64 \
  -O /usr/libexec/zabbix/zabbix-agent2-plugin-apt-updates

# Make it executable
sudo chmod +x /usr/libexec/zabbix/zabbix-agent2-plugin-apt-updates
```

### Step 2: Configure Zabbix Agent 2

Create a configuration file for the plugin:

```bash
# Download the configuration file
wget https://raw.githubusercontent.com/Aecroo/zabbix_agent2-apt-updates/master/apt-updates.conf \
  -O /etc/zabbix/zabbix_agent2.d/apt-updates.conf
```

Save and exit (Ctrl+O, Enter, Ctrl+X in nano).

### Step 3: Test the Plugin

Before restarting Zabbix Agent, test if the plugin works:

```bash
# Run a manual check (as root or zabbix user)
sudo -u zabbix /usr/libexec/zabbix/zabbix-agent2-plugin-apt-updates --version

# Example output:
Zabbix APTUpdates plugin
Version 1.0.0, built with go1.24.12
Protocol version 6.4.0

### Step 4: Restart Zabbix Agent

Apply the configuration changes:

```bash
sudo systemctl restart zabbix-agent2
```

### Step 5: Verify in Zabbix

1. In your Zabbix web interface, go to **Configuration** > **Hosts**
2. Select your monitored host
3. Go to the **Items** tab
4. Create a new item with:
   - **Type**: Zabbix Agent (active)
   - **Key**: `updates.get`
   - **Type of information**: Text
5. Save and wait for data collection

### Step 6: Set Up Monitoring (Optional)

The plugin returns comprehensive JSON data that can be processed using Zabbix's JSONPath preprocessing. For detailed information about creating items with JSONPath preprocessing and trigger configurations, see the [template documentation](templates/7.4/README.md).

Example JSONPath expressions:
- Security updates count: `.security_updates_count`
- All updates count: `.all_updates_count`
- Security updates list: `.security_updates_list`
- Package details: `.all_updates_details[*].name`
- Phased updates (NEW in v0.7.0):
  - Count: `.phased_updates_count`
  - List: `.phased_updates_list`
  - Details: `.phased_updates_details`

### Step 7: Create Triggers (Optional)

For trigger configurations, see the [template documentation](templates/7.4/README.md) which includes pre-configured triggers for security, recommended, and optional updates.

### Updating the Plugin

When new versions are released, simply download and replace the binary:

```bash
# Download the new version
sudo wget https://github.com/Aecroo/zabbix_agent2-apt-updates/releases/download/v1.0.0/zabbix-agent2-plugin-apt-updates-linux-amd64 \
  -O /usr/libexec/zabbix/zabbix-agent2-plugin-apt-updates

# Restart Zabbix Agent
sudo systemctl restart zabbix-agent2
```

### Troubleshooting for Normal Users

**Problem**: Plugin returns "command not found" error
- **Solution**: Ensure the binary is in `/usr/libexec/zabbix/` and has execute permissions

**Problem**: Zabbix shows "Not supported" for the item
- **Solution**: Check if Zabbix Agent 2 is running: `sudo systemctl status zabbix-agent2`
- Verify the configuration file exists in `/etc/zabbix/zabbix_agent2.d/apt-updates.conf`
- Ensure the plugin binary has proper permissions for the zabbix user

**Problem**: Plugin returns error about apt command
- **Solution**: Run `sudo apt update` first to refresh package lists
- Ensure you're running as root or the Zabbix agent user has permissions

**Problem**: No updates detected but `apt upgrade` shows packages
- **Solution**: The plugin uses cached data. Run `sudo apt update` first.

### Uninstalling

To remove the plugin:

```bash
# Remove the binary
sudo rm /usr/libexec/zabbix/zabbix-agent2-plugin-apt-updates

# Remove the configuration file
sudo rm /etc/zabbix/zabbix_agent2.d/apt-updates.conf

# Restart Zabbix Agent
sudo systemctl restart zabbix-agent2
```

## Zabbix Template

A pre-configured Zabbix template is available for easy integration. See [templates/7.4/README.md](templates/7.4/README.md) for comprehensive template documentation including:

- Template features and capabilities
- Complete list of all items with JSONPath expressions
- Trigger recommendations (template includes only items, not pre-configured triggers)
- Import instructions and best practices
- Troubleshooting guide

The template file is located at `templates/7.4/apt_updates_zabbix_agent2.yaml`.

## Project Structure

```
zabbix_agent2-apt-updates/
├── src/                     # Source code directory
│   ├── main.go              # Main plugin entry point
│   └── plugin/              # Official Zabbix Go plugin implementation
│       ├── config.go        # Configuration management
│       ├── handlers/        # Metric collection logic
│       │   └── handlers.go
│       ├── params/          # Parameter definitions
│       │   └── params.go
│       └── plugin.go        # Plugin registration and entry point
├── go.mod                  # Go module definition
├── go.sum                  # Go dependencies checksums
├── apt-updates.conf        # Plugin configuration file
├── dist/                   # Pre-built binaries (created during build)
│   ├── zabbix-agent2-plugin-apt-updates-linux-amd64
│   ├── zabbix-agent2-plugin-apt-updates-linux-arm64
│   └── zabbix-agent2-plugin-apt-updates-linux-armv7
├── templates/              # Zabbix templates
│   └── 7.4/
│       └── apt_updates_zabbix_agent2.yaml
├── README.md               # This file
├── CHANGELOG.md            # Version history
└── .gitignore              # Files to ignore in version control
```

## Build Instructions

### Prerequisites
- Go 1.21 or later (for native builds)
- Git
- Docker & Docker Compose (optional, for containerized builds)

### Building with Docker (Recommended)

The easiest way to build the plugin is using Docker:

```bash
# Clone the repository
git clone https://github.com/Aecroo/zabbix_agent2-apt-updates.git
cd zabbix-agent2-apt-updates

# Build for all platforms using Docker
docker compose up builder

# Artifacts will be in the dist/ directory
ls -lh dist/
```

### Building Natively

If you have Go installed, you can build natively:

```bash
# Clone the repository
git clone https://github.com/Aecroo/zabbix_agent2-apt-updates.git
cd zabbix-agent2-apt-updates

# Build for current platform (Linux AMD64)
make build

# Artifacts will be in the dist/ directory
```

### Cross-compilation

To build for different platforms:

```bash
# Linux AMD64 (default)
make GOOS=linux GOARCH=amd64 build

# Linux ARM64
make GOOS=linux GOARCH=arm64 build

# Linux ARMv7
make GOOS=linux GOARCH=arm GOARM=7 build
```

## Deployment

### Method 1: Direct Execution (for testing)

```bash
/usr/libexec/zabbix/zabbix-agent2-plugin-apt-updates --help
```

### Method 2: Integration with Zabbix Agent 2

The plugin is automatically detected by Zabbix Agent 2 when placed in the correct location:

1. Place the binary in the Zabbix plugin directory:
   ```bash
   sudo cp packages/zabbix-agent2-plugin-apt-updates-linux-amd64 \
       /usr/libexec/zabbix/zabbix-agent2-plugin-apt-updates
   sudo chmod +x /usr/libexec/zabbix/zabbix-agent2-plugin-apt-updates
   ```

2. Configure Zabbix Agent to use the plugin:

   Copy the configuration file:
   ```bash
   sudo cp apt-updates.conf /etc/zabbix/zabbix_agent2.d/apt-updates.conf
   ```

3. Restart Zabbix Agent:
   ```bash
   sudo systemctl restart zabbix-agent2
   ```

## Usage

### Command Line

```bash
# Get version information
/usr/libexec/zabbix/zabbix-agent2-plugin-apt-updates --version

# Get help
/usr/libexec/zabbix/zabbix-agent2-plugin-apt-updates --help
```

### Zabbix Items

Create the following items in your Zabbix template:

| Item Key | Type | Description |
|----------|------|-------------|
| `updates.get` | Zabbix Agent (active) | Returns comprehensive JSON with all update information |

## Configuration

The plugin requires minimal configuration. The only required setting is the path to the plugin executable.

### Timeout Configuration

**Important:** With Zabbix Agent 2 version 7.0 and later, timeout can be configured directly in the item settings with a range of 1-600 seconds (10 minutes). This provides more granular control over timeouts for different monitoring items.

For older versions or as a fallback when not specified at the item level, you can configure a plugin-level timeout:

Example configuration (in `/etc/zabbix/zabbix_agent2.d/apt-updates.conf`):
```ini
# APT Updates Plugin Configuration
# Path to plugin executable
Plugins.APTUpdates.System.Path=/usr/libexec/zabbix/zabbix-agent2-plugin-apt-updates
# Optional: Plugin-level timeout (used as fallback for older Zabbix versions)
# Plugins.APTUpdates.Timeout=30
```

## Testing

Run unit tests:

```bash
go test -v ./...
```

### Mock Testing

The plugin includes mock testing for various scenarios:
- No updates available
- Multiple updates available
- Large update lists
- Error conditions (missing apt command)

## Requirements

- Debian or Ubuntu system with APT package manager
- `apt` command must be in PATH
- Root or sudo privileges may be required depending on configuration

## Troubleshooting

### Common Issues

1. **Permission denied when running apt**:
   - Ensure the Zabbix agent user has permission to run `apt list --upgradable`
   - You may need to configure sudoers or run as root

2. **No updates detected but apt shows updates**:
   - Verify the plugin is using the correct APT command path
   - Check for caching issues (run `sudo apt update` first)

3. **Binary not found by Zabbix Agent**:
   - Ensure the binary path is in the agent's configuration
   - Verify file permissions: `chmod +x /path/to/binary`

## License

This project is licensed under the GPL-2.0 license, consistent with Zabbix Agent 2 licensing.

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Open a pull request

## Support

For issues and questions, please open an issue in the project repository.

## Docker Deployment

The project includes Docker support for easy building and deployment.

### Quick Start with Docker Compose

```bash
# Clone the repository
git clone https://github.com/Aecroo/zabbix_agent2-apt-updates.git
cd zabbix-agent2-apt-updates

# Build and start the Zabbix Agent with the plugin
docker compose up -d agent
```

### Using the Builder Image

To build the plugin for multiple platforms:

```bash
# Build all platform binaries
docker compose up builder

# Artifacts will be available in dist/
ls -lh dist/
```

### Customizing the Deployment

Edit `docker-compose.yml` to customize:
- Zabbix server connection (`ZBX_SERVER_HOST`)
- Hostname reported to Zabbix (`ZBX_HOSTNAME`)
- Warning threshold (`ZBX_UPDATES_THRESHOLD_WARNING`)
- Debug mode (`ZBX_DEBUG=true`)

### Build Script

A convenient build script is provided:

```bash
# Show help
./build.sh help

# Build using Docker
./build.sh build-docker

# Deploy with Docker Compose
./build.sh deploy

# Start/Stop the container
./build.sh start
./build.sh stop

# View logs
./build.sh logs
```

### Manual Docker Build

You can also build and run manually:

```bash
# Build the runtime image
docker build -t zabbix-apt-updates -f Dockerfile .

# Run the container
docker run -d \
  --name zabbix-agent-apt \
  -p 10050:10050 \
  -v /var/lib/apt:/var/lib/apt:ro \
  -v /etc/apt:/etc/apt:ro \
  -e ZBX_SERVER_HOST=your-zabbix-server \
  -e ZBX_HOSTNAME=apt-monitor \
  zabbix-apt-updates
```
