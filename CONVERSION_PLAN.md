# Conversion Plan: Userparameter to Official Zabbix Agent 2 Go Plugin

## Current State Analysis

### Current Implementation (Userparameter Style)
The current implementation is a standalone Go binary that:
1. Accepts command-line arguments (`check` or `version`)
2. Executes APT/DNF commands to check for updates
3. Outputs JSON results to stdout
4. Is designed to be called by Zabbix Agent 2 via UserParameter configuration

**Key characteristics:**
- Simple CLI interface
- Environment variable-based configuration
- No integration with Zabbix SDK
- Manual error handling and output formatting

### Target Implementation (Official Go Plugin)
Based on the zabbix_example directory, the official plugin should:
1. Use the Zabbix Go SDK for plugin development
2. Implement specific interfaces: `Configurator`, `Exporter`, `Runner`
3. Register metrics with unique keys (e.g., `apt.updates[...]`)
4. Support configuration via Zabbix agent 2 config files
5. Handle credential validation and sessions
6. Provide proper logging through Zabbix SDK
7. Follow Zabbix plugin versioning conventions

## Implementation Plan

### Phase 1: Project Structure Reorganization

**Tasks:**
1. Create proper Go module structure following Zabbix plugin conventions
2. Organize code into packages:
   - `plugin/` - Main plugin implementation
   - `plugin/handlers/` - Business logic handlers
   - `plugin/params/` - Parameter definitions
   - `plugin/config.go` - Configuration management
3. Update go.mod to use Zabbix SDK dependencies
4. Remove old main.go and test files (or archive them)

### Phase 2: Core Plugin Implementation

**Main components to implement:**

1. **main.go** - Entry point with proper flag handling
   ```go
   // Based on zabbix_example/main.go
   // Uses flag.HandleFlags() and flag.DecideActionFromFlags()
   ```

2. **plugin/plugin.go** - Main plugin structure
   ```go
   type APTUpdatesPlugin struct {
       plugin.Base
       config *pluginConfig
       metrics map[aptMetricKey]*aptMetric
   }

   // Implement interfaces:
   // - Configurator (Configure, Validate)
   // - Exporter (Export)
   // - Runner (Run, Start, Stop)
   ```

3. **plugin/handlers/handlers.go** - Business logic
   ```go
   type Handler struct {
       sysCalls systemCalls
   }

   func (h *Handler) CheckAPTUpdates(ctx context.Context, params map[string]string) (*CheckResult, error)
   func (h *Handler) GetUpdateCount(ctx context.Context, params map[string]string) (int, error)
   ```

4. **plugin/params/params.go** - Parameter definitions
   ```go
   var Params = []*metric.Param{
       WarningThreshold,
       // Other connection parameters if needed
   }
   ```

5. **plugin/config.go** - Configuration management
   ```go
   type pluginConfig struct {
       System plugin.SystemOptions `conf:"optional"`
       Timeout int `conf:"optional,range=1:30"`
       // Plugin-specific config options
   }
   ```

### Phase 3: Metric Definition

Define the following metric keys:
- `apt.updates.count` - Returns number of available updates (simple integer)
- `apt.updates.list` - Returns JSON list of all available updates
- `apt.updates.details` - Returns detailed information including versions

**Implementation:**
```go
const (
    countMetric = aptMetricKey("apt.updates.count")
    listMetric  = aptMetricKey("apt.updates.list")
    detailsMetric = aptMetricKey("apt.updates.details")
)

func (p *APTUpdatesPlugin) registerMetrics() error {
    handler := handlers.New()

    p.metrics = map[aptMetricKey]*aptMetric{
        countMetric: {
            metric: metric.New(
                "Returns the number of available APT updates.",
                params.Params,
                false,  // Not text
            ),
            handler: handlers.WithJSONResponse(handler.CheckUpdateCount),
        },
        listMetric: {
            metric: metric.New(
                "Returns a list of available APT updates.",
                params.Params,
                true,   // Text output
            ),
            handler: handlers.WithJSONResponse(handler.GetUpdateList),
        },
    }

    return plugin.RegisterMetrics(p, Name, metricSet.List()...)
}
```

### Phase 4: Configuration File

Create `apt-updates.conf` with options:
- `Plugins.APTUpdates.System.Path` - Path to executable (required)
- `Plugins.APTUpdates.Timeout` - Timeout in seconds (1-30)
- `Plugins.APTUpdates.WarningThreshold` - Warning threshold for updates
- `Plugins.APTUpdates.Default.*` - Default parameter values

### Phase 5: Build System Updates

Update build scripts to:
1. Use proper Zabbix plugin versioning (PLUGIN_VERSION_* constants)
2. Include copyright headers
3. Support cross-compilation for all platforms
4. Generate proper binary names (e.g., `zabbix-agent2-plugin-apt-updates`)

### Phase 6: Testing Strategy

1. **Unit tests** - Test individual handlers with mock data
2. **Integration tests** - Test plugin registration and metric export
3. **Docker testing** - Verify in containerized environment
4. **Zabbix Agent 2 integration** - Test actual agent communication

### Phase 7: Documentation Updates

1. Update README.md with:
   - Installation instructions for official plugin
   - Configuration examples
   - Metric key reference
   - Migration guide from userparameter version
2. Create `CHANGELOG.md` entry for v0.2.0
3. Update example configuration files

## Migration Path

### For Users Upgrading from UserParameter Version:

1. **Configuration changes:**
   ```conf
   # Old (UserParameter):
   UserParameter=apt.updates.check[*],/path/to/plugin check $1

   # New (Plugin):
   Plugins.APTUpdates.System.Path=/path/to/zabbix-agent2-plugin-apt-updates
   Include=/etc/zabbix/zabbix_agent2.d/apt-updates.conf
   ```

2. **Item key changes:**
   - Old: `apt.updates.check[json]`
   - New: `apt.updates.count` or `apt.updates.list`

3. **Template updates:** Provide migration guide for template changes

## Timeline Estimate

1. **Phase 1-2 (Core Conversion):** 2-3 days
2. **Phase 3-4 (Metrics & Config):** 1-2 days
3. **Phase 5 (Build System):** 1 day
4. **Phase 6 (Testing):** 2-3 days
5. **Phase 7 (Documentation):** 1 day

**Total:** ~7-9 days for complete conversion and testing

## Risk Assessment

### High Risk Items:
1. **Zabbix SDK integration** - Need to ensure proper dependency management
2. **Metric compatibility** - Ensuring new keys work with existing templates
3. **Configuration migration** - Smooth transition path for users

### Mitigation Strategies:
1. Maintain backward compatibility in v0.2.0 if possible
2. Provide detailed migration documentation
3. Offer dual support period (both userparameter and plugin versions)
4. Thorough testing with multiple Zabbix agent versions

## Success Criteria

1. ✅ Plugin compiles successfully for all target platforms
2. ✅ All metrics return expected values in Docker test environment
3. ✅ Configuration is properly validated
4. ✅ Logging works through Zabbix SDK
5. ✅ Documentation is complete and accurate
6. ✅ v0.2.0 release created with pre-built binaries
7. ✅ Release notes clearly explain changes and migration path
