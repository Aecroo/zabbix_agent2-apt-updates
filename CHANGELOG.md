# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-02-01

### Added
- Official Zabbix Agent 2 Go plugin architecture (converted from userparameter style)
- Dedicated item key: `apt.updates[<package_state>]` for precise monitoring
- Zabbix SDK integration with proper metric parameter system
- Configurable warning threshold via configuration file (`apt-updates.conf`)
- Enhanced error handling and logging integration with Zabbix Agent 2
- Plugin configuration file with official Zabbix format
- Comprehensive conversion documentation and migration guides

### Changed
- **Architecture**: Complete refactor from userparameter script to official Go plugin
- **Build System**: Updated Docker-based build system for Go plugins with GOPRIVATE support
- **Configuration**: Moved from environment variables to proper Zabbix configuration file
- **Binary Naming**: Updated to `zabbix-agent2-plugin-apt-updates-linux-*` format
- **Plugin Structure**: Organized into proper Go package structure (plugin/, plugin/config.go, etc.)

### Fixed
- Invalid SDK version in go.mod (v0.0.0-00010101000000-000000000000 → v1.2.2-0.20251205121637-3b95c058c0e4)
- Malformed go.sum file (deleted and regenerated)
- Type mismatch in params.go: WithDefault(10) → WithDefault("10") (integer → string)

### Removed
- Environment variable configuration (replaced with apt-updates.conf)
- Legacy userparameter-style execution model

## [0.1.0] - 2026-01-31

### Added
- Initial stable release of the Zabbix Agent 2 APT Updates plugin
- Support for Debian/Ubuntu systems (APT package manager)
- Auto-detection of DNF package manager (RHEL/CentOS/Fedora)
- JSON output format with detailed package information
- Configurable warning threshold via environment variable `ZBX_UPDATES_THRESHOLD_WARNING`
- Debug logging support via environment variable `ZBX_DEBUG`
- Comprehensive documentation including user guide
- Pre-built binaries for multiple platforms (AMD64, ARM64, ARMv7)
- Docker-based build and deployment system
- Example configuration files

### Changed
- Improved error handling for missing package managers
- Better parsing of APT output format
- Proper handling of apt exit code 100 (no updates available)

### Fixed
- Correct JSON formatting in all responses
- Proper exit codes for different error conditions
- Package name and version extraction from APT output

## [Unreleased]

### Added
- Initial implementation of APT updates check for Debian/Ubuntu systems
- Support for DNF package manager (RHEL/CentOS/Fedora) detection
- JSON output format with detailed package information
- Configurable warning threshold via environment variable
- Debug logging support
- Comprehensive documentation and examples

### Changed
- Initial version following Zabbix plugin conventions

### Fixed
- Proper handling of apt exit code 100 (no updates available)
- Correct parsing of APT output format
- Error handling for missing package managers

## [1.0.0] - YYYY-MM-DD

Initial release of the Zabbix Agent 2 APT Updates plugin.

### Features
- Detects available package updates using `apt list --upgradable`
- Returns JSON-formatted results with:
  - Count of available updates
  - Detailed package information (name, current version, target version)
  - Warning threshold indicator
- Auto-detection of package manager (APT or DNF)
- Environment variable configuration
- Cross-platform build support

### Usage
```bash
# Basic check
zabbix-apt-updates check

# With configuration
export ZBX_UPDATES_THRESHOLD_WARNING=5
export ZBX_DEBUG=true
zabbix-apt-updates check

# Version information
zabbix-apt-updates version
```

### Build Instructions
```bash
make build          # Build for current platform
make dist           # Create distribution package
make install        # Install to /usr/local/bin
make build-linux-arm64  # Cross-compile for ARM64
```

### Configuration

Create `/etc/zabbix/zabbix_agent2.d/userparameter_apt.conf`:

```ini
UserParameter=apt.updates[check],/usr/local/bin/zabbix-apt-updates check
```

Then restart Zabbix Agent 2:
```bash
sudo systemctl restart zabbix-agent2
```

### Zabbix Template Items

| Item Key | Type | Description |
|----------|------|-------------|
| `apt.updates[check]` | Zabbix Agent | Returns JSON with all update information |
| `apt.updates[count]` | Zabbix Agent (preprocessing) | Extracts count of available updates |
| `apt.updates[warning]` | Zabbix Agent (preprocessing) | Indicates if above warning threshold |

### Preprocessing for Count Item

To extract just the count from the JSON output:
1. Type: `Text`
2. Regular expression: `"available_updates": ([0-9]*)`
3. Output: `\(1\)`

### License

GPL-2.0
