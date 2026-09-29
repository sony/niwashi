// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_test

import (
	"testing"

	"github.com/sony/niwashi/internal/recipe"
)

func TestToolRunSpec_Validate(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		spec    *recipe.ToolRunSpec
		wantErr bool
	}{
		{
			name:    "empty",
			spec:    &recipe.ToolRunSpec{},
			wantErr: true, // toolRef and command are required.
		},
		{
			name: "invalid env name",
			spec: &recipe.ToolRunSpec{
				ToolRef: "tool1",
				Command: "echo hello",
				EnvTemplate: map[string]string{
					"":             "value", // ng
					"invalid-name": "value", // ng
					"invalid.name": "value", // ng
					"invalid name": "value", // ng
					"0":            "value", // ng
					".":            "value", // ng
					"a":            "value", // ok
				},
			},
			wantErr: true, // invalid env names.
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.spec.Validate()
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
