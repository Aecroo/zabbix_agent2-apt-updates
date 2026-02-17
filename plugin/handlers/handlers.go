/*
** Copyright (C) 2001-2026 Zabbix SIA
**
** Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
** documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
** rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
** permit persons to whom the Software is furnished to do so, subject to the following conditions:
**
** The above copyright notice and this permission notice shall be included in all copies or substantial portions
** of the Software.
**
** THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
** WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
** COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
** TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
** SOFTWARE.
**/

package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.zabbix.com/sdk/errs"
)

var (
	_ HandlerFunc = WithJSONResponse(nil)
	_ HandlerFunc = (*Handler)(nil).CheckUpdateCount
	_ HandlerFunc = (*Handler)(nil).GetUpdateList
	_ HandlerFunc = (*Handler)(nil).GetUpdateDetails
	_ systemCalls = osWrapper{}
)

// HandlerFunc describes the signature all metric handler functions must have.
type HandlerFunc func(
	ctx context.Context,
	metricParams map[string]string,
	extraParams ...string,
) (any, error)

// UpdateType represents the type of update to check
type UpdateType string

const (
	UpdateTypeAll        UpdateType = "all"
	UpdateTypeSecurity   UpdateType = "security"
	UpdateTypeRecommended UpdateType = "recommended"
	UpdateTypeOptional   UpdateType = "optional"
)

// Handler holds syscall implementation for request functions.
type Handler struct {
	sysCalls systemCalls
}

// GetAllUpdates returns comprehensive information about all available APT updates
type AllUpdatesResult struct {
	SecurityUpdatesCount     int         `json:"security_updates_count"`
	RecommendedUpdatesCount  int         `json:"recommended_updates_count"`
	OptionalUpdatesCount    int         `json:"optional_updates_count"`
	AllUpdatesCount         int         `json:"all_updates_count"`

	PhasedUpdatesCount     int         `json:"phased_updates_count,omitempty"`
	PhasedUpdatesList       []string   `json:"phased_updates_list,omitempty"`
	PhasedUpdatesDetails    []UpdateInfo `json:"phased_updates_details,omitempty"`

	SecurityUpdatesList     []string   `json:"security_updates_list,omitempty"`
	RecommendedUpdatesList  []string   `json:"recommended_updates_list,omitempty"`
	OptionalUpdatesList    []string   `json:"optional_updates_list,omitempty"`
	AllUpdatesList         []string   `json:"all_updates_list,omitempty"`

	SecurityUpdatesDetails  []UpdateInfo `json:"security_updates_details,omitempty"`
	RecommendedUpdatesDetails []UpdateInfo `json:"recommended_updates_details,omitempty"`
	OptionalUpdatesDetails   []UpdateInfo `json:"optional_updates_details,omitempty"`
	AllUpdatesDetails      []UpdateInfo `json:"all_updates_details,omitempty"`

	CheckDurationSeconds float64 `json:"check_duration_seconds"`
	LastAptUpdateTime     int64    `json:"last_apt_update_time"` // Unix timestamp in seconds
}

// UpdateInfo represents a single package update
type UpdateInfo struct {
	Name     string `json:"name"`
	Current  string `json:"current_version,omitempty"`
	Target   string `json:"target_version,omitempty"`
	IsPhased bool   `json:"is_phased,omitempty"` // Indicates if this update is subject to phased rollout
}

// CheckResult contains the complete check result
type CheckResult struct {
	AvailableUpdates     int         `json:"available_updates"`
	PackageDetailsList   []UpdateInfo `json:"package_details_list,omitempty"`
	CheckDurationSeconds float64 `json:"check_duration_seconds"`
	LastAptUpdateTime     int64       `json:"last_apt_update_time"` // Unix timestamp in seconds
}

type commandExecutor interface {
	execute(ctx context.Context, name string, args ...string) ([]byte, error)
}

type systemCalls interface {
	execCommand(ctx context.Context, name string, args ...string) ([]byte, error)
}

