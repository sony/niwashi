// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package file_test

import (
	"testing"

	"github.com/sony/niwashi/internal/file"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		str  string
		want string
	}{
		{
			name: "replaces @ with _",
			str:  "file@name.txt",
			want: "file_name.txt",
		},
		{
			name: "replaces recipe fqid",
			str:  "nws/recipe.recipe.recipe@0.1.2",
			want: "nws_recipe.recipe.recipe_0.1.2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := file.SanitizeFilename(tt.str)
			if got != tt.want {
				t.Errorf("SanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
