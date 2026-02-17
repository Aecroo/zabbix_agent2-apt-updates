# Zabbix Template for APT Updates Monitoring

## Overview

This directory contains the Zabbix template (`apt_updates_zabbix_agent2.yaml`) for monitoring available package updates on Debian/Ubuntu systems using the APT package manager.

### Compatibility

- **Zabbix version**: 7.4
- **Template version**: 7.4-0
- **Plugin version**: 1.0.0+
- **Operating systems**: Ubuntu 20.04 LTS+, Debian 10+

### Template Features

The template provides comprehensive monitoring of system updates with the following capabilities:

1. **Update Classification Monitoring**
   - Security updates (critical vulnerabilities)
   - Recommended updates (bug fixes and improvements)
   - Optional updates (new features and documentation)
   - Total update count

2. **Detailed Update Information**
   - Complete package lists for each update category
   - Detailed information including version numbers and repository sources

3. **Performance Metrics**
   - Check duration in seconds
   - Last APT database update timestamp

4. **Trigger Support**
   - Items ready for trigger creation (see Triggers section)
   - Recommended trigger expressions provided

## Template Items

The template includes the following pre-configured items:

### Master Item
| Key | Type | Description |
|-----|------|-------------|
| `updates.get` | Zabbix Agent | Returns comprehensive JSON with all update information. This is the master item that collects raw data from the plugin. All other items are dependent on this master item.

**Configuration:**
- Update interval: 15 minutes (configurable)
- History storage: 1 hour
- Timeout: 10 minutes (to accommodate large update lists and slow systems)

### Dependent Items - Counts

These items extract numeric counts from the JSON response using JSONPath preprocessing:

| Key | Type | JSONPath Expression | Description |
|-----|------|-------------------|-------------|
| `updates.all_updates_count` | Dependent | `$.all_updates_count` | Total number of available updates across all categories |
| `updates.security_updates_count` | Dependent | `$.security_updates_count` | Number of security updates available |
| `updates.recommended_updates_count` | Dependent | `$.recommended_updates_count` | Number of recommended updates available |
| `updates.optional_updates_count` | Dependent | `$.optional_updates_count` | Number of optional updates available |

### Dependent Items - Lists

These items extract package lists from the JSON response:

| Key | Type | JSONPath Expression | Description |
|-----|------|-------------------|-------------|
| `updates.all_updates_list` | Dependent | `$.all_updates_list` | Array of all package names available for update |
| `updates.security_updates_list` | Dependent | `$.security_updates_list` | Array of security update package names |
| `updates.recommended_updates_list` | Dependent | `$.recommended_updates_list` | Array of recommended update package names |
| `updates.optional_updates_list` | Dependent | `$.optional_updates_list` | Array of optional update package names |

### Dependent Items - Details

These items extract detailed information including versions and repository sources:

| Key | Type | JSONPath Expression | Description |
|-----|------|-------------------|-------------|
| `updates.all_updates_details` | Dependent | `$.all_updates_details` | Array of objects with full details for all updates |
| `updates.security_updates_details` | Dependent | `$.security_updates_details` | Array of objects with full details for security updates |
| `updates.recommended_updates_details` | Dependent | `$.recommended_updates_details` | Array of objects with full details for recommended updates |
| `updates.optional_updates_details` | Dependent | `$.optional_updates_details` | Array of objects with full details for optional updates |

### Monitoring Items

These items provide operational metrics:

| Key | Type | JSONPath Expression | Description |
|-----|------|-------------------|-------------|
| `updates.last_apt_update_time` | Dependent | `$.last_apt_update_time` | Unix timestamp of the last APT database update. Helps identify stale cache issues. |
| `updates.check_duration_seconds` | Dependent | `$.check_duration_seconds` | Duration in seconds for the last update check. Useful for performance monitoring. |

## Template Triggers

**Note:** This template currently includes only items for monitoring update counts and details. It does not include pre-configured triggers in the YAML file.

### Creating Triggers Manually

You can create triggers in Zabbix to alert when updates are available. Here are recommended trigger configurations:

#### Security Updates Trigger
- **Name**: "Security updates available"
- **Expression**: `{template_name:updates.security_updates_count.last()}>0`
- **Severity**: Information
- **Description**: Alerts when security updates (vulnerability fixes) are available. These should be applied as soon as possible.

#### Recommended Updates Trigger
- **Name**: "Recommended updates available"
- **Expression**: `{template_name:updates.recommended_updates_count.last()}>0`
- **Severity**: Information
- **Description**: Alerts when recommended updates (bug fixes and improvements) are available.

#### Optional Updates Trigger
- **Name**: "Optional updates available"
- **Expression**: `{template_name:updates.optional_updates_count.last()}>0`
- **Severity**: Information
- **Description**: Alerts when optional updates (new features, documentation) are available.

## Template Groups

The template is assigned to the following group:
- **Templates/Applications**

This places it logically alongside other application monitoring templates in the Zabbix web interface.

## Tags

The template uses the following tags for better organization and filtering:

