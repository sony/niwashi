// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cmap_test

import (
	"fmt"
	"testing"

	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/recipe"
)

type MockRecipeFinder struct {
	recipe      *recipe.Recipe
	requestName string
}

func NewMockRecipeFinder(r *recipe.Recipe) *MockRecipeFinder {
	return &MockRecipeFinder{
		recipe: r,
	}
}

func (m *MockRecipeFinder) FindRecipe(n string) (*recipe.Recipe, error) {
	m.requestName = n
	if m.recipe == nil {
		return nil, fmt.Errorf("mock returns no recipe")
	} else {
		return m.recipe, nil
	}
}

func (m *MockRecipeFinder) FindRecipeAsAdapter(n string) (*recipe.Recipe, error) {
	m.requestName = n
	if m.recipe == nil {
		return nil, fmt.Errorf("mock returns no recipe")
	} else {
		return m.recipe, nil
	}
}

func TestNameMapper_FindRecipe(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		r    *recipe.Recipe
		// Named input parameters for receiver constructor.
		m  map[string]string
		n  string
		mk *MockRecipeFinder

		wantRequestName string
		wantErr         bool
		wantErrMsg      string
	}{
		{
			name: "success mapping",
			m:    map[string]string{"a": "b"},
			n:    "a",
			mk: NewMockRecipeFinder(&recipe.Recipe{
				Metadata: &recipe.Metadata{
					Id:      "b",
					Version: "0.0.1",
				},
			}),
			wantErr:         false,
			wantRequestName: "b",
		},
		{
			name: "success without mapping",
			m:    map[string]string{"a": "b"},
			n:    "c",
			mk: NewMockRecipeFinder(&recipe.Recipe{
				Metadata: &recipe.Metadata{
					Id:      "c",
					Version: "0.0.1",
				},
			}),
			wantErr:         false,
			wantRequestName: "c",
		},
		{
			name: "success multi-step mapping",
			m: map[string]string{
				"a": "b",
				"b": "c",
			},
			n: "a",
			mk: NewMockRecipeFinder(&recipe.Recipe{
				Metadata: &recipe.Metadata{
					Id:      "c",
					Version: "0.0.1",
				},
			}),
			wantErr:         false,
			wantRequestName: "c",
		},
		{
			name: "success: nil map",
			m:    nil,
			n:    "c",
			mk: NewMockRecipeFinder(&recipe.Recipe{
				Metadata: &recipe.Metadata{
					Id:      "c",
					Version: "0.0.1",
				},
			}),
			wantErr:         false,
			wantRequestName: "c",
		},

		{
			name:       "no finder",
			m:          nil,
			n:          "a",
			mk:         nil,
			wantErr:    true,
			wantErrMsg: "no finder",
		},
		{
			name: "circular mapping",
			m: map[string]string{
				"a": "b",
				"b": "c",
				"c": "a",
			},
			n: "a",
			mk: NewMockRecipeFinder(&recipe.Recipe{
				Metadata: &recipe.Metadata{
					Id:      "c",
					Version: "0.0.1",
				},
			}),
			wantErr:    true,
			wantErrMsg: "circular name mapping detected: [a b c]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var finder cmap.RecipeFinder
			if tt.mk != nil {
				finder = tt.mk
			}
			cm := cmap.NewNameMapper(tt.m, nil, finder)

			got, gotErr := cm.FindRecipe(tt.n)
			if tt.wantErr {
				// error case
				if gotErr == nil {
					t.Fatal("FindRecipe() succeeded unexpectedly")
				}
				if got != nil {
					t.Errorf("FindRecipe() = %v, want nil", got)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("FindRecipe() error = %v, wantErrMsg %v", gotErr.Error(), tt.wantErrMsg)
				}
			} else {
				// success case
				if gotErr != nil {
					t.Errorf("FindRecipe() error = %v, wantErr %v", gotErr, tt.wantErr)
					return
				}
				if got == nil {
					t.Fatal("FindRecipe() returned nil unexpectedly")
					return
				}
				if tt.wantRequestName != tt.mk.requestName {
					t.Errorf("FindRecipe() request name = %v, want %v", tt.mk.requestName, tt.wantRequestName)
				}
			}
		})
	}
}

