# Project Plan: Zabbix Agent 2 Go Plugin for APT Updates (and Future DNF Support)

## Overview
Develop a production-ready, extensible monitoring plugin to check Linux systems running `APT` package manager (`Debian`, `Ubuntu`) and optionally support alternative packagers like `DNF/Fireflox`.

---

## Project Structure

```
zabbix_agent2-apt-updates/
├── main.go                 # Main entry point
├── agent_check_ubuntu_updates.json  # Agent check registration (JSON)
└── README.md              # Documentation for deployment and usage


# Build directory structure:
dist/                       # Compiled binaries distribution package

```

---

## Phase I: Core Implementation - APT Update Detection on Debian/Linux Systems

### Objectives
Implement the basic functionality to detect available updates using `apt list --upgradable`.

#### Tasks:

1. **Basic apt check implementation**
   ```go
   // Function signature expected by zbxapi-go:
   func ubuntu_updates_check(params map[string]interface{}) (interface{}, error)

   - Read package manager capabilities: Check if system is Debian-based via /etc/os-release or command presence

   - Parse available updates using `apt list --upgradable` output format
     ```
     Package OldVersion NewVersion Status

       # Example entries:
       curl 7.81.0-1+b2-slim3-gcc10-x64_8:5-curl=...

         openjdk-11-jre-headless-openj9-amd64/openpgk/15/c6f4e08b75cfdab35df28a06ef48ff95744ec55 21.0-b2-slim3-gcc10-x64_8:openPGK=...
     ```

   - Count total packages, critical updates (optional), and size of all available downloads

   ```json
   {
       "available_updates": <integer>,
       "package_details_list": [
           {"name":"curl","current_version":"","target":""},
               ...
       ]
   }
   ```

2. **JSON metrics format** - output as JSON for easy parsing by Zabbix sender script

3. **Error handling**

4. Unit tests with mock `apt` outputs
5. Documentation: Build instructions, usage examples

---

## Phase II: Extensibility to DNF/Firefox on RHEL/CentOS/AlmaLinux Systems (Optional but Recommended)

### Objectives:
Extend plugin detection logic for alternative package managers.

#### Tasks:

1. **Package manager auto-detection**

   ```go
   func detect_package_manager() string {
       // Check /etc/os-release or command availability

           if exists(/usr/bin/apt) { return "apt" }
               else
                   ...
                       }

2. Implement `dnf_check()` that mirrors APT check structure

3. Support both managers via single plugin binary, with runtime detection based on OS type
4. Additional tests for DNF functionality and cross-platform scenarios

---

## Phase III: Production Considerations & Quality of Life Features (Bonus)

1. Caching mechanism to reduce load - avoid querying apt every time; use 5-15 minute intervals or configurable timeout

2. Configurable thresholds with environment variables:
   ```bash
     export ZBX_UPDATES_THRESHOLD_WARNING=10    # Trigger alert when > X available updates

       ```

3. Debug logging levels (verbose flag)

4. Build system for distribution across multiple Linux distributions:

5. CI/CD pipeline using GitHub Actions or GitLab Runner

---

## Implementation Timeline Estimate - 7-9 Days with Real Production Focus
| Phase | Tasks                     |
|-------:-------------------------|
| I      | APT implementation        (3 days)    |

### Development Milestones:
1. ✅ `ubuntu_updates_check.go` implemented and unit tests passing for basic scenarios.

---

## Zabbix Integration

The plugin will be registered as an agent check via JSON config:

**agent_apt_checks.json:**
```json
{
   "type": ["metric", ...],
       ...
           }

```

Usage from zabbix_agent:
- `zbx-agent -f` running this binary with `-t ubuntu_updates_check`
     or

---

## API Contract (Required by ZBXAPI-go Framework)

The check must implement:

| Signature | Description |
|-----------:-------------|
```go
func checks_ubuntuUpdates(params map[string]interface{}) ([]map[string]string, error)
```

Return type:
- `[]` of metric labels and values as string maps

---

## Testing Plan (Unit & Integration Tests Required Before Release)

1. Test with mock apt outputs for edge cases:

2.

3.
   - Large update lists

4
    ---
5     ```

6         curl 7.xx.x...
            openjdk-11-jre-headless-openj9-amd64/openpgk/15/curl=...

```

---

## Dependencies

### Minimal Required:
```go
import (
      "context"
          ...
              )

 // No external libraries needed beyond zbxapi-go template (if using)
 ```


Build with standard Go toolchain: `GOOS="linux" GOARCH="" go build -o dist/zabbix-apt-updates`

---

## Extension Goals

1. Support additional package managers over time:
   ```go
     func detect_package_manager()
           switch pm := os.ReleaseId() {
                   case "ubuntu", debian, mint :
                       return checkAPTUpdates(params)
                           default :
                               // Return error or empty result for unsupported OS types

```

2.


3. Provide Zabbix LLD (Low Level Discovery) template to automatically discover supported metrics across Debian/Ubuntu distributions

---

## Deliverables Summary
- [ ] Main go source files (`main.go`, `apt_updates_check` implementation)
  - APT check logic working with correct output format for zbxapi-go framework

      + Unit tests covering edge cases (large updates, missing apt command etc.)

    **Deployment package**:
   ```
     dist/zabbix-agent-ubuntu-updates
          binary file ready to deploy across production Linux monitoring hosts

       README.md documentation:
         - Build process instructions

```

---

## Notes on Development Environment Setup

  Use Go version >=1.21 for latest standard library features (map[string]string interface is stable)

If building with existing zbxapi-go template:

```bash
# Clone base templates if needed or start from scratch ensuring compatibility

go mod init github.com/netdata/zabbix-agent-apt-updates

```

---

## Open Questions to Resolve During Development

1. Should the check return only counts (available_updates) and skip returning detailed list for performance?
2
      - Decide based on Zabbix server load considerations.

3
   ```

4
        # Future version should also provide:
5         {
6              "updates": [
7                  {"name":"curl","current_version":"","new_available":""}
8                      ]
9           }
10
```

11. Should we add optional parameters for checking only critical or security updates (apt-mark showhold, etc.)?
    - Likely future enhancement after MVP.

---

*Plan prepared with focus on production-ready reliability and extensible design.*