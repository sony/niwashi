// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier_test

import (
	"testing"

	"github.com/sony/niwashi/internal/workflow/applier"
)

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		input   string
		val     any
		want    string
		wantErr bool
	}{
		{
			name:  "case 1: normal",
			input: `Hello, {{ .Name }}!`,
			val: map[string]string{
				"Name": "World",
			},
			want:    `Hello, World!`,
			wantErr: false,
		},
		{
			name:  "case 2: Outputs transform",
			input: `The output path is {{ .Outputs.path.to.output }}`,
			val: map[string]string{
				"Outputs": "./custom/root",
			},
			want:    `The output path is ./custom/root/path/to/output`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := applier.RenderTemplate(tt.input, tt.val)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("RenderTemplate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("RenderTemplate() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("RenderTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}
