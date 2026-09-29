// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cmap

import (
	"fmt"
	"slices"

	"github.com/sony/niwashi/internal/recipe"
)

type NameMapper struct {
	capMapper        map[string]string
	toolMapper       map[string]string
	next             RecipeFinder
	CapBounds        map[string]string         // short name -> FQID
	ToolBounds       map[string]string         // short name -> FQID
	ReferenceRecipes map[string]*recipe.Recipe // FQID -> Recipe
}

func NewNameMapper(capBind map[string]string, toolBind map[string]string, f RecipeFinder) *NameMapper {
	if capBind == nil {
		capBind = map[string]string{}
	}
	if toolBind == nil {
		toolBind = map[string]string{}
	}
	return &NameMapper{
		capMapper:        capBind,
		toolMapper:       toolBind,
		next:             f,
		CapBounds:        make(map[string]string),
		ToolBounds:       make(map[string]string),
		ReferenceRecipes: make(map[string]*recipe.Recipe),
	}
}

func (r *NameMapper) find(n string, mapper, bounds map[string]string, findRecipe func(string) (*recipe.Recipe, error)) (*recipe.Recipe, error) {
	if r.next == nil {
		return nil, fmt.Errorf("no finder")
	}

	translated := []string{}
	name := n
	found := true
	for found {
		src := name
		name, found = mapper[src]
		if found {
			if slices.Contains(translated, src) {
				return nil, fmt.Errorf("circular name mapping detected: %v", translated)
			}
		} else {
			name = src
		}
		translated = append(translated, src)
	}

	if len(translated) == 0 {
		name = n
	}

	recipe, err := findRecipe(name)
	if err != nil {
		return nil, err
	}

	if n != recipe.Fqid() {
		bounds[n] = recipe.Fqid()
	}

	r.ReferenceRecipes[recipe.Fqid()] = recipe

	return recipe, nil
}

func (r *NameMapper) FindRecipe(n string) (*recipe.Recipe, error) {
	if r.next == nil {
		return nil, fmt.Errorf("no finder")
	}
	return r.find(n, r.capMapper, r.CapBounds, r.next.FindRecipe)
}

func (r *NameMapper) FindRecipeAsAdapter(n string) (*recipe.Recipe, error) {
	if r.next == nil {
		return nil, fmt.Errorf("no finder")
	}
	return r.find(n, r.toolMapper, r.ToolBounds, r.next.FindRecipeAsAdapter)
}
