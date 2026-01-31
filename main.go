// Zabbix Agent 2 APT Updates Plugin
//
// This plugin checks for available package updates on Debian/Ubuntu systems
// using the APT package manager and returns results in a format compatible
// with Zabbix Agent 2.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// UpdateInfo represents a single package update
type UpdateInfo struct {
	Name    string `json:"name"`
	Current string `json:"current_version,omitempty"`
	Target  string `json:"target_version,omitempty"`
}

// CheckResult contains the complete check result
type CheckResult struct {
	AvailableUpdates     int         `json:"available_updates"`
	PackageDetailsList   []UpdateInfo `json:"package_details_list,omitempty"`
	WarningThreshold     int         `json:"warning_threshold,omitempty"`
	IsAboveWarning       bool        `json:"is_above_warning,omitempty"`
}

// Config holds plugin configuration
 type Config struct {
	 Debug              bool
	 WarningThreshold   int
 }

 var config Config = Config{
	 Debug:            false,
	 WarningThreshold: 10,
 }

 func init() {
	 // Load configuration from environment variables
	 if val, exists := os.LookupEnv("ZBX_DEBUG"); exists {
		 config.Debug = strings.ToLower(val) == "true" || val == "1"
	 }

	 if val, exists := os.LookupEnv("ZBX_UPDATES_THRESHOLD_WARNING"); exists {
		 if threshold, err := strconv.Atoi(val); err == nil {
			 config.WarningThreshold = threshold
		 }
	 }
 }

 // detectPackageManager checks which package manager is available
 func detectPackageManager() string {
	 // Check for APT (Debian/Ubuntu)
	 if _, err := exec.LookPath("apt"); err == nil {
		 return "apt"
	 }

	 // Check for DNF (RHEL/CentOS/Fedora)
	 if _, err := exec.LookPath("dnf"); err == nil {
		 return "dnf"
	 }

	 return "unknown"
 }

 // checkAPTUpdates executes 'apt list --upgradable' and parses the output
 func checkAPTUpdates() (*CheckResult, error) {
	 cmd := exec.Command("apt", "list", "--upgradable")

	 if config.Debug {
		 log.Printf("[DEBUG] Executing: %s", cmd.Args)
	 }

	 output, err := cmd.CombinedOutput()
	 if err != nil {
		 // Check if apt command exists
		 if exitErr, ok := err.(*exec.ExitError); ok {
			 // Exit code 100 means no upgrades available (normal)
			 if exitErr.ExitCode() == 100 {
				 return &CheckResult{
					 AvailableUpdates: 0,
					 PackageDetailsList: []UpdateInfo{},
				 }, nil
			 }
		 }
		 return nil, fmt.Errorf("failed to execute apt list: %w", err)
	 }

	 // Parse the output
	 lines := strings.Split(string(output), "\n")
	 var updates []UpdateInfo

	 for _, line := range lines {
		 line = strings.TrimSpace(line)
		 if line == "" || !strings.Contains(line, "/") {
			 continue
		 }

		 // Parse format: package/state version
		 parts := strings.Fields(line)
		 if len(parts) < 2 {
			 continue
		 }

		 // Extract package name (before /)
		 pkgParts := strings.Split(parts[0], "/")
		 if len(pkgParts) < 1 {
			 continue
		 }

		 pkgName := pkgParts[0]
		 version := parts[len(parts)-1] // Last field is the version

		 updates = append(updates, UpdateInfo{
			 Name:   pkgName,
			 Target: version,
		 })
	 }

	 result := &CheckResult{
		 AvailableUpdates:   len(updates),
		 PackageDetailsList: updates,
		 WarningThreshold:   config.WarningThreshold,
		 IsAboveWarning:     len(updates) > config.WarningThreshold,
	 }

	 return result, nil
 }

 // checkDNFUpdates executes 'dnf check-update' and parses the output
 func checkDNFUpdates() (*CheckResult, error) {
	 cmd := exec.Command("dnf", "check-update")

	 if config.Debug {
		 log.Printf("[DEBUG] Executing: %s", cmd.Args)
	 }

	 output, err := cmd.CombinedOutput()
	 if err != nil {
		 return nil, fmt.Errorf("failed to execute dnf check-update: %w", err)
	 }

	 // Parse the output
	 lines := strings.Split(string(output), "\n")
	 var updates []UpdateInfo

	 for _, line := range lines {
		 line = strings.TrimSpace(line)
		 if line == "" || strings.HasPrefix(line, "Last metadata") {
			 continue
		 }

		 // Parse format: package.x86_64 update version-1.elastic.el7
		 parts := strings.Fields(line)
		 if len(parts) < 2 {
			 continue
		 }

		 pkgName := strings.TrimSuffix(parts[0], ".x86_64")
		 pkgName = strings.TrimSuffix(pkgName, ".noarch")

		 // Version is the last part after multiple dashes
		 versionParts := strings.Split(parts[len(parts)-1], "-")
		 if len(versionParts) >= 2 {
			 version := strings.Join(versionParts[0:len(versionParts)-2], "-")
			 updates = append(updates, UpdateInfo{
				 Name:   pkgName,
				 Target: version,
			 })
		 }
	 }

	 result := &CheckResult{
		 AvailableUpdates:   len(updates),
		 PackageDetailsList: updates,
		 WarningThreshold:   config.WarningThreshold,
		 IsAboveWarning:     len(updates) > config.WarningThreshold,
	 }

	 return result, nil
 }

 // Check executes the appropriate package manager check based on OS detection
 func Check() (*CheckResult, error) {
	 pm := detectPackageManager()

	 if config.Debug {
		 log.Printf("[DEBUG] Detected package manager: %s", pm)
	 }

	 switch pm {
	 case "apt":
		 return checkAPTUpdates()
	 case "dnf":
		 return checkDNFUpdates()
	 default:
		 return nil, fmt.Errorf("unsupported package manager: %s", pm)
	 }
 }

 // PrintJSON marshals and prints the result as JSON
 func PrintJSON(result *CheckResult) error {
	 data, err := json.MarshalIndent(result, "", "  ")
	 if err != nil {
		 return fmt.Errorf("failed to marshal JSON: %w", err)
	 }

	 fmt.Println(string(data))
	 return nil
 }

 func main() {
	 // Check command line arguments
	 if len(os.Args) < 2 {
		 fmt.Fprintf(os.Stderr, "Usage: %s check\n", os.Args[0])
		 os.Exit(1)
	 }

	 command := os.Args[1]

	 switch command {
	 case "check":
		 result, err := Check()
		 if err != nil {
			 fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			 os.Exit(2)
		 }

		 if err := PrintJSON(result); err != nil {
			 fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			 os.Exit(3)
		 }

	 case "version":
		 fmt.Println("zabbix-apt-updates v1.0.0")
	 default:
		 fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		 fmt.Fprintf(os.Stderr, "Usage: %s [check|version]\n", os.Args[0])
		 os.Exit(1)
	 }
 }
