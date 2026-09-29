// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cmap_test

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/recipe"
)

var (
	recipeSample1_v1 = &recipe.Recipe{
		FilePath: "./test1",
		Kind:     "node",
		Metadata: &recipe.Metadata{
			Id:      "test/sample.recipe1",
			Version: "1.0.0",
		},
		Spec: &recipe.Spec{
			Provides: []*recipe.Provide{
				{
					Name: "sample",
					Attributes: map[string]string{
						"key1": "value1",
						"key2": "type1",
					},
				},
			},
		},
	}

	recipeSample1_v2 = &recipe.Recipe{
		FilePath: "./test2",
		Kind:     "node",
		Metadata: &recipe.Metadata{
			Id:      "test/sample.recipe1",
			Version: "2.0.0",
		},
		Spec: &recipe.Spec{
			Provides: []*recipe.Provide{
				{
					Name: "sample",
					Attributes: map[string]string{
						"key1": "value1",
						"key2": "type2",
					},
				},
			},
		},
	}

	recipeSample2_v1_1 = &recipe.Recipe{
		FilePath: "./test3",
		Kind:     "node",
		Metadata: &recipe.Metadata{
			Id:      "test/sample.recipe2",
			Version: "1.1.0",
		},
		Spec: &recipe.Spec{
			Provides: []*recipe.Provide{
				{
					Name: "sample",
					Attributes: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
		},
	}

	recipeAdapter_v1 = &recipe.Recipe{
		FilePath: "./test4",
		Kind:     "adapter",
		Metadata: &recipe.Metadata{
			Id:      "tool/adapter-recipe",
			Version: "1.0.0",
		},
		Spec: &recipe.Spec{
			Provides: []*recipe.Provide{
				{
					Name: "my-adapter",
					Attributes: map[string]string{
						"key1": "value1",
					},
				},
			},
			Tasks: []*recipe.Task{
				{
					Name: "test",
				},
			},
		},
	}
)

type mockLoader struct {
	data        map[string]string
	failOnOpen  bool
	errorReader bool
}

type nopReadSeekCloser struct {
	io.ReadSeeker
}

func (n *nopReadSeekCloser) Close() error {
	return nil
}

type errorReadSeekCloser struct{}

func (e *errorReadSeekCloser) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read error")
}

func (e *errorReadSeekCloser) Seek(offset int64, whence int) (int64, error) {
	return 0, fmt.Errorf("seek error")
}

func (e *errorReadSeekCloser) Close() error {
	return fmt.Errorf("close error")
}

func (l *mockLoader) ListNames() []string {
	keys := make([]string, 0, len(l.data))
	for k := range l.data {
		keys = append(keys, k)
	}
	return keys
}

func (l *mockLoader) Open(name string) (io.ReadSeekCloser, error) {
	if l.failOnOpen {
		return nil, fmt.Errorf("failed to open: %s", name)
	}
	if _, ok := l.data[name]; ok {
		if l.errorReader {
			return &errorReadSeekCloser{}, nil
		}
		r := strings.NewReader(l.data[name])
		return &nopReadSeekCloser{r}, nil
	}
	return nil, fmt.Errorf("file not found: %s", name)
}

func NewDummyCapabilityMap() *cmap.CapabilityMap {
	cm := cmap.NewCapabilityMap(file.NewDefaultFileSystem())

	err := cm.Add(recipeSample1_v1)
	if err != nil {
		panic("")
	}

	err = cm.Add(recipeSample1_v2)
	if err != nil {
		panic("")
	}

	err = cm.Add(recipeSample2_v1_1)
	if err != nil {
		panic("")
	}

	err = cm.Add(recipeAdapter_v1)
	if err != nil {
		panic("")
	}
	return cm
}

func TestCapabilityMap_Add(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		r       *recipe.Recipe
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Error 1: nil recipe",
			r:       nil,
			wantErr: true,
			errMsg:  "argument error: recipe is nil",
		},
		{
			name: "Error 2: invalid id or version",
			r: &recipe.Recipe{
				FilePath: "./nws-recipe.yaml",
			},
			wantErr: true,
			errMsg:  "format error: id=\"\" version=\"\"",
		},
		{
			name: "Error 3: invalid id or version",
			r: &recipe.Recipe{
				Metadata: &recipe.Metadata{
					Id: "xxx",
				},
				FilePath: "./nws-recipe.yaml",
			},
			wantErr: true,
			errMsg:  "format error: id=\"xxx\" version=\"\"",
		},
		{
			name: "Error 4: invalid id or version",
			r: &recipe.Recipe{
				Metadata: &recipe.Metadata{
					Version: "v1.2.3",
				},
				FilePath: "./nws-recipe.yaml",
			},
			wantErr: true,
			errMsg:  "format error: id=\"\" version=\"v1.2.3\"",
		},
		{
			name: "Add",
			r: &recipe.Recipe{
				Metadata: &recipe.Metadata{
					Id:      "test/sample.recipe",
					Version: "v1.0.1",
				},
				FilePath: "./nws-recipe.yaml",
			},
			wantErr: false,
			errMsg:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := cmap.NewCapabilityMap(file.NewDefaultFileSystem())
			tt.r.SetDefaults()
			got := cm.Add(tt.r)
			if (got != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", got, tt.wantErr)
			}
			if got != nil && got.Error() != tt.errMsg {
				t.Errorf("Add() error message = %v, want %v", got.Error(), tt.errMsg)
			}
		})
	}

	test2 := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		r       *recipe.Recipe
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "Error 1: Conflict id and version",
			r: &recipe.Recipe{
				Metadata: &recipe.Metadata{
					Id:      "test/sample.recipe1",
					Version: "1.0.0",
				},
				FilePath: "./nws-recipe.yaml",
			},
			wantErr: true,
			errMsg:  "conflict: FQID=\"test/sample.recipe1@1.0.0\" old=\"./test1\" new=\"./nws-recipe.yaml\"",
		},
	}
	for _, tt := range test2 {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewDummyCapabilityMap()
			got := cm.Add(tt.r)
			if (got != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", got, tt.wantErr)
			}
			if got != nil && got.Error() != tt.errMsg {
				t.Errorf("Add() error message = %v, want %v", got.Error(), tt.errMsg)
			}
		})
	}
}

