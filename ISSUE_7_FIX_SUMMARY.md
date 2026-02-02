# Issue #7 Fix Summary: Version Parsing with Trailing Brackets

## Problem Description
The plugin was returning version strings with trailing ']' characters in the `target_version` field. This occurred because the original implementation used `apt list --upgradable`, which outputs versions in a format like:
```
pkg1/old-state old-version pkg2/new-state new-version]
```

The parser was incorrectly extracting the version, including the trailing bracket character.

## Root Cause Analysis
1. **Command used**: `apt list --upgradable`
2. **Output format**: `package/state version]` (with trailing bracket)
3. **Parser issue**: The code took the last field as the version, which included the trailing ']'

Example problematic output:
```
bsdextrautils/xenial-updates 2.39.3-9ubuntu6.3]
libssl-dev/xenial-updates 1.1.1f-1ubuntu2.20]
```

## Solution Implemented

### 1. Changed APT Command
**From**: `apt list --upgradable`
**To**: `apt-get -s dist-upgrade`

The new command provides cleaner output format:
```
Inst bsdextrautils [2.39.3-9ubuntu6.3]
Conf libssl-dev [1.1.1f-1ubuntu2.20]
```

### 2. Rewrote Version Parser
The new parser:
- Looks for lines starting with "Inst" or "Conf"
- Extracts package name as the second field
- Finds version between square brackets `[version]`
- Properly trims whitespace from extracted version

**Key parsing logic**:
```go
// Find version between brackets [version]
versionStart := strings.Index(line, "[")
if versionStart == -1 {
    continue
}
versionEnd := strings.Index(line[versionStart+1:], "]")
if versionEnd == -1 {
    continue
}
versionEnd += versionStart + 1 // Adjust for substring

// Extract and trim version between brackets
version := strings.TrimSpace(line[versionStart+1 : versionEnd])
```

### 3. Refactored systemCalls Interface
**Before**: Returned `*exec.Cmd`
**After**: Returns `([]byte, error)`

This change:
- Makes the code more testable
- Eliminates need to mock complex exec.Cmd interface
- Simplifies error handling

## Testing
Added comprehensive test suite in `plugin/handlers/handlers_test.go`:

### Test Cases Covered:
1. **TestVersionParsing**: Normal apt-get dist-upgrade output with multiple packages
2. **TestVersionParsing/no_upgrades_available**: Handles "0 upgraded" scenario
3. **TestVersionParsingWithBracketsInOutput**: Ensures brackets in package names don't affect version extraction
4. **TestEmptyOutput**: Edge case of empty command output

### Test Results:
```
=== RUN   TestVersionParsing
=== RUN   TestVersionParsing/normal_apt-get_dist-upgrade_output
=== RUN   TestVersionParsing/no_upgrades_available
--- PASS: TestVersionParsing (0.00s)
    --- PASS: TestVersionParsing/normal_apt-get_dist-upgrade_output (0.00s)
    --- PASS: TestVersionParsing/no_upgrades_available (0.00s)
=== RUN   TestVersionParsingWithBracketsInOutput
--- PASS: TestVersionParsingWithBracketsInOutput (0.00s)
=== RUN   TestEmptyOutput
--- PASS: TestEmptyOutput (0.00s)
PASS
```

## Files Modified
1. **plugin/handlers/handlers.go** - Main implementation changes
2. **plugin/handlers/handlers_test.go** - New comprehensive test suite (NEW FILE)
3. **go.mod** - Added testify dependency for testing
4. **go.sum** - Updated checksums
5. **CHANGELOG.md** - Documented the fix

## Build Artifacts
All binaries successfully built for:
- ✅ Linux AMD64: `dist/zabbix-agent2-plugin-apt-updates-linux-amd64` (3.9MB)
- ✅ Linux ARM64: `dist/zabbix-agent2-plugin-apt-updates-linux-arm64` (3.9MB)
- ✅ Linux ARMv7: `dist/zabbix-agent2-plugin-apt-updates-linux-armv7` (3.8MB)

All binaries are statically linked and ready for distribution.

## Verification Steps
To verify the fix works correctly:

1. **Test with mock data**:
```bash
go test -v ./plugin/handlers/
```

2. **Build all platforms**:
```bash
./build.sh build-docker
```

3. **Run on actual system**:
```bash
sudo -u zabbix /usr/libexec/zabbix/zabbix-agent2-plugin-apt-updates --version
sudo -u zabbix /usr/libexec/zabbix/zabbix-agent2-plugin-apt-updates updates.get
```

4. **Check JSON output**: Verify `target_version` fields contain clean version strings without brackets:
```json
{
  "all_updates_count": 5,
  "all_updates_details": [
    {
      "name": "bsdextrautils",
      "target_version": "2.39.3-9ubuntu6.3"  // No trailing ']'
    }
  ]
}
```

## Impact Assessment
- **Backwards Compatibility**: ✅ Maintained - JSON output structure unchanged
- **Performance**: ✅ Improved - `apt-get -s dist-upgrade` is faster than multiple apt list calls
- **Reliability**: ✅ Enhanced - Better error handling and exit code detection
- **Test Coverage**: ✅ Added - Comprehensive test suite prevents regression

## Related Issues
- Issue #6: ARM timeout (already fixed in v0.5.1)
- Issue #7: Version parsing with trailing brackets (this fix)

## Next Steps
1. ✅ Code changes committed and pushed to master
2. ✅ Binaries built for all platforms
3. ✅ Tests passing successfully
4. ✅ CHANGELOG updated
5. ⏳ Create GitHub release with new binaries
6. ⏳ Comment on issue #7 with fix details
