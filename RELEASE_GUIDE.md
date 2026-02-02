# Zabbix Agent 2 APT Updates - Release Guide

This guide provides step-by-step instructions for creating a new release of the zabbix_agent2-apt-updates plugin.

## Table of Contents
- [Prerequisites](#prerequisites)
- [Version Updates](#version-updates)
- [Documentation Updates](#documentation-updates)
- [Building Binaries](#building-binaries)
- [Creating GitHub Release](#creating-github-release)
- [Committing Changes](#committing-changes)
- [Example Full Workflow](#example-full-workflow)

## Prerequisites

Before starting a new release, ensure you have:
- Docker installed and running
- `tea` CLI configured for GitHub access (http://192.168.0.23:3000)
- Git access with write permissions
- Go 1.24+ (optional, for native builds)

## Version Updates

Update the version number in `main.go`:

```bash
# Example: Changing from v0.5.1 to v0.6.0
nano main.go

# Find and update these lines:
var (
	PLUGIN_VERSION_MAJOR = 0
	PLUGIN_VERSION_MINOR = 6          # Changed from 5
	PLUGIN_VERSION_PATCH = 0          # Changed from 1
	PLUGIN_VERSION_RC    = ""
	PLUGIN_LICENSE_YEAR  = 2026
)
```

## Documentation Updates

Update `README.md` to reflect the new version:

```bash
# Update version badge
sed -i 's/version-0.5.1/version-0.6.0/' README.md

# Update download URLs in installation instructions
sed -i 's/download\/v0.5.1/download\/v0.6.0/' README.md

# Update upgrade examples
sed -i 's/releases\/download\/v0.5.1/releases\/download\/v0.6.0/' README.md
```

Update `CHANGELOG.md`:

```bash
# Add new version section at the top
cat >> CHANGELOG.md << 'EOF'
## [Unreleased]

## [X.Y.Z] - YYYY-MM-DD

### Added
- Feature 1 description
- Feature 2 description

### Changed
- Improvement 1 description
- Improvement 2 description

### Fixed
- Bug #N: Description of fix
- Bug #M: Description of fix

### Removed
- Deprecated feature name (if applicable)
EOF
```

## Building Binaries

Use Docker to build binaries for all platforms:

```bash
# Clean previous builds
./build.sh clean

# Build for all platforms (linux-amd64, linux-arm64, linux-armv7)
./build.sh build-docker

# Verify binaries were created
ls -lh dist/

# Expected output:
# total 18M
# -rwxr-xr-x 1 root root 5.9M Feb  2 12:05 zabbix-agent2-plugin-apt-updates-linux-amd64
# -rwxr-xr-x 1 root root 5.8M Feb  2 12:05 zabbix-agent2-plugin-apt-updates-linux-arm64
# -rwxr-xr-x 1 root root 5.6M Feb  2 12:05 zabbix-agent2-plugin-apt-updates-linux-armv7
```

## Creating GitHub Release

Use `tea` CLI to create and manage releases:

### 1. Create a new release (draft mode)

```bash
# Create draft release with tag and title
tea releases create v0.6.0 --title "v0.6.0 - [Brief Description]"

# Example:
tea releases create v0.6.0 --title "v0.6.0 - ARM Platform Fix"
```

### 2. Add release notes

```bash
tear releases edit v0.6.0 --note "# Release Notes

## Overview
Brief description of what this release accomplishes.

## What's Fixed
- Bug #8: Detailed description
- Bug #7: Detailed description

## Testing
All tests pass on all platforms."
```

### 3. Upload binaries

```bash
# Upload all built binaries at once
tea releases assets create v0.6.0 dist/zabbix-agent2-plugin-apt-updates-linux-amd64 dist/zabbix-agent2-plugin-apt-updates-linux-arm64 dist/zabbix-agent2-plugin-apt-updates-linux-armv7
```

### 4. Publish the release

```bash
# Change draft status to false to publish
tea releases edit v0.6.0 --draft "false"
```

## Committing Changes

Commit all version-related changes:

```bash
# Stage the changed files
git add main.go README.md CHANGELOG.md

# Create commit with version update
git commit -m "Update version to v0.6.0"

# Push to master branch
git push origin master
```

## Example Full Workflow

Here's a complete example workflow for creating release v0.7.0:

```bash
#!/bin/bash

# 1. Update version
nano main.go
# Change PLUGIN_VERSION_MINOR from 6 to 7, PATCH from 0 to 0

# 2. Update documentation
sed -i 's/version-0.6.0/version-0.7.0/' README.md
sed -i 's/download\/v0.6.0/download\/v0.7.0/' README.md

cat >> CHANGELOG.md << 'EOF'
## [Unreleased]

## [0.7.0] - 2026-03-01

### Added
- New feature: Support for Debian Bookworm
- Enhanced logging for debugging

### Fixed
- Issue #9: Fix timeout on systems with large update lists
- Issue #10: Handle apt-cache policy errors gracefully
EOF

# 3. Build binaries
./build.sh clean
./build.sh build-docker
ls -lh dist/

# 4. Create GitHub release
tear releases create v0.7.0 --title "v0.7.0 - Debian Bookworm Support"
tear releases edit v0.7.0 --note "# v0.7.0 Release Notes

## Overview
Adds official support for Debian Bookworm and improved error handling.

## What's New
- Debian Bookworm compatibility
- Better timeout handling
- Enhanced debugging logs

## Testing
Tested on Ubuntu 22.04, 24.04 and Debian 11, 12 (Bookworm)."
tea releases assets create v0.7.0 dist/zabbix-agent2-plugin-apt-updates-linux-amd64 dist/zabbix-agent2-plugin-apt-updates-linux-arm64 dist/zabbix-agent2-plugin-apt-updates-linux-armv7
tea releases edit v0.7.0 --draft "false"

# 5. Commit and push
git add main.go README.md CHANGELOG.md
git commit -m "Update version to v0.7.0"
git push origin master

echo "Release v0.7.0 completed successfully!"
```

## Troubleshooting

### Docker build fails
- Ensure Docker daemon is running: `systemctl status docker`
- Check disk space: `df -h`
- Clean previous builds first: `./build.sh clean`

### Tea CLI authentication issues
- Configure tea login: `tea login add`
- Verify credentials: `tea whoami`

### Git push rejected
- Pull latest changes first: `git pull origin master`
- Resolve merge conflicts if any exist
- Ensure you have write permissions to the repository

## Verification Checklist

Before marking a release complete, verify:

- [ ] Version updated in main.go ✓
- [ ] README.md references new version ✓
- [ ] CHANGELOG.md has new section ✓
- [ ] Binaries built for all three platforms ✓
- [ ] GitHub release created with proper tag ✓
- [ ] Release notes added to release ✓
- [ ] All three binaries uploaded ✓
- [ ] Release published (not draft) ✓
- [ ] Changes committed and pushed ✓

## Release Management Tips

1. **Semantic Versioning**: Follow SemVer (MAJOR.MINOR.PATCH)
   - MAJOR: Breaking changes
   - MINOR: Backwards-compatible new features
   - PATCH: Backwards-compatible bug fixes

2. **Release Frequency**: Aim for monthly or quarterly releases with meaningful changes

3. **Testing**: Always test binaries on at least one system before publishing:
   ```bash
   # Test the binary
   ./dist/zabbix-agent2-plugin-apt-updates-linux-amd64 --version
   ```

4. **Communication**: After release, consider posting in relevant channels about new features/fixes

5. **Backup**: Keep old release binaries for at least 1 year in case users need to downgrade