func TestCapabilityMap_GetRecipe(t *testing.T) {

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		id      string
		version string
		want    *recipe.Recipe
		wantErr bool
		errMsg  string
	}{
		{
			name:    "success 1",
			id:      "test/sample.recipe1",
			version: "1.0.0",
			want:    recipeSample1_v1,
			wantErr: false,
			errMsg:  "",
		},
		{
			name:    "success 2",
			id:      "test/sample.recipe1",
			version: "2.0.0",
			want:    recipeSample1_v2,
			wantErr: false,
			errMsg:  "",
		},
		{
			name:    "success 3",
			id:      "test/sample.recipe2",
			version: "1.1.0",
			want:    recipeSample2_v1_1,
			wantErr: false,
			errMsg:  "",
		},
		{
			name:    "success 4",
			id:      "test/sample.recipe1",
			version: "*",
			want:    recipeSample1_v2,
			wantErr: false,
			errMsg:  "",
		},
		{
			name:    "success 4",
			id:      "test/sample.recipe1",
			version: "latest",
			want:    recipeSample1_v2,
			wantErr: false,
			errMsg:  "",
		},
		{
			name:    "success 5",
			id:      "test/sample.recipe1",
			version: "<2.0.0",
			want:    recipeSample1_v1,
			wantErr: false,
			errMsg:  "",
		},

		{
			name:    "not found 1: version mismatch",
			id:      "test/sample.recipe1",
			version: "1.0.1",
			want:    nil,
			wantErr: true,
			errMsg:  "not found: no version found version=\"1.0.1\"",
		},
		{
			name:    "not found 2: unknown id",
			id:      "recipe99",
			version: "1.0.0",
			want:    nil,
			wantErr: true,
			errMsg:  "not found: no id found id=\"recipe99\"",
		},
		{
			name:    "not found 3: invalid version constraint",
			id:      "test/sample.recipe2",
			version: "3s,-3zz",
			want:    nil,
			wantErr: true,
			errMsg:  "improper constraint: 3s,-3zz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewDummyCapabilityMap()
			got, got2 := cm.GetRecipe(tt.id, tt.version)
			if (got != nil) != (tt.want != nil) {
				t.Errorf("GetRecipe() = %v, want %v", got, tt.want)
			} else {
				if got != nil {
					if got.Metadata.Id != tt.want.Metadata.Id || got.Metadata.Version != tt.want.Metadata.Version {
						t.Errorf("GetRecipe() = %v, want %v", got, tt.want)
					}
				}
			}
			if (got2 != nil) != tt.wantErr {
				t.Errorf("GetRecipe() = %v, want %v", got2, tt.wantErr)
			}
			if got2 != nil && got2.Error() != tt.errMsg {
				t.Errorf("GetRecipe() error message = %v, want %v", got2.Error(), tt.errMsg)
			}
		})
	}
}