type osWrapper struct{}

// CheckUpdateCount returns the number of available APT updates
func (h *Handler) CheckUpdateCount(ctx context.Context, metricParams map[string]string, extraParams ...string) (any, error) {
	updateType := getUpdateTypeFromExtra(extraParams)

	result, err := h.checkAPTUpdates(ctx, updateType, false)
	if err != nil {
		return nil, errs.Wrap(err, "failed to check APT updates")
	}

	return result.AvailableUpdates, nil
}

// GetUpdateList returns a JSON list of available APT updates
func (h *Handler) GetUpdateList(ctx context.Context, metricParams map[string]string, extraParams ...string) (any, error) {
	updateType := getUpdateTypeFromExtra(extraParams)
	result, err := h.checkAPTUpdates(ctx, updateType, false)
	if err != nil {
		return nil, errs.Wrap(err, "failed to check APT updates")
	}

	// Return just the list of packages
	var packageNames []string
	for _, pkg := range result.PackageDetailsList {
		packageNames = append(packageNames, pkg.Name)
	}

	return packageNames, nil
}

// GetUpdateDetails returns detailed information about available APT updates
func (h *Handler) GetUpdateDetails(ctx context.Context, metricParams map[string]string, extraParams ...string) (any, error) {
	updateType := getUpdateTypeFromExtra(extraParams)
	result, err := h.checkAPTUpdates(ctx, updateType, false)
	if err != nil {
		return nil, errs.Wrap(err, "failed to check APT updates")
	}

	return result, nil
}

