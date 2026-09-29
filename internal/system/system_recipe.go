// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package system

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"slices"
)

var systemRecipes = map[string][]byte{}

func RegisterSystemRecipe(name string, yaml []byte) {
	systemRecipes[name] = yaml
}

func OpenSystemRecipe(name string) (io.ReadSeekCloser, error) {
	yaml, found := systemRecipes[name]
	if !found {
		return nil, fmt.Errorf("system recipe not found: %s", name)
	}

	return &nopCloser{bytes.NewReader(yaml)}, nil
}

func GetSystemRecipeNames() []string {
	return slices.Collect(maps.Keys(systemRecipes))
}

type nopCloser struct {
	*bytes.Reader
}

func (*nopCloser) Close() error {
	return nil
}

type SystemRecipeLoader struct{}

func NewSystemRecipeLoader() *SystemRecipeLoader {
	return &SystemRecipeLoader{}
}

func (s *SystemRecipeLoader) Open(name string) (io.ReadSeekCloser, error) {
	return OpenSystemRecipe(name)
}

func (s *SystemRecipeLoader) ListNames() []string {
	return GetSystemRecipeNames()
}
