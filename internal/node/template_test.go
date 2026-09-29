// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package node_test

import (
	"testing"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/node"
)

func TestTemplate_Validate(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		tpl     *node.Template
		wantErr bool
	}{
		{
			name:    "empty template",
			tpl:     &node.Template{},
			wantErr: false,
		},
		{
			name: "invalid capability param",
			tpl: &node.Template{
				Capabilities: cap.CapabilityList{
					"cap1": &cap.Capability{
						Params: map[string]any{
							"invalid-name": "value", // ng
							"invalid.name": "value", // ng
							"invalid name": "value", // ng
							"0":            "value", // ng
							".":            "value", // ng
							"a":            "value", // ok
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid label key",
			tpl: &node.Template{
				Labels: map[string]string{
					"invalid-key": "value", // ng
					"invalid.key": "value", // ng
					"invalid key": "value", // ng
					"0":           "value", // ng
					".":           "value", // ng
					"a":           "value", // ok
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.tpl.Validate()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Validate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Validate() succeeded unexpectedly")
			}
		})
	}
}
