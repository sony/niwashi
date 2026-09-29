// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_test

import (
	"testing"

	"github.com/sony/niwashi/internal/recipe"
)

func TestProvide_Validate(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		p       *recipe.Provide
		wantErr bool
	}{
		{
			name: "valid provide",
			p: &recipe.Provide{
				Name: "host.tool.abcdefg-hijklmn",
				Attributes: map[string]string{
					"by": "package-manager",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid name",
			p: &recipe.Provide{
				Name: "InvalidName", // ng: uppercase letters are not allowed
			},
			wantErr: true,
		},
		{
			name: "invalid attribute key",
			p: &recipe.Provide{
				Name: "valid.name",
				Attributes: map[string]string{
					"InvalidKey": "value", // ng: uppercase letters are not allowed in keys
				},
			},
			wantErr: true,
		},
		{
			name: "invalid attribute value",
			p: &recipe.Provide{
				Name: "valid.name",
				Attributes: map[string]string{
					"valid_key": "InvalidValue", // ng: uppercase letters are not allowed in values
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.p.Validate()
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
