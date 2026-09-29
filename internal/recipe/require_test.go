// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_test

import (
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/recipe"
	"gopkg.in/yaml.v3"
)

func TestRequireList_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		src  string // YAML source to unmarshal
		// Named input parameters for target function.
		wantErr bool
		result  recipe.RequireList // expected result after unmarshalling
	}{
		{
			name: "string list",
			src: `
- foo
- bar
`,
			wantErr: false,
			result: recipe.RequireList{
				&recipe.Require{Name: "foo"},
				&recipe.Require{Name: "bar"},
			},
		},
		{
			name: "map list",
			src: `
- name: foo
  as: foo_alias
- name: bar
`,
			wantErr: false,
			result: recipe.RequireList{
				&recipe.Require{Name: "foo", As: "foo_alias"},
				&recipe.Require{Name: "bar"},
			},
		},
		{
			name:    "empty list",
			src:     `[]`,
			wantErr: false,
			result:  recipe.RequireList{},
		},
		{
			name:    "invalid format",
			src:     `foo`,
			wantErr: true,
		},
		{
			name:    "invalid item format",
			src:     `- name: foo\n  as: foo_alias\n- invalid_item`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got recipe.RequireList
			gotErr := yaml.Unmarshal([]byte(tt.src), &got)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UnmarshalYAML() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UnmarshalYAML() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.result) {
				t.Errorf("UnmarshalYAML() = %#v, want %#v", got, tt.result)
			}
		})
	}
}
