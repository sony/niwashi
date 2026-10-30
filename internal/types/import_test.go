// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package types_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/types"
	"gopkg.in/yaml.v3"
)

func TestImport_Validate(t *testing.T) {
	tests := []struct {
		name    string
		imp     *types.Import
		wantErr bool
	}{
		{name: "from is set", imp: &types.Import{From: "proxy_ca_cert"}},
		{name: "optional with from", imp: &types.Import{From: "proxy_ca_cert", Optional: true}},
		{name: "empty from", imp: &types.Import{}, wantErr: true},
		{name: "optional without from", imp: &types.Import{Optional: true}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.imp.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImports_Validate(t *testing.T) {
	tests := []struct {
		name    string
		imports types.Imports
		wantErr string // empty means no error
	}{
		{
			name:    "all valid",
			imports: types.Imports{"cert": {From: "proxy_ca_cert"}, "endpoint": {From: "proxy_endpoint", Optional: true}},
		},
		{
			name:    "one has an empty from",
			imports: types.Imports{"cert": {From: "proxy_ca_cert"}, "endpoint": {}},
			wantErr: `import "endpoint"`,
		},
		{
			// a YAML null value, e.g. "cert:" with nothing after it
			name:    "nil entry",
			imports: types.Imports{"cert": nil},
			wantErr: `import "cert"`,
		},
		{name: "empty", imports: types.Imports{}},
		{name: "nil", imports: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.imports.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("Validate() succeeded unexpectedly")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate() error = %v, want %q in it", err, tt.wantErr)
			}
		})
	}
}

func TestImports_Merge(t *testing.T) {
	tests := []struct {
		name  string
		base  types.Imports
		other types.Imports
		want  types.Imports
	}{
		{
			// the later definition replaces the earlier one as a whole (029 3-3)
			name:  "same param",
			base:  types.Imports{"cert": {From: "old"}},
			other: types.Imports{"cert": {From: "new", Optional: true}},
			want:  types.Imports{"cert": {From: "new", Optional: true}},
		},
		{
			// optional is not kept from the earlier definition
			name:  "later definition without optional",
			base:  types.Imports{"cert": {From: "old", Optional: true}},
			other: types.Imports{"cert": {From: "new"}},
			want:  types.Imports{"cert": {From: "new"}},
		},
		{
			name:  "different params",
			base:  types.Imports{"cert": {From: "proxy_ca_cert"}},
			other: types.Imports{"endpoint": {From: "proxy_endpoint"}},
			want:  types.Imports{"cert": {From: "proxy_ca_cert"}, "endpoint": {From: "proxy_endpoint"}},
		},
		{
			name:  "nil base",
			base:  nil,
			other: types.Imports{"cert": {From: "new"}},
			want:  types.Imports{"cert": {From: "new"}},
		},
		{
			name:  "nil other",
			base:  types.Imports{"cert": {From: "old"}},
			other: nil,
			want:  types.Imports{"cert": {From: "old"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.base.Merge(tt.other)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Merge() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("does not share entries with the inputs", func(t *testing.T) {
		base := types.Imports{"cert": {From: "old"}}
		other := types.Imports{"endpoint": {From: "proxy_endpoint"}}
		got := base.Merge(other)
		got["cert"].From = "changed"
		got["endpoint"].From = "changed"
		if base["cert"].From != "old" || other["endpoint"].From != "proxy_endpoint" {
			t.Errorf("Merge() result shares entries with the inputs: base=%v other=%v", base, other)
		}
	})
}

// decodeImports decodes src as the loader does, with unknown fields rejected.
func decodeImports(src string) (types.Imports, error) {
	var v struct {
		Import types.Imports `yaml:"import"`
	}
	d := yaml.NewDecoder(bytes.NewBufferString(src))
	d.KnownFields(true)
	err := d.Decode(&v)
	return v.Import, err
}

func TestImport_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		want    types.Imports
		wantErr bool
	}{
		{
			name: "short form",
			src:  "import:\n  cert: proxy_ca_cert\n",
			want: types.Imports{"cert": {From: "proxy_ca_cert"}},
		},
		{
			name: "mapping",
			src:  "import:\n  cert: {from: proxy_ca_cert, optional: true}\n",
			want: types.Imports{"cert": {From: "proxy_ca_cert", Optional: true}},
		},
		{
			name: "short form and mapping",
			src:  "import:\n  cert: proxy_ca_cert\n  endpoint:\n    from: proxy_endpoint\n",
			want: types.Imports{"cert": {From: "proxy_ca_cert"}, "endpoint": {From: "proxy_endpoint"}},
		},
		{
			// node.Decode does not inherit KnownFields(true), so UnmarshalYAML checks the keys
			name:    "unknown field",
			src:     "import:\n  cert: {form: proxy_ca_cert}\n",
			wantErr: true,
		},
		{
			name:    "sequence",
			src:     "import:\n  cert: [a, b]\n",
			wantErr: true,
		},
		{
			name:    "optional is not a bool",
			src:     "import:\n  cert: {from: proxy_ca_cert, optional: maybe}\n",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeImports(tt.src)
			if (err != nil) != tt.wantErr {
				t.Fatalf("decode error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decoded = %v, want %v", got, tt.want)
			}
		})
	}
}
