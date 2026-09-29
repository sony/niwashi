// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package catalog

import "testing"

func TestCatalog_GetVersion(t *testing.T) {
	type fields struct {
		Version string
		Recipes map[string]Recipe
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "base",
			fields: fields{
				Version: "aaa",
				Recipes: nil,
			},
			want: "aaa",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Catalog{
				Version: tt.fields.Version,
				Recipes: tt.fields.Recipes,
			}
			if got := s.GetVersion(); got != tt.want {
				t.Errorf("Catalog.GetVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
