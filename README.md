# Zabbix Agent 2 APT Updates Plugin

[![Docker](https://img.shields.io/badge/Docker-Supported-blue)](docker-compose.yml)

A monitoring plugin for Zabbix Agent 2 that checks available package updates on Debian/Ubuntu systems using APT.

## Overview

This plugin detects available system updates by executing `apt list --upgradable` and returns the count of available updates in a format compatible with Zabbix Agent 2.

## Project Structure

```
zabbix_agent2-apt-updates/
├── main.go                 # Main entry point
├── apt_updates_check.go    # APT update detection logic
├── go.mod                  # Go module definition
├── go.sum                  # Go dependencies checksums
├── README.md               # This file
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
git clone http://192.168.0.23:3000/zbx/zabbix_agent2-apt-updates.git
cd zabbix-agent2-apt-updates

# Build for all platforms using Docker
docker-compose up builder

# Artifacts will be in the dist/ directory
ls -lh dist/
```

### Building Natively

If you have Go installed, you can build natively:

```bash
# Clone the repository
git clone http://192.168.0.23:3000/zbx/zabbix_agent2-apt-updates.git
cd zabbix-agent2-apt-updates

# Initialize Go module (if not already done)
go mod init github.com/netdata/zabbix-agent-apt-updates

# Build the binary
go build -o dist/zabbix-apt-updates
```

### Cross-compilation
### Cross-compilation

To build for different platforms:

```bash
# Linux AMD64 (default)
GOOS=linux GOARCH=amd64 go build -o dist/zabbix-apt-updates-linux-amd64

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o dist/zabbix-apt-updates-linux-arm64
```

## Deployment

### Method 1: Direct Execution (for testing)

```bash
./dist/zabbix-apt-updates --help
```

### Method 2: Integration with Zabbix Agent 2

1. Place the binary in a location accessible by Zabbix Agent:
   ```bash
   sudo cp dist/zabbix-apt-updates /usr/local/bin/
   sudo chmod +x /usr/local/bin/zabbix-apt-updates
   ```

2. Configure Zabbix Agent to use the plugin:

   Edit `/etc/zabbix/zabbix_agent2.d/userparameter_apt.conf`:
   ```ini
   # Check for available APT updates
   UserParameter=apt.updates[*],/usr/local/bin/zabbix-apt-updates check $1
   ```

3. Restart Zabbix Agent:
   ```bash
   sudo systemctl restart zabbix-agent2
   ```

## Usage

### Command Line

```bash
# Get available updates count
./zabbix-apt-updates check

# Example output (JSON format):
# {"available_updates": 5, "package_details_list": [{"name":"curl","current_version":"7.81.0-1","target":"7.81.0-1+b2"}]}
```

### Zabbix Items

Create the following items in your Zabbix template:

| Item Key | Type | Description |
|----------|------|-------------|
| `apt.updates[available]` | Zabbix Agent | Returns count of available updates |
| `apt.updates[details]` | Zabbix Agent | Returns detailed JSON with package information |

## Configuration

The plugin can be configured using environment variables:

| Environment Variable | Default | Description |
|----------------------|---------|-------------|
| `ZBX_UPDATES_THRESHOLD_WARNING` | 10 | Warning threshold for number of available updates |
| `ZBX_DEBUG` | false | Enable debug logging |

Example:
```bash
export ZBX_UPDATES_THRESHOLD_WARNING=5
export ZBX_DEBUG=true
./zabbix-apt-updates check
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
git clone http://192.168.0.23:3000/zbx/zabbix_agent2-apt-updates.git
cd zabbix-agent2-apt-updates

# Build and start the Zabbix Agent with the plugin
docker-compose up -d agent
```

### Using the Builder Image

To build the plugin for multiple platforms:

```bash
# Build all platform binaries
docker-compose up builder

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
