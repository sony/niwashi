// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import (
	"fmt"
	"testing"

	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/recipe"
)

type mockFinder struct {
	recipes map[string]*recipe.Recipe
}

func (m *mockFinder) FindRecipeAsAdapter(n string) (*recipe.Recipe, error) {
	return m.FindRecipe(n)
}

func (m *mockFinder) FindRecipe(n string) (*recipe.Recipe, error) {
	if r, ok := m.recipes[n]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("recipe %q not found", n)
}

func newMockAdapterRecipe() *recipe.Recipe {
	return &recipe.Recipe{
		Kind: "adapter",
		Metadata: &recipe.Metadata{
			Id:      "test_tool",
			Version: "1.0.0",
		},
		Spec: &recipe.Spec{
			AllowedScope: "node",
		},
	}
}

func Test_getAdapterInfo(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		finder  cmap.RecipeFinder
		task    TaskSpec
		want    *AdapterInfo
		wantErr bool
	}{
		{
			name:   "not tool run spec",
			finder: &mockFinder{},
			task: &taskSpec{
				actionSpec: &nilActionSpec{},
			},
			want:    nil, // no error but nil returns
			wantErr: false,
		},
		{
			name:   "recipe not found",
			finder: &mockFinder{},
			task: &taskSpec{
				RecipeSpec: &recipeSpec{
					metaSpec: &metaSpec{
						kind: "node",
					},
				},
				actionSpec: &recipe.ToolRunSpec{
					ToolRef: "test_tool",
				},
			},
			wantErr: true, // not found error
		},
		{
			name: "not allowed scope",
			finder: &mockFinder{
				recipes: map[string]*recipe.Recipe{
					"test_tool": newMockAdapterRecipe(),
				},
			},
			task: &taskSpec{
				RecipeSpec: &recipeSpec{
					metaSpec: &metaSpec{
						kind: "cluster",
					},
				},
				actionSpec: &recipe.ToolRunSpec{
					ToolRef: "test_tool",
					Command: "xyzCommand",
				},
			},
			wantErr: true, // not allowed
		},
		{
			name: "allowed scope(same scope)",
			finder: &mockFinder{
				recipes: map[string]*recipe.Recipe{
					"test_tool": newMockAdapterRecipe(),
				},
			},
			task: &taskSpec{
				RecipeSpec: &recipeSpec{
					metaSpec: &metaSpec{
						kind: "node",
					},
				},
				actionSpec: &recipe.ToolRunSpec{
					ToolRef: "test_tool",
					Command: "xyzCommand",
				},
			},
			want: &AdapterInfo{
				CommandName: "xyzCommand",
			},
			wantErr: false, // allowed
		},
		{
			name: "allowed scope(wildcard)",
			finder: &mockFinder{
				recipes: map[string]*recipe.Recipe{
					"test_tool": func() *recipe.Recipe {
						r := newMockAdapterRecipe()
						r.Spec.AllowedScope = "*"
						return r
					}(),
				},
			},
			task: &taskSpec{
				RecipeSpec: &recipeSpec{
					metaSpec: &metaSpec{
						kind: "node",
					},
				},
				actionSpec: &recipe.ToolRunSpec{
					ToolRef: "test_tool",
					Command: "xyzCommand",
				},
			},
			want: &AdapterInfo{
				CommandName: "xyzCommand",
			},
			wantErr: false, // allowed
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := getAdapterInfo(tt.finder, tt.task)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("getAdapterInfo() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("getAdapterInfo() succeeded unexpectedly")
			}

			if tt.want == nil {
				if got != nil {
					t.Errorf("getAdapterInfo() = %v, want nil", got)
				}
				return
			} else {
				// compare
				if tt.want.CommandName != got.CommandName {
					t.Errorf("getAdapterInfo() CommandName = %v, want %v", got.CommandName, tt.want.CommandName)
				}
			}
		})
	}
}
