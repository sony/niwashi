// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cap_test

import (
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/cap"
	"gopkg.in/yaml.v3"
)

func TestCapabilityList_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		input    string
		expected cap.CapabilityList
		wantErr  bool
	}{
		{
			name:     "empty",
			input:    "",
			expected: nil,
			wantErr:  false,
		},
		{
			name: "string list style",
			input: `
- cap1
- cap2
- cap3
`,
			expected: cap.CapabilityList{
				"cap1": &cap.Capability{},
				"cap2": &cap.Capability{},
				"cap3": &cap.Capability{},
			},
			wantErr: false,
		},
		{
			name: "mix string list and object style",
			input: `
- cap1
- id: cap2
  version: 0.0.2
- cap3
`,
			expected: cap.CapabilityList{
				"cap1": &cap.Capability{},
				"cap2": &cap.Capability{Version: "0.0.2"},
				"cap3": &cap.Capability{},
			},
			wantErr: false,
		},
		{
			name: "object style",
			input: `
cap1:
  version: 0.0.1
cap2:
  version: 0.0.2
cap3:
  version: 0.0.3
`,
			expected: cap.CapabilityList{
				"cap1": &cap.Capability{Version: "0.0.1"},
				"cap2": &cap.Capability{Version: "0.0.2"},
				"cap3": &cap.Capability{Version: "0.0.3"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c cap.CapabilityList
			gotErr := yaml.Unmarshal([]byte(tt.input), &c)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UnmarshalYAML() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UnmarshalYAML() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(c, tt.expected) {
				t.Errorf("UnmarshalYAML() = %v, want %v", c, tt.expected)
			}
		})
	}
}
