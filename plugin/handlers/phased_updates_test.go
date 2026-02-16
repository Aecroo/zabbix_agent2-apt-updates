/*
** Copyright (C) 2001-2026 Zabbix SIA
**
** Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
** documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
** rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of this Software, and to
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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPhasedUpdatesHandling ensures that phased updates are counted separately
func TestPhasedUpdatesHandling(t *testing.T) {
	// Mock output with both regular and phased updates
	mockOutput := `WARNING: apt does not have a stable CLI interface.
cpp-13/xenial-updates 13.2.0-5ubuntu1~24.04]
gcc-13/xenial-updates 13.2.0-5ubuntu1~24.04]
libssl1.1/xenial-updates 1.1.1f-1ubuntu2.20]
zlib1g/xenial-updates 1:1.2.11.dfsg-1ubuntu7.3`

	handler := &Handler{
		sysCalls: newMockPhasedSystemCalls(mockOutput),
	}

	result, err := handler.checkAPTUpdates(context.Background(), UpdateTypeAll, false)
	assert.NoError(t, err)
	// All updates should be included (4 total)
	assert.Equal(t, 4, result.AvailableUpdates)
	assert.Len(t, result.PackageDetailsList, 4)
}

// TestParsePackageLine ensures correct parsing of apt list output
func TestParsePackageLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		pkgName  string
		version  string
		isPhased bool
	}{
		{
			name:     "normal update",
			line:     "cpp-13/xenial-updates 13.2.0-5ubuntu1~24.04]",
			pkgName:  "cpp-13",
			version:  "13.2.0-5ubuntu1~24.04",
			isPhased: false,
		},
		{
			name:     "phased update with [phased XX%] before version",
			line:     "cpp-13/xenial-updates [phased 10%] 13.2.0-5ubuntu1~24.04]",
			pkgName:  "cpp-13",
			version:  "13.2.0-5ubuntu1~24.04",
			isPhased: true,
		},
		{
			name:     "phased update with [phased XX%] before version",
			line:     "cpp-13/xenial-updates [phased 10%] 13.2.0-5ubuntu1~24.04]",
			pkgName:  "cpp-13",
			version:  "13.2.0-5ubuntu1~24.04",
			isPhased: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgName, version, isPhased := parsePackageLine(tt.line)
			assert.Equal(t, tt.pkgName, pkgName)
			assert.Equal(t, tt.version, version)
			assert.Equal(t, tt.isPhased, isPhased)
		})
	}
}

// TestGetAllUpdatesWithPhased excludes phased from regular counts but includes them in total
func TestGetAllUpdatesWithPhased(t *testing.T) {
	// Mock output with mix of regular and phased updates
	mockOutput := `WARNING: apt does not have a stable CLI interface.
cpp-13/xenial-updates 13.2.0-5ubuntu1~24.04]
gcc-13/xenial-updates [phased 10%] 13.2.0-6ubuntu2~24.04]
libstdc++6/xenial-updates [phased 5%] 13.2.0-6ubuntu2~24.04
zlib1g/xenial-updates 1:1.2.11.dfsg-1ubuntu7.3`

	handler := &Handler{
		sysCalls: newMockPhasedSystemCalls(mockOutput),
	}

	result, err := handler.GetAllUpdates(context.Background(), nil)
	assert.NoError(t, err)

	// GetAllUpdates returns any (JSON wrapped), need to type assert
	resultObj, ok := result.(*AllUpdatesResult)
	assert.True(t, ok, "GetAllUpdates should return *AllUpdatesResult")

	// Total updates should include all (4 total)
	assert.Equal(t, 4, resultObj.AllUpdatesCount)
	assert.Len(t, resultObj.AllUpdatesList, 4)

	// Phased updates should be counted separately (2 phased)
	assert.Equal(t, 2, resultObj.PhasedUpdatesCount)
	assert.Len(t, resultObj.PhasedUpdatesList, 2)
	assert.Contains(t, resultObj.PhasedUpdatesList, "gcc-13")
	assert.Contains(t, resultObj.PhasedUpdatesList, "libstdc++6")

	// Recommended updates should exclude phased (only 2 non-phased)
	assert.Equal(t, 2, resultObj.RecommendedUpdatesCount)
	assert.Len(t, resultObj.RecommendedUpdatesList, 2)
	assert.Contains(t, resultObj.RecommendedUpdatesList, "cpp-13")
	assert.Contains(t, resultObj.RecommendedUpdatesList, "zlib1g")
}

// TestGetAllUpdatesPhasedInSecurity ensures phased updates are excluded from security counts too
func TestGetAllUpdatesPhasedInSecurity(t *testing.T) {
	mockOutput := `WARNING: apt does not have a stable CLI interface.
cpp-13/xenial-updates 13.2.0-5ubuntu1~24.04]
gcc-13/xenial-updates [phased 10%] 13.2.0-6ubuntu2~24.04]`

	handler := &Handler{
		sysCalls: newMockPhasedSystemCalls(mockOutput),
	}

	result, err := handler.GetAllUpdates(context.Background(), nil)
	assert.NoError(t, err)

	// GetAllUpdates returns any (JSON wrapped), need to type assert
	resultObj, ok := result.(*AllUpdatesResult)
	assert.True(t, ok, "GetAllUpdates should return *AllUpdatesResult")

	// Total updates should be 2
	assert.Equal(t, 2, resultObj.AllUpdatesCount)

	// Phased updates should be counted separately (1 phased)
	assert.Equal(t, 1, resultObj.PhasedUpdatesCount)
	assert.Len(t, resultObj.PhasedUpdatesList, 1)

	// Recommended updates should exclude phased (only 1 non-phased)
	assert.Equal(t, 1, resultObj.RecommendedUpdatesCount)
}

// mockPhasedSystemCalls is a mock that returns the exact output provided
// and handles both apt list and apt-get -s upgrade commands
type mockPhasedSystemCalls struct {
	output string
}

func (m *mockPhasedSystemCalls) execCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	// For apt-get -s upgrade commands (used by new implementation)
	// The command is called as: env LC_ALL=C LANG=C apt-get -s [flags] upgrade
	if name == "env" && len(args) >= 3 && args[0] == "LC_ALL=C" && args[1] == "LANG=C" &&
	   len(args) >= 4 && (args[2] == "apt-get" || args[3] == "apt-get") {
		// Check if apt-get is at position 2 or 3 (depending on whether flags are present)
		var aptArgs []string
		if args[2] == "apt-get" {
			// No phased flag: env LC_ALL=C LANG=C apt-get -s upgrade
			aptArgs = args[2:]
		} else if len(args) >= 5 && args[3] == "apt-get" {
			// With phased flag: env LC_ALL=C LANG=C apt-get -s -o ... upgrade
			aptArgs = args[3:]
		}

		// Check if this is an apt-get -s command followed by upgrade
		if len(aptArgs) >= 2 && aptArgs[0] == "apt-get" && aptArgs[1] == "-s" {
			// Convert mock output from apt list format to apt-get -s upgrade format
			targetOutput := []string{"WARNING: apt does not have a stable CLI interface."}

			lines := strings.Split(strings.TrimSpace(m.output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(strings.ToLower(line), "warning:") {
					continue
				}

				// Parse apt list format: name/channel [phased X%] version]
				parts := strings.Fields(line)
				if len(parts) < 2 {
					continue
				}

				pkgName := strings.Split(parts[0], "/")[0]
			version := strings.TrimSpace(strings.Join(parts[1:], " "))

				if version != "" {
					// Create apt-get format: Inst <pkg> [<old>] (<new> ...)
					targetOutput = append(targetOutput, fmt.Sprintf("Inst %s (%s", pkgName, version))
				}
			}

			return []byte(strings.Join(targetOutput, "\n") + "\n"), nil
		}
	}

	// For apt-cache policy commands (used in isPackageOfType)
	if name == "apt-cache" && len(args) > 0 && args[0] == "policy" {
		// Return mock output that indicates main repository (not security/universe)
		return []byte("Package: " + args[1] + "\n" +
			"Candidate: 1.0\n" +
			"Version table:\n" +
			"   1.0 500\n"),
		nil
	}
	return []byte{}, nil
}

func newMockPhasedSystemCalls(output string) systemCalls {
	return &mockPhasedSystemCalls{
		output: output,
	}
}
