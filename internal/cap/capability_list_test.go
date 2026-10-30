// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cap_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/types"
	"gopkg.in/yaml.v3"
)

func TestCapabilityList_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		input    string
		expected cap.CapabilityList
		wantErr  bool
	}{
		{
			name:     "empty",
			input:    "",
			expected: nil,
			wantErr:  false,
		},
		{
			name: "string list style",
			input: `
- cap1
- cap2
- cap3
`,
			expected: cap.CapabilityList{
				"cap1": &cap.Capability{},
				"cap2": &cap.Capability{},
				"cap3": &cap.Capability{},
			},
			wantErr: false,
		},
		{
			name: "mix string list and object style",
			input: `
- cap1
- id: cap2
  version: 0.0.2
- cap3
`,
			expected: cap.CapabilityList{
				"cap1": &cap.Capability{},
				"cap2": &cap.Capability{Version: "0.0.2"},
				"cap3": &cap.Capability{},
			},
			wantErr: false,
		},
		{
			name: "object style",
			input: `
cap1:
  version: 0.0.1
cap2:
  version: 0.0.2
cap3:
  version: 0.0.3
`,
			expected: cap.CapabilityList{
				"cap1": &cap.Capability{Version: "0.0.1"},
				"cap2": &cap.Capability{Version: "0.0.2"},
				"cap3": &cap.Capability{Version: "0.0.3"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c cap.CapabilityList
			gotErr := yaml.Unmarshal([]byte(tt.input), &c)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UnmarshalYAML() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UnmarshalYAML() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(c, tt.expected) {
				t.Errorf("UnmarshalYAML() = %v, want %v", c, tt.expected)
			}
		})
	}
}

// importRecipeFinder resolves "ca-trust" (alias), "org/ca-trust" and
// "org/ca-trust@1.0.0" to the same recipe.
type importRecipeFinder struct{}

func (importRecipeFinder) FindRecipe(n string) (*recipe.Recipe, error) {
	switch n {
	case "ca-trust", "org/ca-trust", "org/ca-trust@1.0.0":
		return &recipe.Recipe{Metadata: &recipe.Metadata{Id: "org/ca-trust", Version: "1.0.0"}}, nil
	}
	return nil, fmt.Errorf("recipe not found: %q", n)
}

func (importRecipeFinder) FindRecipeAsAdapter(n string) (*recipe.Recipe, error) {
	return nil, fmt.Errorf("recipe not found: %q", n)
}

// TestCapabilityList_Import_DecodeAndNormalize checks that import survives
// decoding and normalization, which rebuild each capability.
func TestCapabilityList_Import_DecodeAndNormalize(t *testing.T) {
	wantImport := types.Imports{
		"cert":     {From: "proxy_ca_cert"},
		"endpoint": {From: "proxy_endpoint", Optional: true},
	}
	wantParams := types.Params{"update": true}

	tests := []struct {
		name  string
		input string
		key   string // key after decoding
	}{
		{
			name: "list style",
			input: `
- id: ca-trust
  params:
    update: true
  import:
    cert: proxy_ca_cert
    endpoint:
      from: proxy_endpoint
      optional: true
`,
			key: "ca-trust",
		},
		{
			name: "list style with FQID",
			input: `
- id: org/ca-trust@1.0.0
  params:
    update: true
  import:
    cert: proxy_ca_cert
    endpoint: {from: proxy_endpoint, optional: true}
`,
			key: "org/ca-trust@1.0.0",
		},
		{
			name: "map style",
			input: `
org/ca-trust:
  params:
    update: true
  import:
    cert: proxy_ca_cert
    endpoint: {from: proxy_endpoint, optional: true}
`,
			key: "org/ca-trust",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l cap.CapabilityList
			if err := yaml.Unmarshal([]byte(tt.input), &l); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			decoded, ok := l[tt.key]
			if !ok {
				t.Fatalf("capability %q not found after decoding: %v", tt.key, l)
			}
			if !reflect.DeepEqual(decoded.Import, wantImport) {
				t.Errorf("Import after decoding = %v, want %v", decoded.Import, wantImport)
			}

			normalized, err := l.Normalize(importRecipeFinder{})
			if err != nil {
				t.Fatalf("Normalize() error = %v", err)
			}
			got, ok := normalized["org/ca-trust"]
			if !ok {
				t.Fatalf("capability %q not found after normalization: %v", "org/ca-trust", normalized)
			}
			if !reflect.DeepEqual(got.Import, wantImport) {
				t.Errorf("Import after normalization = %v, want %v", got.Import, wantImport)
			}
			if !reflect.DeepEqual(got.Params, wantParams) {
				t.Errorf("Params after normalization = %v, want %v", got.Params, wantParams)
			}
			if got.Version != "1.0.0" {
				t.Errorf("Version after normalization = %q, want %q", got.Version, "1.0.0")
			}
		})
	}
}