type recipeFinder2 struct {
	recipes  map[string]*recipe.Recipe
	capBind  map[string]string
	toolBind map[string]string
}

func newRecipeFinder2() *recipeFinder2 {
	r1 := &recipe.Recipe{
		Kind: "node",
		Metadata: &recipe.Metadata{
			Id:      "test/test-recipe",
			Version: "1.0.0",
		},
	}
	r2 := &recipe.Recipe{
		Kind: "adapter",
		Metadata: &recipe.Metadata{
			Id:      "test/test-recipe2",
			Version: "1.0.0",
		},
	}
	return &recipeFinder2{
		capBind: map[string]string{
			"recipe1": "test/test-recipe@1.0.0",
		},
		toolBind: map[string]string{
			"adapter1": "test/test-recipe2@1.0.0",
		},
		recipes: map[string]*recipe.Recipe{
			"test/test-recipe@1.0.0":  r1,
			"test/test-recipe2@1.0.0": r2,
		},
	}
}

func (f *recipeFinder2) findRecipe(n string) (*recipe.Recipe, error) {
	if r, ok := f.recipes[n]; ok {
		return r, nil
	} else {
		return nil, fmt.Errorf("recipe not found: %q", n)
	}
}

func (f *recipeFinder2) FindRecipe(n string) (*recipe.Recipe, error) {
	if r, ok := f.capBind[n]; ok {
		n = r
	}
	return f.findRecipe(n)
}

func (f *recipeFinder2) FindRecipeAsAdapter(n string) (*recipe.Recipe, error) {
	if r, ok := f.toolBind[n]; ok {
		n = r
	}
	return f.findRecipe(n)
}

func TestNameMapper_FindRecipe2(t *testing.T) {

	tests := []struct {
		name         string // description of this test case
		capNames     []string
		adapterNames []string
		references   []string // FQIDs
	}{
		{
			name:         "no reference",
			capNames:     []string{},
			adapterNames: []string{},
			references:   []string{},
		},
		{
			name:         "error case: no recipe found",
			capNames:     []string{"dummy-recipe"},
			adapterNames: []string{},
			references:   []string{},
		},
		{
			name:         "one reference with alias name",
			capNames:     []string{"recipe1"},
			adapterNames: []string{},
			references:   []string{"test/test-recipe@1.0.0"},
		},
		{
			name:         "one reference with FQID",
			capNames:     []string{"test/test-recipe@1.0.0"},
			adapterNames: []string{},
			references:   []string{"test/test-recipe@1.0.0"},
		},
		{
			name:         "full mapping",
			capNames:     []string{"test/test-recipe@1.0.0"},
			adapterNames: []string{"adapter1"},
			references:   []string{"test/test-recipe@1.0.0", "test/test-recipe2@1.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := cmap.NewNameMapper(nil, nil, newRecipeFinder2())

			// find names first
			for _, n := range tt.capNames {
				_, _ = cm.FindRecipe(n)
			}
			for _, n := range tt.adapterNames {
				_, _ = cm.FindRecipeAsAdapter(n)
			}

			// check FQID of reference recipes in NameMapper
			if len(cm.ReferenceRecipes) != len(tt.references) {
				t.Errorf("FindRecipe() number of references = %v, want %v", len(cm.ReferenceRecipes), len(tt.references))
			}
			for _, ref := range tt.references {
				if _, ok := cm.ReferenceRecipes[ref]; !ok {
					t.Errorf("FindRecipe() reference = %v, want %v", cm.ReferenceRecipes[ref], ref)
				}
			}
		})
	}
}
