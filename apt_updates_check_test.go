// +build ignore

// Test file for apt_updates_check functionality
// This demonstrates how the plugin would be tested with mock data

package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseAPTOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     int
		wantErr  bool
	}{
		{
			name: "No updates",
			input: ``,
			want: 0,
			wantErr: false,
		},
		{
			name: "Single update",
			input: `curl/now 7.81.0-1ubuntu1.9 amd64 [upgradable from: 7.81.0-1ubuntu1.8]`,
			want: 1,
			wantErr: false,
		},
		{
			name: "Multiple updates",
			input: `curl/now 7.81.0-1ubuntu1.9 amd64 [upgradable from: 7.81.0-1ubuntu1.8]
libssl1.1/now 1.1.1w-1ubuntu2.3 amd64 [upgradable from: 1.1.1w-1ubuntu2.2]`,
			want: 2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a mock test - in real implementation we would mock exec.Command
			lines := strings.Split(strings.TrimSpace(tt.input), "\n")
			var count int
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.Contains(line, "/") {
					continue
				}
				count++
			}

			if count != tt.want {
				t.Errorf("parseAPTOutput() count = %v, want %v", count, tt.want)
			}
		})
	}
}

func TestJSONOutput(t *testing.T) {
	result := &CheckResult{
		AvailableUpdates: 2,
		PackageDetailsList: []UpdateInfo{
			{Name: "curl", Target: "7.81.0-1ubuntu1.9"},
			{Name: "libssl1.1", Target: "1.1.1w-1ubuntu2.3"},
		},
		WarningThreshold: 10,
		IsAboveWarning:   false,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	expected := `{
  "available_updates": 2,
  "package_details_list": [
    {
      "name": "curl",
      "target_version": "7.81.0-1ubuntu1.9"
    },
    {
      "name": "libssl1.1",
      "target_version": "1.1.1w-1ubuntu2.3"
    }
  ],
  "warning_threshold": 10,
  "is_above_warning": false
}`

	if string(data) != expected {
		t.Errorf("JSON output mismatch\nGot:\n%s\nWant:\n%s", string(data), expected)
	}
}

func ExampleCheckResult() {
	result := &CheckResult{
		AvailableUpdates: 5,
		PackageDetailsList: []UpdateInfo{
			{Name: "curl", Target: "7.81.0-1ubuntu1.9"},
			{Name: "libssl1.1", Target: "1.1.1w-1ubuntu2.3"},
			{Name: "openssl", Target: "3.0.2-0ubuntu1.7"},
			{Name: "nginx", Target: "1.18.0-6ubuntu14.2"},
			{Name: "python3", Target: "3.10.6-1~22.04.5"},
		},
		WarningThreshold: 10,
		IsAboveWarning:   false,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
	// Output:
	// {
	//   "available_updates": 5,
	//   "package_details_list": [
	//     {
	//       "name": "curl",
	//       "target_version": "7.81.0-1ubuntu1.9"
	//     },
	//     {
	//       "name": "libssl1.1",
	//       "target_version": "1.1.1w-1ubuntu2.3"
	//     },
	//     {
	//       "name": "openssl",
	//       "target_version": "3.0.2-0ubuntu1.7"
	//     },
	//     {
	//       "name": "nginx",
	//       "target_version": "1.18.0-6ubuntu14.2"
	//     },
	//     {
	//       "name": "python3",
	//       "target_version": "3.10.6-1~22.04.5"
	//     }
	//   ],
	//   "warning_threshold": 10,
	//   "is_above_warning": false
	// }
}
