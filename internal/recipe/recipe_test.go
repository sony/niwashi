// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_test

import (
	"testing"

	"github.com/sony/niwashi/internal/recipe"
)

func NewMockRecipe() *recipe.Recipe {
	r := &recipe.Recipe{}
	r.SetDefaults()

	r.Kind = "host"
	r.Metadata.Id = "test/mock"
	r.Metadata.Version = "0.1.0"
	r.Spec.Tasks = append(r.Spec.Tasks, &recipe.Task{
		Name: "task1",
		Action: recipe.Action{
			ActionType: "mock",
			Spec:       &mockActionSpec{},
		},
	})
	return r
}

func TestRecipe_Validate(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		r       *recipe.Recipe
		wantErr bool
	}{
		{
			name:    "valid recipe",
			r:       NewMockRecipe(),
			wantErr: false,
		},
		{
			name: "invalid version",
			r: func() *recipe.Recipe {
				r := NewMockRecipe()
				r.Version = "invalid version"
				return r
			}(),
			wantErr: true,
		},
		{
			name: "invalid kind",
			r: func() *recipe.Recipe {
				r := NewMockRecipe()
				r.Kind = "invalid kind"
				return r
			}(),
			wantErr: true,
		},
		{
			name: "invalid metadata id",
			r: func() *recipe.Recipe {
				r := NewMockRecipe()
				r.Metadata.Id = "invalid id"
				return r
			}(),
			wantErr: true,
		},
		{
			name: "empty metadata id",
			r: func() *recipe.Recipe {
				r := NewMockRecipe()
				r.Metadata.Id = ""
				return r
			}(),
			wantErr: true,
		},
		{
			name: "invalid metadata version",
			r: func() *recipe.Recipe {
				r := NewMockRecipe()
				r.Metadata.Version = "invalid version"
				return r
			}(),
			wantErr: true,
		},
		{
			name: "empty metadata version",
			r: func() *recipe.Recipe {
				r := NewMockRecipe()
				r.Metadata.Version = ""
				return r
			}(),
			wantErr: true,
		},
		{
			name: "nil spec",
			r: func() *recipe.Recipe {
				r := NewMockRecipe()
				r.Spec = nil
				return r
			}(),
			wantErr: true,
		},
		{
			name: "invalid adapter",
			r: func() *recipe.Recipe {
				r := NewMockRecipe()
				r.Kind = recipe.KindAdapter
				a := newMockAdapter()
				a.AllowedScope = "invalid scope"
				r.Spec = (*recipe.Spec)(a)
				return r
			}(),
			wantErr: true,
		},
		{
			name: "commands is not allowed for non-adapter recipe",
			r: func() *recipe.Recipe {
				r := NewMockRecipe()
				r.Spec.Commands = map[string]*recipe.Command{
					"cmd1": &recipe.Command{
						Task: "task1",
					},
				}
				return r
			}(),
			wantErr: true,
		},
		{
			name: "executionUnit is not allowed for non-adapter recipe",
			r: func() *recipe.Recipe {
				r := NewMockRecipe()
				r.Spec.ExecutionUnit = "default"
				return r
			}(),
			wantErr: true,
		},
		{
			name: "allowedScope is not allowed for non-adapter recipe",
			r: func() *recipe.Recipe {
				r := NewMockRecipe()
				r.Spec.AllowedScope = "host"
				return r
			}(),
			wantErr: true,
		},
		{
			name: "invalid defaults params",
			r: func() *recipe.Recipe {
				r := NewMockRecipe()
				r.Spec.Defaults = &recipe.Defaults{
					Params: map[string]any{
						"invalid-name": "value", // ng
						"invalid.name": "value", // ng
						"invalid name": "value", // ng
						"0":            "value", // ng
						".":            "value", // ng
						"a":            "value", // ok
					},
				}
				return r
			}(),
			wantErr: true,
		},
		{
			name: "invalid defaults env",
			r: func() *recipe.Recipe {
				r := NewMockRecipe()
				r.Spec.Defaults = &recipe.Defaults{
					Environments: map[string]string{
						"":             "value", // ng
						"invalid-name": "value", // ng
						"invalid.name": "value", // ng
						"invalid name": "value", // ng
						"0":            "value", // ng
						".":            "value", // ng
						"a":            "value", // ok
					},
				}
				return r
			}(),
			wantErr: true,
		},
		{
			name: "invalid runtime name",
			r: func() *recipe.Recipe {
				r := NewMockRecipe()
				r.Spec.Runtime = &recipe.Runtime{
					Type: "tool",
					Name: "invalid name", // ng
				}
				return r
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.r.Validate()
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