// GetAllUpdates returns comprehensive information about all types of available APT updates
func (h *Handler) GetAllUpdates(ctx context.Context, metricParams map[string]string, extraParams ...string) (any, error) {
	result := &AllUpdatesResult{}

	// Track start time for duration calculation
	startTime := time.Now()

	// First pass: get all updates with phased updates excluded (includePhased=false)
	// This will parse the "deferred due to phasing" section from apt-get output
	// Pass a slice with nil so checkAPTUpdates can populate it with detected phased packages
	deferredPackages := []map[string]bool{nil}
	_, err := h.checkAPTUpdates(ctx, UpdateTypeAll, false, deferredPackages...)
	if err != nil {
		return nil, errs.Wrap(err, "failed to check APT updates for first pass")
	}


	// Second pass: get all updates including phased ones
	// Pass the deferred packages map from the first pass so IsPhased can be set correctly
	var deferredPackagesMap map[string]bool = nil
	if len(deferredPackages) > 0 && deferredPackages[0] != nil {
		deferredPackagesMap = deferredPackages[0]
	}
	allUpdates, err := h.checkAPTUpdates(ctx, UpdateTypeAll, true, deferredPackagesMap)
	if err != nil {
		return nil, errs.Wrap(err, "failed to check APT updates for 'all'")
	}

	// Calculate check duration
	result.CheckDurationSeconds = time.Since(startTime).Seconds()

	// Initialize slice fields to empty arrays (not nil) for consistent JSON output
	result.PhasedUpdatesList = []string{}
	result.PhasedUpdatesDetails = []UpdateInfo{}
	result.SecurityUpdatesList = []string{}
	result.SecurityUpdatesDetails = []UpdateInfo{}
	result.RecommendedUpdatesList = []string{}
	result.RecommendedUpdatesDetails = []UpdateInfo{}
	result.OptionalUpdatesList = []string{}
	result.OptionalUpdatesDetails = []UpdateInfo{}

	// Get last apt update time from package lists (also available in allUpdates)
	if allUpdates.LastAptUpdateTime != 0 {
		result.LastAptUpdateTime = allUpdates.LastAptUpdateTime
	} else {
		// Fallback: try to get it directly if not already set by checkAPTUpdates
		lastUpdateTime, err := h.getLastAptUpdateTime()
		if err == nil {
			result.LastAptUpdateTime = lastUpdateTime.Unix()
		} else {
			// If we can't get the time (e.g., no package lists), set to 0
			result.LastAptUpdateTime = 0
		}
	}

	// Set all updates data (including phased)
	result.AllUpdatesCount = len(allUpdates.PackageDetailsList)
	result.AllUpdatesList = make([]string, len(allUpdates.PackageDetailsList))
	result.AllUpdatesDetails = make([]UpdateInfo, len(allUpdates.PackageDetailsList))
	for i, pkg := range allUpdates.PackageDetailsList {
		result.AllUpdatesList[i] = pkg.Name
		result.AllUpdatesDetails[i] = pkg
	}

	// Filter updates by type in-memory instead of calling apt multiple times
	// This significantly reduces execution time and prevents timeout issues on ARM platforms
	for _, pkg := range allUpdates.PackageDetailsList {
		// Phased updates should be counted separately, not included in regular categories
		if isPhasedUpdate(pkg) {
			result.PhasedUpdatesCount++
			result.PhasedUpdatesList = append(result.PhasedUpdatesList, pkg.Name)
			result.PhasedUpdatesDetails = append(result.PhasedUpdatesDetails, pkg)
			continue
		}

		isSecurity, err := h.isPackageOfType(ctx, pkg.Name, UpdateTypeSecurity)
		if err != nil {
			// If we can't determine the type, skip it for security updates
			continue
		}
		if isSecurity {
			result.SecurityUpdatesCount++
			result.SecurityUpdatesList = append(result.SecurityUpdatesList, pkg.Name)
			result.SecurityUpdatesDetails = append(result.SecurityUpdatesDetails, pkg)
		}

		// For recommended and optional, we use the same logic as before
		// Recommended is treated as all updates (can be enhanced later)
		result.RecommendedUpdatesCount++
		result.RecommendedUpdatesList = append(result.RecommendedUpdatesList, pkg.Name)
		result.RecommendedUpdatesDetails = append(result.RecommendedUpdatesDetails, pkg)

		isOptional, err := h.isPackageOfType(ctx, pkg.Name, UpdateTypeOptional)
		if err != nil {
			// If we can't determine the type, skip it for optional updates
			continue
		}
		if isOptional {
			result.OptionalUpdatesCount++
			result.OptionalUpdatesList = append(result.OptionalUpdatesList, pkg.Name)
			result.OptionalUpdatesDetails = append(result.OptionalUpdatesDetails, pkg)
		}
	}

	return result, nil
}

// New creates a new handler with initialized clients for system calls.
func New() *Handler {
	return &Handler{
		sysCalls: osWrapper{},
	}
}

// WithJSONResponse wraps a handler function, marshaling its response
// to a JSON object and returning it as string.
func WithJSONResponse(handler HandlerFunc) HandlerFunc {
	return func(
		ctx context.Context, metricParams map[string]string, extraParams ...string,
	) (any, error) {
		res, err := handler(ctx, metricParams, extraParams...)
		if err != nil {
			return nil, errs.Wrap(err, "failed to receive the result")
		}

		jsonRes, err := json.Marshal(res)
		if err != nil {
			return nil, errs.Wrap(err, "failed to marshal result to JSON")
		}

		return string(jsonRes), nil
	}
}

// getUpdateTypeFromExtra extracts the update type from extra parameters
// When user calls apt.updates[security], Zabbix passes "security" as first extra param
func getUpdateTypeFromExtra(extraParams []string) UpdateType {
	if len(extraParams) > 0 {
		typeStr := strings.TrimSpace(extraParams[0])
		switch typeStr {
		case "security":
			return UpdateTypeSecurity
		case "recommended":
			return UpdateTypeRecommended
		case "optional":
			return UpdateTypeOptional
		}
	}
	return UpdateTypeAll
}

