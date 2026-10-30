// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package instance_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sony/niwashi/internal/instance"
	"github.com/sony/niwashi/internal/types"
)

func TestGenerator_Validate(t *testing.T) {
	tests := []struct {
		name      string // description of this test case
		generator *instance.Generator
		wantErr   bool
	}{
		{
			name:      "empty",
			generator: &instance.Generator{},
			wantErr:   true,
		},
		{
			name: "invalid store",
			generator: &instance.Generator{
				Provisioner: "provisioner",
				Store: map[string]any{
					"invalid-name": "value", // ng
					"invalid.name": "value", // ng
					"invalid name": "value", // ng
					"0":            "value", // ng
					".":            "value", // ng
					"a":            "value", // ok
					"b": map[string]any{
						"x x x x": "value", // ng
						"y.y":     "value", // ng
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.generator
			gotErr := g.Validate()
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

// generatorTemplates follows the example in the user guide (generators.md).
func generatorTemplates() map[string]*instance.Template {
	return map[string]*instance.Template{
		"standard-vm": {
			Provisioner: "infra.vm.driver=vagrant",
			Params:      types.Params{"box": "ubuntu/jammy64", "cpus": 2, "memory": 2048},
		},
		"large-vm": {
			Params: types.Params{"cpus": 8},
		},
	}
}

// TestGenerator_ApplyTemplate_Precedence follows the user guide (templates.md):
// the definition of the generator overrides its templates, and a later
// template overrides an earlier one.
func TestGenerator_ApplyTemplate_Precedence(t *testing.T) {
	g := &instance.Generator{
		Templates: []string{"standard-vm", "large-vm"},
		Params:    types.Params{"count": 1, "memory": 4096},
	}
	if err := g.ApplyTemplate(generatorTemplates()); err != nil {
		t.Fatalf("ApplyTemplate() error = %v", err)
	}

	wantParams := types.Params{
		"count":  1,                // only on the generator
		"memory": 4096,             // the generator overrides the template ("Override memory only")
		"cpus":   8,                // large-vm overrides standard-vm
		"box":    "ubuntu/jammy64", // only in standard-vm
	}
	if diff := cmp.Diff(wantParams, g.Params); diff != "" {
		t.Errorf("params mismatch (-want +got):\n%s", diff)
	}
	if g.Provisioner != "infra.vm.driver=vagrant" {
		t.Errorf("provisioner = %q, want the one of the template", g.Provisioner)
	}

	t.Run("generator overrides the provisioner", func(t *testing.T) {
		g := &instance.Generator{Templates: []string{"standard-vm"}, Provisioner: "infra.vm.driver=libvirt"}
		if err := g.ApplyTemplate(generatorTemplates()); err != nil {
			t.Fatalf("ApplyTemplate() error = %v", err)
		}
		if g.Provisioner != "infra.vm.driver=libvirt" {
			t.Errorf("provisioner = %q, want %q", g.Provisioner, "infra.vm.driver=libvirt")
		}
	})
}

func TestGenerator_ApplyTemplate_Import(t *testing.T) {
	tpl := map[string]*instance.Template{
		"standard-vm": {
			Provisioner: "infra.vm.driver=vagrant",
			Import:      types.Imports{"cert": {From: "template_cert"}, "key": {From: "template_key"}},
		},
		"large-vm": {
			Import: types.Imports{"key": {From: "large_key"}},
		},
	}
	g := &instance.Generator{
		Templates: []string{"standard-vm", "large-vm"},
		Import:    types.Imports{"cert": {From: "generator_cert", Optional: true}},
	}
	if err := g.ApplyTemplate(tpl); err != nil {
		t.Fatalf("ApplyTemplate() error = %v", err)
	}

	want := types.Imports{
		"cert": {From: "generator_cert", Optional: true}, // the generator overrides the template
		"key":  {From: "large_key"},                      // large-vm overrides standard-vm
	}
	if diff := cmp.Diff(want, g.Import); diff != "" {
		t.Errorf("import mismatch (-want +got):\n%s", diff)
	}
}
