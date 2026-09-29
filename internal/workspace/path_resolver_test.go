// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workspace

import "testing"

func TestSeparatorResolver_Join(t *testing.T) {
	tests := []struct {
		name      string
		separator string
		paths     []string
		want      string
	}{
		{
			name:      "windows separator",
			separator: "\\",
			paths:     []string{"C:\\Users\\testuser", ".niwashi"},
			want:      "C:\\Users\\testuser\\.niwashi",
		},
		{
			name:      "unix separator",
			separator: "/",
			paths:     []string{"/home/testuser", ".niwashi"},
			want:      "/home/testuser/.niwashi",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewSeparatorResolver(tt.separator)
			got := r.Join(tt.paths...)
			if got != tt.want {
				t.Errorf("Join() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_render(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		t    string
		p    any
		want string
	}{
		{
			name: "renders a simple template",
			t:    "Hello, {{.Name}}!",
			p:    map[string]string{"Name": "World"},
			want: "Hello, World!",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := render(tt.t, tt.p)
			if got != tt.want {
				t.Errorf("render() = %v, want %v", got, tt.want)
			}
		})
	}
}