// getUpdateType extracts the update type from metric parameters (fallback method)
func getUpdateType(metricParams map[string]string) UpdateType {
	// The parameter name is the bracket content in apt.updates[security]
	for key := range metricParams {
		if strings.Contains(key, "[") && strings.Contains(key, "]") {
			// Extract the content between brackets
			start := strings.Index(key, "[")
			end := strings.Index(key, "]")
			if start >= 0 && end > start {
				return UpdateType(strings.TrimSpace(key[start+1 : end]))
			}
		}
	}
	return UpdateTypeAll
}

// isPhasedUpdate checks if a package update is subject to phased updates
// Phased updates are gradually rolled out to users and should be counted separately
func isPhasedUpdate(pkg UpdateInfo) bool {
	// First check the IsPhased field (set during parsing)
	if pkg.IsPhased {
		return true
	}
	// Fallback: Check if the target version contains "phased" indicator
	return strings.Contains(strings.ToLower(pkg.Target), "[phased") ||
		strings.Contains(strings.ToLower(pkg.Name+" "+pkg.Target), "phased")
}

// getUpdateTypeAndFlagsFromExtra extracts the update type and flags from extra parameters
// When user calls apt.updates[security], Zabbix passes "security" as first extra param
func getUpdateTypeAndFlagsFromExtra(extraParams []string) (UpdateType, bool) {
	updateType := UpdateTypeAll
	includePhased := false

	if len(extraParams) > 0 {
		typeStr := strings.TrimSpace(extraParams[0])
		switch typeStr {
		case "security":
			updateType = UpdateTypeSecurity
		case "recommended":
			updateType = UpdateTypeRecommended
		case "optional":
			updateType = UpdateTypeOptional
		}
	}

	// Check for flags - look for strings like "include-phased" or "phased"
	for _, param := range extraParams {
		if strings.Contains(strings.ToLower(param), "phased") ||
		   strings.Contains(strings.ToLower(param), "include") {
			includePhased = true
		}
	}

	return updateType, includePhased
}

// isPackageOfType checks if a package belongs to a specific update type category
func (h *Handler) isPackageOfType(ctx context.Context, pkgName string, updateType UpdateType) (bool, error) {
	switch updateType {
	case UpdateTypeSecurity:
		// Check if package comes from security repository
		output, err := h.sysCalls.execCommand(ctx, "apt-cache", "policy", pkgName)
		if err != nil {
			return false, fmt.Errorf("failed to check policy for %s: %w", pkgName, err)
		}

		// Look for security repository in the output
		outputStr := string(output)
		// Security packages come from repositories like:
		//   https://security.ubuntu.com/ubuntu
		//   http://security.debian.org
		return strings.Contains(outputStr, "security.") ||
			strings.Contains(outputStr, "Ubuntu: noble-security") ||
			strings.Contains(outputStr, "Debian-Security"), nil

	case UpdateTypeRecommended:
		// For now, treat recommended as all updates (can be enhanced later)
		// In Debian/Ubuntu, there's no direct way to distinguish recommended vs optional
		// Both are typically in the main repositories
		return true, nil

	case UpdateTypeOptional:
		// Optional packages - these would be from universe/multiverse
		output, err := h.sysCalls.execCommand(ctx, "apt-cache", "policy", pkgName)
		if err != nil {
			return false, fmt.Errorf("failed to check policy for %s: %w", pkgName, err)
		}

		outputStr := string(output)
		// Optional packages typically come from universe/multiverse
		return strings.Contains(outputStr, "universe") ||
			strings.Contains(outputStr, "multiverse"), nil

	default:
		return true, nil
	}
}

