// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package instance_test

import (
	"testing"

	"github.com/sony/niwashi/internal/instance"
)

func TestGenerator_Validate(t *testing.T) {
	tests := []struct {
		name      string // description of this test case
		generator *instance.Generator
		wantErr   bool
	}{
		{
			name:      "empty",
			generator: &instance.Generator{},
			wantErr:   true,
		},
		{
			name: "invalid store",
			generator: &instance.Generator{
				Provisioner: "provisioner",
				Store: map[string]any{
					"invalid-name": "value", // ng
					"invalid.name": "value", // ng
					"invalid name": "value", // ng
					"0":            "value", // ng
					".":            "value", // ng
					"a":            "value", // ok
					"b": map[string]any{
						"x x x x": "value", // ng
						"y.y":     "value", // ng
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.generator
			gotErr := g.Validate()
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
