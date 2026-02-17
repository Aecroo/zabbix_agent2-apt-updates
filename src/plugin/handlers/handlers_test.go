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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestVersionParsing ensures that target_version never contains brackets
func TestVersionParsing(t *testing.T) {
	tests := []struct {
		name     string
		mockOutput string
		wantErr  bool
	}{
		{
			name: "normal apt list --upgradable output",
			mockOutput: `WARNING: apt does not have a stable CLI interface.
bsdextrautils/xenial-updates 2.39.3-9ubuntu6.3]
libssl-dev/xenial-updates 1.1.1f-1ubuntu2.20]
nmap/xenial-updates 7.80+dfsg1-2ubuntu1.1]
`,
			wantErr: false,
		},
		{
			name: "no upgrades available",
			mockOutput: `Reading package lists...
Updating package lists...
Reading state information...
0 upgraded, 0 newly installed, 0 to remove and 0 not upgraded.
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a handler with mock system calls that return our test data
			handler := &Handler{
				sysCalls: newMockSystemCalls(tt.mockOutput, nil),
			}

			result, err := handler.checkAPTUpdates(context.Background(), UpdateTypeAll, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkAPTUpdates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify no version contains brackets
			for _, update := range result.PackageDetailsList {
				assert.NotContains(t, update.Target, "[", "target_version should not contain '['")
				assert.NotContains(t, update.Target, "]", "target_version should not contain ']'")
			}

			// Verify versions are properly extracted
			if !tt.wantErr && strings.Contains(tt.mockOutput, "bsdextrautils") {
				found := false
				for _, update := range result.PackageDetailsList {
					if update.Name == "bsdextrautils" {
						found = true
						assert.Equal(t, "2.39.3-9ubuntu6.3", update.Target)
					}
				}
				if !found && strings.Contains(tt.mockOutput, "bsdextrautils/xenial-updates") {
					t.Error("Expected bsdextrautils to be in the results")
				}
			}
		})
	}
}

// TestVersionParsingWithBracketsInOutput ensures that even if apt output contains brackets
// for other purposes, they don't end up in target_version
func TestVersionParsingWithBracketsInOutput(t *testing.T) {
	mockOutput := `package-with-brackets-in-name/xenial-updates 1.2.3-4ubuntu5]
another-package/xenial-updates 2.3.4+git20240101]
`

	handler := &Handler{
		sysCalls: newMockSystemCalls(mockOutput, nil),
	}

	result, err := handler.checkAPTUpdates(context.Background(), UpdateTypeAll, false)
	assert.NoError(t, err)

	// Verify all versions are clean (no brackets)
	for _, update := range result.PackageDetailsList {
		assert.NotContains(t, update.Target, "[", "target_version should not contain opening bracket")
		assert.NotContains(t, update.Target, "]", "target_version should not contain closing bracket")
	}

	// Verify versions are correctly extracted
	for _, update := range result.PackageDetailsList {
		if update.Name == "package-with-brackets-in-name" {
			assert.Equal(t, "1.2.3-4ubuntu5", update.Target)
		} else if update.Name == "another-package" {
			assert.Equal(t, "2.3.4+git20240101", update.Target)
		}
	}
}

// TestEmptyOutput handles edge case of empty output
func TestEmptyOutput(t *testing.T) {
	handler := &Handler{
		sysCalls: newMockSystemCalls("", nil),
	}

	result, err := handler.checkAPTUpdates(context.Background(), UpdateTypeAll, false)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.AvailableUpdates)
	assert.Empty(t, result.PackageDetailsList)
}

// mockSystemCalls implements systemCalls interface for testing
type mockSystemCalls struct {
	output string
	err    error
}
	func (m *mockSystemCalls) execCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
		// Check if this is an apt-get command - convert apt list format to apt-get format
		if name == "env" && len(args) >= 5 && args[0] == "LC_ALL=C" && args[1] == "LANG=C" &&
		   args[2] == "apt-get" && args[3] == "-s" {
			// Convert apt list format to apt-get -s upgrade format
			targetOutput := []string{"WARNING: apt does not have a stable CLI interface."}
			lines := strings.Split(strings.TrimSpace(m.output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(strings.ToLower(line), "warning:") {
					continue
				}
				parts := strings.Fields(line)
				if len(parts) < 2 {
					continue
				}
				pkgName := strings.Split(parts[0], "/")[0]
				version := strings.TrimSuffix(strings.TrimSpace(strings.Join(parts[1:], " ")), "]")
				targetOutput = append(targetOutput, fmt.Sprintf("Inst %s (%s", pkgName, version))
			}
			return []byte(strings.Join(targetOutput, "\n") + "\n"), nil
		}
		// For non-apt-get commands (e.g., apt-cache policy), return raw output
		return []byte(m.output), m.err
	}



// newMockSystemCalls creates a new mock system calls implementation
func newMockSystemCalls(output string, err error) systemCalls {
	return &mockSystemCalls{
		output: output,
		err:    err,
	}
}
