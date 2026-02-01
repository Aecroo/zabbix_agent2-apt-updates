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
	"strconv"
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

// Handler holds syscall implementation for request functions.
type Handler struct {
	sysCalls systemCalls
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
	WarningThreshold     int         `json:"warning_threshold,omitempty"`
	IsAboveWarning       bool        `json:"is_above_warning,omitempty"`
}

type systemCalls interface {
	execCommand(context.Context, string, ...string) *exec.Cmd
}

type osWrapper struct{}

// CheckUpdateCount returns the number of available APT updates
func (h *Handler) CheckUpdateCount(ctx context.Context, metricParams map[string]string, _ ...string) (any, error) {
	thresholdStr := metricParams["WarningThreshold"]
	threshold, err := strconv.Atoi(thresholdStr)
	if err != nil {
		threshold = 10 // default
	}

	result, err := h.checkAPTUpdates(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "failed to check APT updates")
	}

	// Store threshold in result for potential use in Zabbix triggering
	result.WarningThreshold = threshold
	result.IsAboveWarning = result.AvailableUpdates > threshold

	return result.AvailableUpdates, nil
}

// GetUpdateList returns a JSON list of available APT updates
func (h *Handler) GetUpdateList(ctx context.Context, metricParams map[string]string, _ ...string) (any, error) {
	result, err := h.checkAPTUpdates(ctx)
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
func (h *Handler) GetUpdateDetails(ctx context.Context, metricParams map[string]string, _ ...string) (any, error) {
	result, err := h.checkAPTUpdates(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "failed to check APT updates")
	}

	thresholdStr := metricParams["WarningThreshold"]
	threshold, err := strconv.Atoi(thresholdStr)
	if err != nil {
		threshold = 10 // default
	}

	result.WarningThreshold = threshold
	result.IsAboveWarning = result.AvailableUpdates > threshold

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

// checkAPTUpdates executes 'apt list --upgradable' and parses the output
func (h *Handler) checkAPTUpdates(ctx context.Context) (*CheckResult, error) {
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
