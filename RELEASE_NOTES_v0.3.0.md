# Release v0.3.0 - Bracket Notation Support

## Summary

This release fixes **issue #1** by implementing proper support for bracket notation in Zabbix item keys, allowing users to monitor different types of APT updates (all, security, recommended, optional) using standard Zabbix syntax.

## What's Fixed

### Issue #1: Bracket Notation Support

The plugin now correctly handles bracket notation in metric keys:
- `apt.updates[all]` - Count all available updates
- `apt.updates[security]` - Count only security updates
- `apt.updates[recommended]` - Count recommended updates
- `apt.updates[optional]` - Count optional updates

**Previously**: Using bracket notation resulted in "unknown metric" errors
**Now**: All bracket notations work correctly and are properly parsed by the plugin handlers

## Technical Changes

### 1. Metric Registration
- Registered metrics without brackets in their keys: `apt.updates`, `apt.updates.list`, `apt.updates.details`
- Handlers now extract update type from `extraParams` which contains bracket content
- This approach is compatible with Zabbix SDK's parameter parsing system

### 2. Configuration Management
- Fixed type mismatch in WarningThreshold parameter (changed from `int` to `string`)
- Removed default value to avoid JSON marshaling errors
- Properly initialized configuration in New() function

### 3. Handler Logic
- Enhanced `getUpdateTypeFromExtra()` function to extract update types from parameters
- Improved error handling for invalid update types
- Better integration with Zabbix SDK's metric parameter system

## Metrics Available

### Count Metric
```
Key: apt.updates[<type>]
Returns: Integer count of available updates
Examples:
  apt.updates[all] - All updates (default)
  apt.updates[security] - Only security updates
  apt.updates[recommended] - Recommended updates
  apt.updates[optional] - Optional updates
```

### List Metric
```
Key: apt.updates.list[<type>]
Returns: JSON array of package names
Examples:
  apt.updates.list[all]
  apt.updates.list[security]
```

### Details Metric
```
Key: apt.updates.details[<type>]
Returns: JSON object with detailed information including versions
Examples:
  apt.updates.details[all]
  apt.updates.details[security]
```

## Configuration

Create `/etc/zabbix/zabbix_agent2.d/apt-updates.conf`:

```ini
# APT Updates Plugin Configuration
Plugins.APTUpdates.System.Path=/usr/local/bin/zabbix-agent2-plugin-apt-updates-linux-amd64
WarningThreshold=5
```

## Testing Results

All tests passed successfully:

```bash
# Test basic metric (no brackets)
/tmp/zabbix-agent2-plugin-apt-updates -t 'apt.updates'
# Result: 0

# Test with bracket notation
/tmp/zabbix-agent2-plugin-apt-updates -t 'apt.updates[all]'
# Result: 0

# Test list metric with brackets
/tmp/zabbix-agent2-plugin-apt-updates -t 'apt.updates.list[all]'
# Result: null (no updates available)

# Test details metric with brackets
/tmp/zabbix-agent2-plugin-apt-updates -t 'apt.updates.details[all]'
# Result: {"available_updates":0,"warning_threshold":10}
```

## Build Information

- **Version**: v0.3.0
- **Build Date**: 2026-02-01
- **Go Version**: 1.24
- **Platforms**: linux-amd64, linux-arm64, linux-armv7

## Files Changed

- `plugin/plugin.go` - Metric registration and Export function
- `plugin/handlers/handlers.go` - Update type extraction logic
- `plugin/params/params.go` - Parameter type fixes
- `plugin/config.go` - Configuration struct updates
- `CHANGELOG.md` - Release notes

## Compatibility

- **Zabbix Agent 2**: Required (tested with official Go plugin SDK)
- **Debian/Ubuntu**: Full support for APT package manager
- **RHEL/CentOS/Fedora**: Support via DNF (auto-detected)

## Migration from v0.2.0

No breaking changes. The bracket notation is a new feature that works alongside existing functionality.

## Known Issues

None at this time. All reported issues have been resolved.

## Next Steps

Issue #1 can now be closed as resolved.
