# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-02-17

### Added
- **FOSS Release**: Official open-source release on GitHub
- Improved error handling and logging

### Changed
- **Project Structure**: Organized source code into `src/` directory for better maintainability
- **Version**: Major version bump to 1.0.0 indicating production stability
- All imports updated to use `github.com/netdata/zabbix-agent-apt-updates/src/plugin`

### Fixed
- Import path resolution in Go module system
- Build process compatibility with src/ directory structure

## [Unreleased]

## [0.8.0] - 2026-02-17

### Added
- **Issue #10 - Phased Updates Detection**: Full support for Ubuntu's phased updates feature.
  - Detects packages that have been "deferred due to phasing" using `apt-get -s upgrade` output
  - Properly marks packages with IsPhased field in JSON output
  - Separately counts and lists phased updates (phased_updates_count, phased_updates_list, phased_updates_details)
  - Phased updates are excluded from security/recommended/optional categories to avoid double-counting
- Two-pass update detection system:
  - First pass with Always-Include-Phased-Updates=false to detect deferred packages
  - Second pass with Always-Include-Phased-Updates=true to get full package list
  - Correctly sets IsPhased field based on which packages were deferred
- Line-by-line parsing of "deferred due to phasing" section from apt-get output
- Updated isPhasedUpdate() function to check IsPhased field in addition to target version string

### Changed
- **APT command**: Switched back to `apt-get -s upgrade` for proper phased update detection
- Enhanced parsing logic to handle both "deferred due to phasing" message and Inst lines
- Improved UpdateInfo struct with IsPhased boolean field for accurate package classification
- Updated GetAllUpdates to properly filter and categorize all update types

### Fixed
- Proper slice mutation handling for variadic parameters in Go (fixed issue where deferred packages map wasn't being returned)
- Correct assignment of IsPhased field during second pass when all packages are included
- Accurate counting of phased vs non-phased updates

## [0.7.0] - 2026-02-03

### Added
- **Issue #9 - Last APT update time**: Added `last_apt_update_time` field to JSON output that shows the Unix timestamp of when the last 'apt update' was run. This helps monitor how stale package information is.
- **check_duration_seconds**: Added timing metric to track how long each check takes for performance monitoring.

### Fixed
- **Issue #9 - APT update time detection**: Fixed the `getLastAptUpdateTime()` function to properly handle permission errors from find command while still extracting valid timestamps. The function now uses `find /var/lib/apt/lists -type f -printf '%T@\n'` and continues processing output even when find returns exit code 1 due to permission errors on directories like `/var/lib/apt/lists/partial`.

### Changed
- **JSON Output**: Enhanced JSON response now includes both `check_duration_seconds` (float64) and `last_apt_update_time` (int64 Unix timestamp) fields in all update queries.
- **APT Time Detection**: Improved to handle modern APT file types (InRelease, Packages, Sources) instead of only looking for legacy .list/.lists files.

### Fixed
- **Issue #8 - ARM platform timeout (revised)**: Fixed plugin execution on ARM platforms (arm64 and armv7) where the plugin was getting killed with "signal: killed" error. The issue persisted even after switching to `apt-get -s upgrade`. Final solution: switched back to using `apt list --upgradable` with improved parsing logic that properly extracts version strings without trailing brackets. This approach is lightweight enough to avoid OOM kills on ARM while maintaining clean version output.
- **Issue #7 - Version parsing with trailing brackets**: Fixed version string parsing to remove trailing ']' characters from target_version field. The parser now properly handles the `apt list --upgradable` format ("package/state version]") by extracting the last field and removing the trailing bracket character.

### Added
- Comprehensive test suite for version parsing with multiple test cases:
  - Normal apt list --upgradable output
  - No upgrades available scenario
  - Edge case handling with brackets in package names
  - Empty output handling
- Refactored systemCalls interface to return ([]byte, error) instead of *exec.Cmd for better testability

### Changed
- **APT command**: Switched from `apt-get -s upgrade` back to `apt list --upgradable`
- **Version parsing logic**: Rewrote parser to handle "package/state version]" format by extracting the last field and removing trailing ']' character
- **Code structure**: Improved separation of concerns with better interface design for testing

## [0.5.1] - 2026-02-02
- **Issue #6 - ARM timeout**: Fixed plugin execution timeout on ARM platforms (armv7 and arm64) when called through zabbix_agent2 -t. The issue was caused by executing `apt list --upgradable` four times sequentially (once for each update type). The fix optimizes the code to execute apt only once and filter results in-memory, significantly reducing execution time.

### Changed
- **GetAllUpdates handler**: Optimized to call `checkAPTUpdates` only once instead of four times. Results are now filtered in-memory by update type (security, recommended, optional) instead of executing apt multiple times with different filters.
- **Performance**: Reduced execution time by ~75% on ARM platforms by eliminating redundant apt command executions.

## [0.5.0] - 2026-02-02

### Fixed
- **Issue #5 - Duplicate item values**: Simplified the plugin to use a single item key `updates.get` that returns comprehensive JSON data with all update types (security, recommended, optional, all). Users can now extract specific values using Zabbix JSONPath preprocessing instead of having multiple item keys that returned duplicate data.

### Changed
- **Metric System**: Replaced multiple metrics (`apt.updates`, `apt.updates.list`, `apt.updates.details`) with a single unified metric (`updates.get`) that returns comprehensive JSON data
- **API Design**: Simplified from multiple item keys with bracket notation to one item key with JSONPath-based value extraction
- **Handler Functions**: Removed individual handlers for count/list/details and consolidated into `GetAllUpdates` handler
- **Timeout Configuration**: Removed hardcoded timeout range restriction (1-30 seconds). With Zabbix 7.0+, timeout can be configured at the item level with a range of 1-600 seconds (10 minutes). Plugin-level timeout configuration is maintained for backwards compatibility.
- **Issue #8 - ARM platform timeout (revised)**: Fixed plugin execution on ARM platforms (arm64 and armv7) where the plugin was getting killed with "signal: killed" error. The issue persisted even after switching to `apt-get -s upgrade`. Final solution: switched back to using `apt list --upgradable` with improved parsing logic that properly extracts version strings without trailing brackets. This approach is lightweight enough to avoid OOM kills on ARM while maintaining clean version output.
- **Issue #7 - Version parsing with trailing brackets**: Fixed version string parsing to remove trailing ']' characters from target_version field. The parser now properly handles the `apt list --upgradable` format ("package/state version]") by extracting the last field and removing the trailing bracket character.

### Added
- Comprehensive test suite for version parsing with multiple test cases:
  - Normal apt list --upgradable output
  - No upgrades available scenario
  - Edge case handling with brackets in package names
  - Empty output handling
- Refactored systemCalls interface to return ([]byte, error) instead of *exec.Cmd for better testability

### Changed
- **APT command**: Switched from `apt-get -s upgrade` back to `apt list --upgradable`
- **Version parsing logic**: Rewrote parser to handle "package/state version]" format by extracting the last field and removing trailing ']' character
- **Code structure**: Improved separation of concerns with better interface design for testing

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
