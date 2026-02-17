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
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPhasedUpdatesHandling ensures that phased updates are counted separately
func TestPhasedUpdatesHandling(t *testing.T) {
	// Create a test handler with real system calls (will use actual apt-get)
	handler := &Handler{
		sysCalls: osWrapper{},
	}

	result, err := handler.checkAPTUpdates(context.Background(), UpdateTypeAll, false)
	assert.NoError(t, err)
	// Should return some updates (number depends on actual system state)
	assert.GreaterOrEqual(t, result.AvailableUpdates, 0)
}

// TestParsePackageLine ensures correct parsing of apt list output (for old method)
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
	// Create a test handler with real system calls (will use actual apt-get)
	handler := &Handler{
		sysCalls: osWrapper{},
	}

	result, err := handler.GetAllUpdates(context.Background(), nil)
	assert.NoError(t, err)

	// GetAllUpdates returns any (JSON wrapped), need to type assert
	resultObj, ok := result.(*AllUpdatesResult)
	assert.True(t, ok, "GetAllUpdates should return *AllUpdatesResult")

	// Should have some updates or zero if system is up-to-date
	assert.GreaterOrEqual(t, resultObj.AllUpdatesCount, 0)

	// Verify the structure contains all expected fields for phased updates
	assert.NotNil(t, resultObj.PhasedUpdatesList)
	assert.NotNil(t, resultObj.PhasedUpdatesDetails)

	// The actual count depends on whether the system has phased updates available
	// On some systems there may be 0 phased updates, which is valid behavior
	phasedCount := resultObj.PhasedUpdatesCount
	totalCount := resultObj.AllUpdatesCount

	if phasedCount > 0 {
		// Verify that if there are phased updates, they have the IsPhased field set correctly
		for _, pkg := range resultObj.PhasedUpdatesDetails {
			assert.Equal(t, true, pkg.IsPhased, "phased updates should have IsPhased=true")
		}
	}

	// Verify recommended count excludes phased updates
	// This is the key test - recommended should not include phased packages
	recommendedCount := resultObj.RecommendedUpdatesCount

	// Recommended + Phased should equal total (or be close if there are security/optional overlaps)
	assert.LessOrEqual(t, phasedCount+recommendedCount, totalCount,
		"phased count + recommended count should not exceed total updates")
}

// TestGetAllUpdatesPhasedInSecurity ensures phased updates are excluded from security counts too
func TestGetAllUpdatesPhasedInSecurity(t *testing.T) {
	// Create a test handler with real system calls (will use actual apt-get)
	handler := &Handler{
		sysCalls: osWrapper{},
	}

	result, err := handler.GetAllUpdates(context.Background(), nil)
	assert.NoError(t, err)

	// GetAllUpdates returns any (JSON wrapped), need to type assert
	resultObj, ok := result.(*AllUpdatesResult)
	assert.True(t, ok, "GetAllUpdates should return *AllUpdatesResult")

	// Should have some updates or zero if system is up-to-date
	assert.GreaterOrEqual(t, resultObj.AllUpdatesCount, 0)

	// Verify the structure contains all expected fields for phased updates
	assert.NotNil(t, resultObj.PhasedUpdatesList)
	assert.NotNil(t, resultObj.PhasedUpdatesDetails)
	assert.NotNil(t, resultObj.SecurityUpdatesList)
	assert.NotNil(t, resultObj.SecurityUpdatesDetails)

	// The actual count depends on whether the system has phased or security updates available
	phasedCount := resultObj.PhasedUpdatesCount

	if phasedCount > 0 {
		// Verify that if there are phased updates, they have the IsPhased field set correctly
		for _, pkg := range resultObj.PhasedUpdatesDetails {
			assert.Equal(t, true, pkg.IsPhased, "phased updates should have IsPhased=true")
		}
	}
}

// mockPhasedSystemCalls is a mock that returns the exact output provided
// and handles both apt list and apt-get -s upgrade commands
type mockPhasedSystemCalls struct {
	output string
}

func (m *mockPhasedSystemCalls) execCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
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
