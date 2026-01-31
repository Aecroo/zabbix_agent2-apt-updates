# Release Notes - Version 0.1.0

## 🎉 Version 0.1.0 Released!

The Zabbix Agent 2 APT Updates plugin is now ready for production use.

## What's Included in This Release

### 📦 Pre-built Binaries
Three platform-specific binaries are available in the `dist/` directory:
- **zabbix-apt-updates-linux-amd64** - For 64-bit x86 systems (most desktops and servers)
- **zabbix-apt-updates-linux-arm64** - For ARM 64-bit systems (Raspberry Pi 3/4, cloud ARM instances)
- **zabbix-apt-updates-linux-armv7** - For ARM 32-bit systems (older Raspberry Pi models)

All binaries are statically linked for maximum compatibility.

### 📚 Documentation Updates

#### Enhanced README.md
The documentation now includes:

1. **Normal User Guide** - Step-by-step installation instructions for Ubuntu/Debian systems
   - Download and install pre-built binaries
   - Configure Zabbix Agent 2
   - Test the plugin
   - Set up monitoring in Zabbix web interface
   - Create triggers for alerts
   - Update and uninstall procedures

2. **Troubleshooting Section** - Common issues and solutions for end users

3. **Version Badge** - Clear indication of current release version

#### Updated CHANGELOG.md
- Formalized changelog following Keep a Changelog format
- Detailed list of features, changes, and fixes in v0.1.0
- Semantic versioning compliance

### 🔧 Features

#### Core Functionality
- Detects available package updates using `apt list --upgradable`
- Returns JSON-formatted results with:
  - Count of available updates
  - Detailed package information (name, target version)
  - Warning threshold indicator
  - Boolean flag for exceeding warning threshold

#### Configuration Options
- **ZBX_UPDATES_THRESHOLD_WARNING** - Set warning threshold (default: 10)
- **ZBX_DEBUG** - Enable debug logging (default: false)

#### Multi-platform Support
- Auto-detects APT (Debian/Ubuntu) or DNF (RHEL/CentOS/Fedora)
- Cross-compiled for multiple architectures

### 📖 Usage Examples

#### Basic Check
```bash
/usr/local/bin/zabbix-plugins/zabbix-apt-updates check
```

Example Output:
```json
{
  "available_updates": 5,
  "package_details_list": [
    {"name": "curl", "target_version": "7.81.0-1+b2"},
    {"name": "nginx", "target_version": "1.18.0-6ubuntu14.3"}
  ],
  "warning_threshold": 10,
  "is_above_warning": false
}
```

#### With Configuration
```bash
export ZBX_UPDATES_THRESHOLD_WARNING=5
export ZBX_DEBUG=true
/usr/local/bin/zabbix-plugins/zabbix-apt-updates check
```

### 🛠️ Installation Quick Start (Ubuntu/Debian)

```bash
# 1. Download and install
sudo mkdir -p /usr/local/bin/zabbix-plugins
wget http://192.168.0.23:3000/zbx/zabbix_agent2-apt-updates/-/raw/master/dist/zabbix-apt-updates-linux-amd64 \
  -O /usr/local/bin/zabbix-plugins/zabbix-apt-updates
sudo chmod +x /usr/local/bin/zabbix-plugins/zabbix-apt-updates

# 2. Configure Zabbix Agent 2
sudo mkdir -p /etc/zabbix/zabbix_agent2.d/
sudo nano /etc/zabbix/zabbix_agent2.d/userparameter_apt.conf

# Add this line:
UserParameter=apt.updates[check],/usr/local/bin/zabbix-plugins/zabbix-apt-updates check

# 3. Restart Zabbix Agent
sudo systemctl restart zabbix-agent2
```

### 🎯 Zabbix Monitoring Setup

#### Item Configuration
Create an item with:
- **Type**: Zabbix Agent (active)
- **Key**: `apt.updates[check]`
- **Type of information**: Text

#### Preprocessing for Count
To extract just the count:
1. Type: Regular expression
2. Pattern: `"available_updates": ([0-9]*)`
3. Result: `\1`

#### Trigger Example
```
Trigger name: "Many APT updates available"
Expression: {template_name:apt.updates[check].str(Warning).regexp("true")}=1
Severity: Warning
```

### 📦 Build Information

- **Build System**: Docker-based cross-compilation
- **Go Version**: 1.21+
- **Static Linking**: Yes (for maximum compatibility)
- **Debug Symbols**: Included in binaries

### 🔒 License

This project is licensed under the GPL-2.0 license, consistent with Zabbix Agent 2 licensing.

### 🐛 Known Limitations

1. Requires root or sudo privileges to run `apt list --upgradable`
2. Uses cached APT data (run `sudo apt update` first if package lists are stale)
3. No support for APT proxy configuration in the plugin itself

### 📞 Support

For issues and questions:
- Open an issue in the project repository
- Check the Troubleshooting section in README.md
- Review the CHANGELOG.md for recent changes

### 🚀 Future Enhancements (Planned)

- Support for apt-get proxy configuration
- Caching mechanism to reduce APT command execution
- More detailed version information (current vs target)
- Additional package managers (yum, zypper)
- Configuration file support instead of environment variables only

## 📊 Release Statistics

- **Lines of Code**: ~250 lines in main.go
- **Test Coverage**: Unit tests included
- **Platforms Supported**: 3 (AMD64, ARM64, ARMv7)
- **Documentation Pages**: 8 major sections in README.md

## 🎯 Target Audience

This release is ideal for:
- System administrators monitoring Ubuntu/Debuntu servers
- DevOps teams integrating with Zabbix monitoring
- Security teams tracking unapplied updates
- Anyone needing automated update notifications

## ✅ Quality Assurance

All binaries have been tested on:
- Build verification (compiles without errors)
- Static linking verification (no external dependencies)
- Platform compatibility checks
- Documentation accuracy review

---

**Release Date**: 2026-01-31
**Version**: 0.1.0
**Status**: Production Ready
**Git Tag**: v0.1.0
