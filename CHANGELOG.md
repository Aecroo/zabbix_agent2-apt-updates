# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
