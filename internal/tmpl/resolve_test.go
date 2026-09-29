// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package tmpl_test

import (
	"testing"

	"github.com/sony/niwashi/internal/tmpl"
)

type mockTemplate struct {
	from  string
	value string
}

func (m *mockTemplate) Merge(other *mockTemplate) (*mockTemplate, error) {
	return &mockTemplate{
		value: m.value + other.value,
	}, nil
}
func (m *mockTemplate) Clone() *mockTemplate {
	return &mockTemplate{
		value: m.value,
		from:  m.from,
	}
}
func (m *mockTemplate) FromField() string {
	return m.from
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		tpl     map[string]*mockTemplate
		base    *mockTemplate
		want    *mockTemplate
		wantErr bool
	}{
		{
			name: "simple case",
			tpl: map[string]*mockTemplate{
				"base":  {value: "base"},
				"dummy": {value: "dummy", from: "base"},
			},
			base: &mockTemplate{from: "base"},
			want: &mockTemplate{
				value: "base",
			},
			wantErr: false,
		},
		{
			name: "multiple inheritance",
			tpl: map[string]*mockTemplate{
				"v1": {value: "v1", from: "v2"},
				"v2": {value: "v2", from: "v3"},
				"v3": {value: "v3"},
			},
			base: &mockTemplate{from: "v1"},
			want: &mockTemplate{
				value: "v3v2v1",
			},
			wantErr: false,
		},
		{
			name: "circular reference",
			tpl: map[string]*mockTemplate{
				"v1": {value: "v1", from: "v2"},
				"v2": {value: "v2", from: "v1"},
			},
			base:    &mockTemplate{from: "v1"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := tmpl.Resolve(tt.tpl, tt.base)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Resolve() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Resolve() succeeded unexpectedly")
			}
			if got.value != tt.want.value {
				t.Errorf("Resolve() = %v, want %v", got, tt.want)
			}
		})
	}
}