func TestCapabilityMap_FindRecipe(t *testing.T) {
	/*
	   * FindRecipe
	   success patterns:
	   "test/sample.recipe1@1.0.0" -> "test/sample.recipe1@1.0.0"
	   "test/sample.recipe1@2.0.0" -> "test/sample.recipe1@2.0.0"
	   "sample.key1=value1.key2=type1" -> "test/sample.recipe1@1.0.0"
	   "sample.key1=value1.key2=type2" -> "test/sample.recipe1@2.0.0"
	   "sample.key2=type2" -> "test/sample.recipe1@2.0.0"
	   "adapter-recipe" -> "adapter-recipe@1.0.0"
	   "adapter-recipe@1.0.0" -> "adapter-recipe@1.0.0"
	   "my-adapter.key1=value1" -> "adapter-recipe@1.0.0"
	   "my-adapter" -> "adapter-recipe@1.0.0"

	   error patterns:
	   "my-adapter.key1=value1" -> error (multiple candidates)
	*/
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		input   string
		want    *recipe.Recipe
		wantErr bool
		errMsg  string
	}{
		{
			name:    "success 1: FQID exact match",
			input:   "test/sample.recipe1@1.0.0",
			want:    recipeSample1_v1,
			wantErr: false,
		},
		{
			name:    "success 2: FQID exact match",
			input:   "test/sample.recipe1@2.0.0",
			want:    recipeSample1_v2,
			wantErr: false,
		},
		{
			name:    "success 3: FQID without version",
			input:   "test/sample.recipe1",
			want:    recipeSample1_v2,
			wantErr: false,
		},
		{
			name:    "success 4: attribute match",
			input:   "sample.key1=value1.key2=type1",
			want:    recipeSample1_v1,
			wantErr: false,
		},
		{
			name:    "success 5: attribute match",
			input:   "sample.key1=value1.key2=type2",
			want:    recipeSample1_v2,
			wantErr: false,
		},
		{
			name:    "success 6: attribute match",
			input:   "sample.key2=type2",
			want:    recipeSample1_v2,
			wantErr: false,
		},

		{ // todo: should be fail (recipe is adapter)
			name:    "success 7: FQID without version",
			input:   "tool/adapter-recipe",
			want:    recipeAdapter_v1,
			wantErr: false,
		},
		{ // todo: should be fail (recipe is adapter)
			name:    "success 8: FQID exact match",
			input:   "tool/adapter-recipe@1.0.0",
			want:    recipeAdapter_v1,
			wantErr: false,
		},
		{
			name:    "success 9: attribute match",
			input:   "my-adapter.key1=value1",
			want:    recipeAdapter_v1,
			wantErr: true,
		},
		{
			// 2.0.0 matches the constraint ">=1.0.0"
			name:    "success 11: version constraint",
			input:   "test/sample.recipe1@>=1.0.0",
			want:    recipeSample1_v2,
			wantErr: false,
		},
		{
			// 1.0.0 matches the constraint "<=1.1.0"
			name:    "success 12: version constraint",
			input:   "test/sample.recipe1@<=1.1.0",
			want:    recipeSample1_v1,
			wantErr: false,
		},
		{
			// 2.0.0 matches the constraint "!=1.0.0"
			name:    "success 13: version constraint",
			input:   "test/sample.recipe1@!=1.0.0",
			want:    recipeSample1_v2,
			wantErr: false,
		},

		{
			name:    "error 1: unknown id",
			input:   "xxx",
			want:    nil,
			wantErr: true,
			errMsg:  "not found: no matching alias found for key=\"xxx\"",
		},
		{
			name:    "error 2: ambiguous alias",
			input:   "sample.key1=value1",
			want:    nil,
			wantErr: true,
			errMsg:  "not found: ambiguous alias: multiple candidates match the given attributes key=\"sample.key1=value1\"",
		},
		{
			name:    "error 3: version not found",
			input:   "test/sample.recipe1@3.0.0",
			want:    nil,
			wantErr: true,
			errMsg:  "not found: no version found version=\"3.0.0\"",
		},
		{
			name:    "error 3: version not found",
			input:   "test/sample.recipe1@3",
			want:    nil,
			wantErr: true,
			errMsg:  "not found: no version found version=\"3\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewDummyCapabilityMap()
			got, gotErr := cm.FindRecipe(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("FindRecipe() failed: %v", gotErr)
				}
				// if gotErr.Error() != tt.errMsg {
				// 	t.Errorf("FindRecipe() error = %v, want %v", gotErr.Error(), tt.errMsg)
				// }
				return
			}
			if tt.wantErr {
				t.Fatal("FindRecipe() succeeded unexpectedly")
			}
			if got.Metadata.Id != tt.want.Metadata.Id || got.Metadata.Version != tt.want.Metadata.Version {
				t.Errorf("FindRecipe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCapabilityMap_FindRecipeAsAdapter(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		input   string
		want    *recipe.Recipe
		wantErr bool
		errMsg  string
	}{
		{
			name:    "success 1: FQID exact match",
			input:   "test/sample.recipe1@1.0.0",
			want:    recipeSample1_v1,
			wantErr: true, // should be fail
		},
		{
			name:    "success 7: FQID without version",
			input:   "tool/adapter-recipe",
			want:    recipeAdapter_v1,
			wantErr: false,
		},
		{
			name:    "success 8: FQID exact match",
			input:   "tool/adapter-recipe@1.0.0",
			want:    recipeAdapter_v1,
			wantErr: false,
		},
		{
			name:    "success 9: attribute match",
			input:   "my-adapter.key1=value1",
			want:    recipeAdapter_v1,
			wantErr: false,
		},
		{
			name:    "success 10: alias only",
			input:   "my-adapter",
			want:    recipeAdapter_v1,
			wantErr: false,
		},

		{
			name:    "error 1: unknown id",
			input:   "xxx",
			want:    nil,
			wantErr: true,
			errMsg:  "not found: no matching alias found for key=\"xxx\"",
		},
		{
			name:    "error 3: version not found",
			input:   "tool/adapter-recipe@3.0.0",
			want:    nil,
			wantErr: true,
			errMsg:  "not found: no version found version=\"3.0.0\"",
		},
		{
			name:    "error 3: version not found",
			input:   "tool/adapter-recipe@3",
			want:    nil,
			wantErr: true,
			errMsg:  "not found: no version found version=\"3\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewDummyCapabilityMap()
			got, gotErr := cm.FindRecipeAsAdapter(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("FindRecipe() failed: %v", gotErr)
				}
				// if gotErr.Error() != tt.errMsg {
				// 	t.Errorf("FindRecipe() error = %v, want %v", gotErr.Error(), tt.errMsg)
				// }
				return
			}
			if tt.wantErr {
				t.Fatal("FindRecipe() succeeded unexpectedly")
			}
			if got.Metadata.Id != tt.want.Metadata.Id || got.Metadata.Version != tt.want.Metadata.Version {
				t.Errorf("FindRecipe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func correctHash(t *testing.T) string {
	path := "./testdata/hash-validation/nws-recipe.yaml"
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		e := f.Close()
		if e != nil {
			t.Fatalf("Failed to close file: %v", e)
		}
	}()
	r, err := recipe.LoadWithHash(f, nil, path)
	if err != nil {
		t.Fatalf("Failed to load recipe: %v", err)
	}
	return r.Hash
}

func loadCapabilityMap(t *testing.T, path string) *cmap.CapabilityMap {
	m := cmap.NewCapabilityMap(nil)
	// #nosec G304 -- test code
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() { _ = f.Close() }()
	r, err := recipe.Load(f, "xxxxxx")
	if err != nil {
		t.Fatalf("Failed to load recipe: %v", err)
	}

	if err = m.Add(r); err != nil {
		t.Fatalf("Failed to add recipe: %v", err)
	}
	return m
}

func TestCapabilityMap_ValidateHash(t *testing.T) {

	recipe.EnableNopLoader()

	masterHash := correctHash(t)

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		fqid        string
		hash        string
		cm          *cmap.CapabilityMap
		strictCheck bool
		wantErr     bool
	}{
		{
			name:        "success: complete match",
			hash:        correctHash(t),
			fqid:        "cmap-test/hash-validation@0.0.1",
			strictCheck: true,
			wantErr:     false,
		},
		{
			name:        "success: hash mismatch but strictCheck=false",
			hash:        masterHash,
			fqid:        "cmap-test/hash-validation@0.0.1",
			strictCheck: false,
			wantErr:     false, // through the error
		},
		{
			name:        "error: hash mismatch",
			hash:        "invalid hash",
			fqid:        "cmap-test/hash-validation@0.0.1",
			strictCheck: true,
			wantErr:     true,
		},
		{
			name:        "error: parse fqid failed",
			hash:        masterHash,
			fqid:        "cmap-test/hash-validation",
			strictCheck: false,
			wantErr:     true,
		},
		{
			name:        "error: no recipe found",
			hash:        masterHash,
			fqid:        "cmap-test/xxxx@0.0.1",
			strictCheck: false,
			wantErr:     true,
		},
		{
			name:        "error: no version found",
			hash:        masterHash,
			fqid:        "cmap-test/hash-validation@0.0.2",
			strictCheck: false,
			wantErr:     true,
		},
		{
			name:        "error: failed to open recipe file",
			hash:        masterHash,
			fqid:        "cmap-test/hash-validation@0.0.1",
			strictCheck: false,
			wantErr:     true,
			cm:          loadCapabilityMap(t, "./testdata/hash-validation/nws-recipe.yaml"),
		},
		{
			name: "error: failed to open asset file",
			hash: func() string {
				path := "./testdata/hash-validation/invalid-asset.yaml"
				f, err := os.Open(path)
				if err != nil {
					t.Fatalf("Failed to open file: %v", err)
				}
				defer func() { _ = f.Close() }()
				r, err := recipe.LoadWithHash(f, nil, path)
				if err != nil {
					t.Fatalf("Failed to load recipe: %v", err)
				}
				return r.Hash
			}(),
			fqid:        "cmap-test/hash-validation@0.0.1",
			strictCheck: false,
			wantErr:     true,
			cm:          loadCapabilityMap(t, "./testdata/hash-validation/invalid-asset.yaml"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := file.NewDefaultFileSystem()

			m := tt.cm

			var err error
			if m == nil {
				m, err = cmap.NewCapabilityMapFrom(
					nil,
					fs,
					"./testdata/hash-validation")
				if err != nil {
					t.Fatalf("NewCapabilityMapFrom() failed: %v", err)
				}
			}

			gotErr := m.ValidateHash(
				tt.fqid,
				tt.hash,
				tt.strictCheck,
				fs)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ValidateHash() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ValidateHash() succeeded unexpectedly")
			}
		})
	}
}

func TestNewCapabilityMapFrom(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		loader      cmap.Loader
		fs          file.FileSystem
		dirs        []string
		wantErr     bool
		holdRecipes []string
	}{
		{
			name:        "nil loader",
			fs:          file.NewDefaultFileSystem(),
			wantErr:     false,
			holdRecipes: []string{},
		},
		{
			name:        "Error: error dir",
			fs:          file.NewDefaultFileSystem(),
			dirs:        []string{"./testdata/invalid-dir"},
			wantErr:     true,
			holdRecipes: []string{},
		},
		{
			name: "loader error",
			loader: &mockLoader{
				data: map[string]string{
					"test/sample.recipe1@1.0.0": "",
				},
				// Simulate an error when opening any file
				failOnOpen: true,
			},
			fs:          file.NewDefaultFileSystem(),
			wantErr:     true,
			holdRecipes: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := cmap.NewCapabilityMapFrom(tt.loader, tt.fs, tt.dirs...)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("NewCapabilityMapFrom() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("NewCapabilityMapFrom() succeeded unexpectedly")
			}

			for _, r := range tt.holdRecipes {
				if _, err := got.GetRecipe(r, "*"); err != nil {
					t.Errorf("NewCapabilityMapFrom() missing recipe: %v", r)
				}
			}
		})
	}
}
