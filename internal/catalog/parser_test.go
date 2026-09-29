// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package catalog

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		reader io.Reader
	}
	tests := []struct {
		name string
		args args
		want *Catalog
	}{
		{
			name: "empty",
			args: args{
				reader: strings.NewReader(""),
			},
			want: &Catalog{},
		},
		{
			name: "unmarshal error #1",
			args: args{
				reader: strings.NewReader("version: [invalid"),
			},
			want: nil,
		},
		{
			name: "unmarshal error #2",
			args: args{
				reader: strings.NewReader(`
version:
  - 1
  - 2
`),
			},
			want: nil,
		},
		{
			name: "compact",
			args: args{
				reader: strings.NewReader(`
version: nws.catalog/v1
`),
			},
			want: &Catalog{
				Version: Version,
				Recipes: nil,
			},
		},
		{
			name: "normal",
			args: args{
				reader: strings.NewReader(`
version: nws.catalog/v1
recipes:
  a:
    path: a/b/c
  b:
    path: x/y/z
`),
			},
			want: &Catalog{
				Version: Version,
				Recipes: map[string]Recipe{
					"a": Recipe{
						Path: "a/b/c",
					},
					"b": Recipe{
						Path: "x/y/z",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Parse(tt.args.reader); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