// getLastAptUpdateTime returns the most recent modification time of APT package lists
// This indicates when the last 'apt update' was run
func (h *Handler) getLastAptUpdateTime() (time.Time, error) {
	listDir := "/var/lib/apt/lists"

	// Use find command to get the most recent file modification time
	// This is more reliable than walking the directory as it handles all APT file types
	// (InRelease, Packages, Sources, etc.) and permission issues better
	output, err := h.sysCalls.execCommand(context.Background(), "find", listDir, "-type", "f", "-printf", "%T@\n")
	if err != nil {
		// find may return exit code 1 if there are permission errors on some directories
		// (e.g., /var/lib/apt/lists/partial), but still produces valid output for accessible files
		// Only return zero time if there's no output at all
		if len(output) == 0 {
			return time.Time{}, nil
		}
		// Continue processing the output even if find had permission errors
	}

	// Parse the output to find the maximum timestamp
	var maxTime float64
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		timestamp, err := strconv.ParseFloat(line, 64)
		if err != nil {
			// Skip invalid lines
			continue
		}
		// find -printf "%T@" returns seconds (with nanosecond precision as decimal)
		// No conversion needed - the value is already in seconds
		if timestamp > maxTime {
			maxTime = timestamp
		}
	}

	// If no files found, return zero time (not an error)
	if maxTime == 0 {
		return time.Time{}, nil
	}

	// Convert float64 timestamp to time.Time
	return time.Unix(int64(maxTime), 0), nil
}

// parsePackageLine parses a single line from 'apt list --upgradable' output (DEPRECATED)
// Returns: package name, target version (cleaned), and whether it's a phased update
func parsePackageLine(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", false
	}

	// Check for warning lines or other non-package lines
	if strings.HasPrefix(strings.ToLower(line), "warning:") {
		return "", "", false
	}

	parts := strings.Fields(line)
	if len(parts) < 2 {
		return "", "", false
	}

	// Extract package name (first part before '/')
	pkgName := strings.Split(parts[0], "/")[0]

	// Check if this line contains a phased update indicator [phased XX%]
	isPhased := strings.Contains(strings.ToLower(line), "[phased")

	// Find the field that contains the actual version (ends with ']')
	// The version is the last field ending with ] that doesn't contain [ (which would be [phased XX%])
	var lastFieldWithBracket string
	for i := len(parts) - 1; i >= 0; i-- {
		if strings.HasSuffix(parts[i], "]") && !strings.Contains(strings.ToLower(parts[i]), "[") {
			// Found a field ending with ] that doesn't contain [ (the version)
			lastFieldWithBracket = parts[i]
			break
		}
	}

	if lastFieldWithBracket == "" {
		return "", "", false
	}

	// Extract version by removing the trailing ']' from last field
	versionStr := strings.TrimSuffix(lastFieldWithBracket, "]")
	version := strings.TrimSpace(versionStr)

	return pkgName, version, isPhased
}

