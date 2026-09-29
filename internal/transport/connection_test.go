// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package transport_test

import (
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/transport"
	"gopkg.in/yaml.v3"
)

func TestConnection_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		input   string
		wantErr bool
		want    *transport.Connection
	}{
		{
			name:    "nil input",
			input:   "",
			wantErr: false,
			want:    &transport.Connection{},
		},
		{
			name: "error transport type",
			input: `
dummmy@type: {}
`,
			wantErr: true,
		},
		{
			name: "invalid data",
			input: `
- 0
- 1
- 2
`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c transport.Connection
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
			if !reflect.DeepEqual(&c, tt.want) {
				t.Errorf("UnmarshalYAML() = %v, want %v", &c, tt.want)
			}
		})
	}
}