| Tag | Value | Purpose |
|-----|-------|--------|
| class | software | Identifies this as software-related monitoring |
| target | linux | Specifies Linux as the target operating system |
| component | updates | Marks items related to package updates |
| component | raw | Marks the master item that collects raw data |

## Importing the Template

### Method 1: Using Zabbix Web Interface

1. Download the template file:
   ```bash
   wget https://raw.githubusercontent.com/Aecroo/zabbix_agent2-apt-updates/master/templates/7.4/apt_updates_zabbix_agent2.yaml -O apt_updates_template.yaml
   ```

2. In the Zabbix web interface:
   - Navigate to **Configuration** > **Templates**
   - Click **Import**
   - Select the downloaded YAML file (`apt_updates_zabbix_agent2.yaml`)
   - Review the template information
   - Click **Import** to complete the import process

### Method 2: Using Zabbix API

```bash
curl -s -X POST -H "Content-Type: application/json" -d '{
  "jsonrpc": "2.0",
  "method": "configuration.import",
  "params": {
    "format": "yaml",
    "source": "file",
    "rule": "{"templates":[{"template":"apt updates by Zabbix agent 2","groups":[{"name":"Templates/Applications"}]}}}"
  },
  "auth": "YOUR_AUTH_TOKEN",
  "id": 1
}' http://your-zabbix-server/api/jsonrpc.php
```

## Linking the Template to Hosts

After importing the template, link it to your monitored hosts:

1. Go to **Configuration** > **Hosts** in the Zabbix web interface
2. Select the host where you've installed the APT updates plugin
3. Click on the **Templates** tab
4. Click **Add**
5. Search for "apt updates by Zabbix agent 2"
6. Select the template and click **Add**
7. The template will now be linked to the host

## Template Configuration Options

### Update Interval

The master item (`updates.get`) is configured with a 15-minute update interval by default. You can adjust this based on your monitoring needs:

- **For production systems**: 15-30 minutes (default)
- **For development/testing**: 5 minutes or less
- **For critical security monitoring**: Consider using lower thresholds for the security updates trigger

### Timeout Settings

The template uses a 10-minute timeout for the master item to accommodate:
- Systems with large numbers of updates
- Slow network connections
- Systems under heavy load

Adjust this if your environment has different requirements.

## JSONPath Preprocessing

All dependent items use JSONPath preprocessing to extract specific fields from the JSON response. The error handlers are configured as follows:

- **Numeric counts**: Return `0` on error (safe default)
- **Lists and details**: Return empty array `[]` on error (safe default)
- **Performance metrics**: Discard value on error (don't store invalid data)

## Troubleshooting Template Issues

### Template Not Showing Data

**Symptoms**: Items show "Not supported" or "No data"

**Solutions**:
1. Verify the plugin is installed and configured correctly on the host
2. Check that the Zabbix Agent 2 service is running: `sudo systemctl status zabbix-agent2`
3. Test the plugin manually: `sudo -u zabbix /usr/libexec/zabbix/zabbix-agent2-plugin-apt-updates --version`
4. Verify the configuration file exists at `/etc/zabbix/zabbix_agent2.d/apt-updates.conf`
5. Check Zabbix Agent logs: `journalctl -u zabbix-agent2 -f`

### Incorrect Data or Zero Counts

**Symptoms**: Template shows zero updates even though `apt list --upgradable` shows packages

**Solutions**:
1. Run `sudo apt update` on the monitored host to refresh package cache
2. Check that the Zabbix agent user has permissions to run APT commands
3. Verify the plugin binary has execute permissions
4. Check for error messages in Zabbix Agent logs

### Trigger Not Firing

**Symptoms**: Updates are available but triggers don't fire

**Solutions**:
1. Verify the trigger expression is correct in the template
2. Check that the host is properly linked to the template
3. Review Zabbix event log for any issues: **Monitoring** > **Events**
4. Test the trigger manually by creating a problem event

## Best Practices

### Monitoring Strategy

1. **Security updates**: Monitor closely with low thresholds (immediate action recommended)
2. **Recommended updates**: Monitor regularly, schedule updates during maintenance windows
3. **Optional updates**: Can be monitored less frequently or excluded if not needed

### Alerting Recommendations

- **Security updates**: Configure to alert immediately with high priority
- **Recommended updates**: Weekly digest or threshold-based alerts (e.g., >5 updates)
- **Optional updates**: Monthly review or manual inspection

### Performance Considerations

- The plugin executes `apt list --upgradable` which reads from the local APT cache
- For large systems with thousands of packages, consider:
  - Increasing the update interval
  - Adjusting timeout settings
  - Monitoring during off-peak hours

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 7.4-0 | July 2025 | Initial release with Zabbix 7.4 compatibility |

## Support and Resources

For issues and questions related to this template:

- **Project repository**: https://github.com/Aecroo/zabbix_agent2-apt-updates
- **Documentation**: See the main [README.md](../../README.md) in the project root
- **Plugin version**: Ensure your plugin is version 1.0.0 or later for full compatibility

## Related Documentation

- [Main README](../../README.md) - General installation and usage guide
- [CHANGELOG](../../CHANGELOG.md) - Plugin version history
- [Template YAML file](apt_updates_zabbix_agent2.yaml) - Raw template definition