// checkAPTUpdates executes 'apt-get -s upgrade' and parses the output
// This method uses apt-get simulation which respects phased updates by default
// Using 'apt-get -s upgrade' instead of 'apt list --upgradable' ensures phasing is respected
func (h *Handler) checkAPTUpdates(ctx context.Context, updateType UpdateType, includePhased bool, deferredPackages ...map[string]bool) (*CheckResult, error) {
	startTime := time.Now()

	// Use apt-get simulation so phasing is respected
	// Force C locale so parsing is stable even on de_DE systems
	args := []string{"env", "LC_ALL=C", "LANG=C", "apt-get", "-s"}

	// Explicitly set phased updates behavior
	if includePhased {
		args = append(args, "-o", "APT::Get::Always-Include-Phased-Updates=true")
	} else {
		args = append(args, "-o", "APT::Get::Always-Include-Phased-Updates=false")
	}

	args = append(args, "upgrade")

	output, err := h.sysCalls.execCommand(ctx, args[0], args[1:]...)
	if err != nil {
		// For simulation, non-zero is still possible; but if output is empty -> treat as error
		if len(output) == 0 {
			return nil, fmt.Errorf("failed to execute apt-get -s upgrade: %w", err)
		}
		// Continue with what output we have
	}

	var updates []UpdateInfo
	deferredPhasedPackages := make(map[string]bool)
	var sc *bufio.Scanner

	// If no pre-populated map was provided, parse the "deferred due to phasing" section first
	if len(deferredPackages) == 0 || deferredPackages[0] == nil {
		// Process output line by line to collect all deferred packages
		sc = bufio.NewScanner(strings.NewReader(string(output)))
		foundPhasingHeader := false

		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())

			// Check if we found the "deferred due to phasing" header line
			if strings.Contains(line, "deferred due to phasing:") {
				foundPhasingHeader = true
				continue
			}

			// If we found the header, collect package names from subsequent lines
			if foundPhasingHeader && line != "" {
				// Skip summary lines like "0 upgraded", "1 newly installed", etc.
				summaryRe := regexp.MustCompile(`^\d+ (upgraded|newly installed|to remove)`)
				if !summaryRe.MatchString(line) {
					// Collect all package names from this line
					packageNames := strings.Fields(line)
					for _, pkgName := range packageNames {
						deferredPhasedPackages[pkgName] = true
					}
				}
			}

			// Stop when we reach the summary line with numbers + "upgraded"
			if foundPhasingHeader && regexp.MustCompile(`^\d+ upgraded`).MatchString(line) {
				break
			}
		}


		// Store for return if this is the first pass and we found deferred packages
		if len(deferredPhasedPackages) > 0 {
			// Populate the caller's slice element instead of reassigning (so changes propagate back)
			if len(deferredPackages) == 0 {
				// No parameters were passed, create a new slice
				deferredPackages = []map[string]bool{deferredPhasedPackages}
			} else {
				// A slice with nil at index 0 was passed, populate it
				deferredPackages[0] = deferredPhasedPackages
			}
		}
	}

	// Parse the output from apt-get -s upgrade
	// Format: Inst <pkg> [<old>] (<new> <repo> ...)
	re := regexp.MustCompile(`^Inst\s+(\S+)(?:\s+\[([^\]]+)\])?\s+\(([^ )]+)`)
	sc = bufio.NewScanner(strings.NewReader(string(output)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || !strings.HasPrefix(line, "Inst ") {
			continue
		}

		m := re.FindStringSubmatch(line)
		if len(m) < 4 {
			continue
		}

		pkgName := m[1]
		current := strings.TrimSpace(m[2])
		target := strings.TrimSpace(m[3])

		// Check if this package is phased
		isPhased := false
		if len(deferredPackages) > 0 && deferredPackages[0] != nil {
			// If a deferred packages map was provided, check if this package is in it
			if _, isDeferred := deferredPackages[0][pkgName]; isDeferred {
				isPhased = true
			}
		}

		// Filter by update type if needed
		if updateType != UpdateTypeAll && updateType != "" {
			isMatch, err := h.isPackageOfType(ctx, pkgName, updateType)
			if err != nil {
				// If we can't determine the type, skip this package for filtered queries
				continue
			}
			if !isMatch {
				continue
			}
		}

		updates = append(updates, UpdateInfo{
			Name:    pkgName,
			Current: current,
			Target:  target,
			IsPhased: isPhased,
		})
	}

	// Filter updates by type if needed (for security, recommended, optional)
	if updateType != UpdateTypeAll && updateType != "" {
		filteredUpdates := []UpdateInfo{}
		for _, pkg := range updates {
			isMatch, err := h.isPackageOfType(ctx, pkg.Name, updateType)
			if err != nil {
				// If we can't determine the type, skip this package
				continue
			}
			if !isMatch {
				continue
			}
			filteredUpdates = append(filteredUpdates, pkg)
		}
		updates = filteredUpdates
	}

	result := &CheckResult{
		AvailableUpdates:      len(updates),
		PackageDetailsList:     updates,
		CheckDurationSeconds: time.Since(startTime).Seconds(),
	}

	// Get last apt update time from package lists
	lastUpdateTime, err := h.getLastAptUpdateTime()
	if err == nil {
		result.LastAptUpdateTime = lastUpdateTime.Unix()
	}

	return result, nil
}

func (osWrapper) execCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}
