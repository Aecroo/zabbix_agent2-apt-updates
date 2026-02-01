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
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

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

	SecurityUpdatesList     []string   `json:"security_updates_list,omitempty"`
	RecommendedUpdatesList  []string   `json:"recommended_updates_list,omitempty"`
	OptionalUpdatesList    []string   `json:"optional_updates_list,omitempty"`
	AllUpdatesList         []string   `json:"all_updates_list,omitempty"`

	SecurityUpdatesDetails  []UpdateInfo `json:"security_updates_details,omitempty"`
	RecommendedUpdatesDetails []UpdateInfo `json:"recommended_updates_details,omitempty"`
	OptionalUpdatesDetails   []UpdateInfo `json:"optional_updates_details,omitempty"`
	AllUpdatesDetails      []UpdateInfo `json:"all_updates_details,omitempty"`
}

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
}

type systemCalls interface {
	execCommand(context.Context, string, ...string) *exec.Cmd
}

type osWrapper struct{}

// CheckUpdateCount returns the number of available APT updates
func (h *Handler) CheckUpdateCount(ctx context.Context, metricParams map[string]string, extraParams ...string) (any, error) {
	updateType := getUpdateTypeFromExtra(extraParams)

	result, err := h.checkAPTUpdates(ctx, updateType)
	if err != nil {
		return nil, errs.Wrap(err, "failed to check APT updates")
	}

	return result.AvailableUpdates, nil
}

// GetUpdateList returns a JSON list of available APT updates
func (h *Handler) GetUpdateList(ctx context.Context, metricParams map[string]string, extraParams ...string) (any, error) {
	updateType := getUpdateTypeFromExtra(extraParams)
	result, err := h.checkAPTUpdates(ctx, updateType)
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
	result, err := h.checkAPTUpdates(ctx, updateType)
	if err != nil {
		return nil, errs.Wrap(err, "failed to check APT updates")
	}

	return result, nil
}

// GetAllUpdates returns comprehensive information about all types of available APT updates
func (h *Handler) GetAllUpdates(ctx context.Context, metricParams map[string]string, extraParams ...string) (any, error) {
	result := &AllUpdatesResult{}

	// Get all updates first
	allUpdates, err := h.checkAPTUpdates(ctx, UpdateTypeAll)
	if err != nil {
		return nil, errs.Wrap(err, "failed to check APT updates for 'all'")
	}

	result.AllUpdatesCount = allUpdates.AvailableUpdates
	result.AllUpdatesList = make([]string, len(allUpdates.PackageDetailsList))
	result.AllUpdatesDetails = make([]UpdateInfo, len(allUpdates.PackageDetailsList))
	for i, pkg := range allUpdates.PackageDetailsList {
		result.AllUpdatesList[i] = pkg.Name
		result.AllUpdatesDetails[i] = pkg
	}

	// Get security updates
	securityUpdates, err := h.checkAPTUpdates(ctx, UpdateTypeSecurity)
	if err != nil {
		return nil, errs.Wrap(err, "failed to check APT updates for 'security'")
	}

	result.SecurityUpdatesCount = securityUpdates.AvailableUpdates
	result.SecurityUpdatesList = make([]string, len(securityUpdates.PackageDetailsList))
	result.SecurityUpdatesDetails = make([]UpdateInfo, len(securityUpdates.PackageDetailsList))
	for i, pkg := range securityUpdates.PackageDetailsList {
		result.SecurityUpdatesList[i] = pkg.Name
		result.SecurityUpdatesDetails[i] = pkg
	}

	// Get recommended updates
	recommendedUpdates, err := h.checkAPTUpdates(ctx, UpdateTypeRecommended)
	if err != nil {
		return nil, errs.Wrap(err, "failed to check APT updates for 'recommended'")
	}

	result.RecommendedUpdatesCount = recommendedUpdates.AvailableUpdates
	result.RecommendedUpdatesList = make([]string, len(recommendedUpdates.PackageDetailsList))
	result.RecommendedUpdatesDetails = make([]UpdateInfo, len(recommendedUpdates.PackageDetailsList))
	for i, pkg := range recommendedUpdates.PackageDetailsList {
		result.RecommendedUpdatesList[i] = pkg.Name
		result.RecommendedUpdatesDetails[i] = pkg
	}

	// Get optional updates
	optionalUpdates, err := h.checkAPTUpdates(ctx, UpdateTypeOptional)
	if err != nil {
		return nil, errs.Wrap(err, "failed to check APT updates for 'optional'")
	}

	result.OptionalUpdatesCount = optionalUpdates.AvailableUpdates
	result.OptionalUpdatesList = make([]string, len(optionalUpdates.PackageDetailsList))
	result.OptionalUpdatesDetails = make([]UpdateInfo, len(optionalUpdates.PackageDetailsList))
	for i, pkg := range optionalUpdates.PackageDetailsList {
		result.OptionalUpdatesList[i] = pkg.Name
		result.OptionalUpdatesDetails[i] = pkg
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

// isPackageOfType checks if a package belongs to a specific update type category
func (h *Handler) isPackageOfType(ctx context.Context, pkgName string, updateType UpdateType) (bool, error) {
	switch updateType {
	case UpdateTypeSecurity:
		// Check if package comes from security repository
		cmd := h.sysCalls.execCommand(ctx, "apt-cache", "policy", pkgName)
		output, err := cmd.CombinedOutput()
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
		cmd := h.sysCalls.execCommand(ctx, "apt-cache", "policy", pkgName)
		output, err := cmd.CombinedOutput()
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

// checkAPTUpdates executes 'apt list --upgradable' and parses the output
func (h *Handler) checkAPTUpdates(ctx context.Context, updateType UpdateType) (*CheckResult, error) {
	cmd := h.sysCalls.execCommand(ctx, "apt", "list", "--upgradable")

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if apt command exists
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 100 means no upgrades available (normal)
			if exitErr.ExitCode() == 100 {
				return &CheckResult{
					AvailableUpdates:    0,
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

		// Filter by update type if needed
		if updateType != UpdateTypeAll && updateType != "" {
			isMatch, err := h.isPackageOfType(ctx, pkgName, updateType)
			if err != nil {
				// If we can't determine the type, skip this package for filtered queries
				if updateType != UpdateTypeAll {
					continue
				}
			}
			if !isMatch {
				continue
			}
		}

		updates = append(updates, UpdateInfo{
			Name:   pkgName,
			Target: version,
		})
	}

	result := &CheckResult{
		AvailableUpdates:    len(updates),
		PackageDetailsList: updates,
	}

	return result, nil
}

func (osWrapper) execCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}
