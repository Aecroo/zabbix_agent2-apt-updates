# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- **Issue #5 - Duplicate item values**: Simplified the plugin to use a single item key `updates.get` that returns comprehensive JSON data with all update types (security, recommended, optional, all). Users can now extract specific values using Zabbix JSONPath preprocessing instead of having multiple item keys that returned duplicate data.

### Changed
- **Metric System**: Replaced multiple metrics (`apt.updates`, `apt.updates.list`, `apt.updates.details`) with a single unified metric (`updates.get`) that returns comprehensive JSON data
- **API Design**: Simplified from multiple item keys with bracket notation to one item key with JSONPath-based value extraction
- **Handler Functions**: Removed individual handlers for count/list/details and consolidated into `GetAllUpdates` handler

## [0.4.1] - 2026-02-01

### Fixed
- **ARM builds**: Rebuilt all binaries (amd64, arm64, and armv7) to fix issue #3. ARM binaries were built from outdated code that still contained the removed WarningThreshold feature, causing JSON unmarshal errors.

## [0.4.0] - 2026-02-01

### Removed
- **WarningThreshold feature**: Completely removed WarningThreshold from the plugin as requested in issue #2. Zabbix handles warnings through triggers, making this built-in functionality redundant.
  - Removed WarningThreshold field from session struct
  - Removed WarningThreshold parameter definition
  - Removed WarningThreshold and IsAboveWarning fields from CheckResult
  - Removed all threshold-related logic from handlers
  - Updated configuration file to remove related options
  - Updated documentation to reflect simplified configuration

### Changed
- **Configuration**: Simplified plugin configuration with fewer required parameters
- **CheckResult structure**: Now only contains AvailableUpdates and PackageDetailsList fields
- **Handler functions**: Streamlined to focus on core update detection without threshold calculations

## [0.3.0] - 2026-02-01

### Added
- Support for bracket notation in metric keys: `apt.updates[all]`, `apt.updates[security]`, etc.
- Three dedicated metrics:
  - `apt.updates` - Returns count of available updates by type
  - `apt.updates.list` - Returns JSON list of package names by type
  - `apt.updates.details` - Returns detailed JSON with versions by type
- Automatic update type extraction from bracket notation in handlers

### Changed
- **Metric System**: Complete rewrite to use Zabbix SDK's metric parameter system
- **Configuration**: WarningThreshold now uses string type instead of integer for compatibility
- **Build System**: Updated Docker builder with Go 1.24 and improved SDK handling
- **Error Handling**: Enhanced validation and error messages for configuration

### Fixed
- Type mismatch in WarningThreshold parameter (int → string)
- JSON unmarshal errors from default values
- Metric registration to properly handle bracket notation
- Configuration initialization preventing nil pointer dereference

### Removed
- Default value for WarningThreshold to avoid JSON marshaling conflicts

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
