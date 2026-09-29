// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cmap

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/recipe"
)

var validRecipe = `
version: nws.recipe/v1
kind: node
metadata:
  id: test/test-recipe
  version: 1.0.0
spec:
  assets:
    - assets/asset1.txt
  tasks:
    - name: task1
      action:
        nop: {}
`

func Test_loadFromRecipe(t *testing.T) {
	recipe.EnableNopLoader()

	tests := []struct {
		name    string
		cm      *CapabilityMap
		path    string
		wantErr bool
	}{
		{
			name:    "valid recipe",
			cm:      NewCapabilityMap(file.NewDefaultFileSystem()),
			path:    "testdata/valid-recipe.yaml",
			wantErr: false,
		},
		{
			name:    "error recipe",
			cm:      NewCapabilityMap(file.NewDefaultFileSystem()),
			path:    "testdata/error-recipe.yaml",
			wantErr: true,
		},
		{
			name:    "wrong path",
			cm:      NewCapabilityMap(file.NewDefaultFileSystem()),
			path:    "testdata/nonexistent-recipe.yaml",
			wantErr: true,
		},
		{
			name: "conflicting recipe",
			cm: func() *CapabilityMap {
				cm := NewCapabilityMap(file.NewDefaultFileSystem())
				err := cm.Add(&recipe.Recipe{Metadata: &recipe.Metadata{Id: "test/test-recipe", Version: "1.0.0"}})
				if err != nil {
					t.Fatalf("Failed to add recipe: %v", err)
				}
				return cm
			}(),
			path:    "testdata/valid-recipe.yaml",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cm.loadRecipeFromFile(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("loadRecipeFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_loadFromCatalog(t *testing.T) {
	recipe.EnableNopLoader()

	tests := []struct {
		name    string
		cm      *CapabilityMap
		path    string
		wantErr bool
	}{
		{
			name:    "valid catalog",
			cm:      NewCapabilityMap(file.NewDefaultFileSystem()),
			path:    "testdata/valid-catalog.yaml",
			wantErr: false,
		},
		{
			name:    "no catalog",
			cm:      NewCapabilityMap(file.NewDefaultFileSystem()),
			path:    "testdata/nonexistent-catalog.yaml",
			wantErr: true,
		},
		{
			name:    "error catalog",
			cm:      NewCapabilityMap(file.NewDefaultFileSystem()),
			path:    "testdata/error-catalog.yaml",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cm.loadFromCatalog(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("loadFromCatalog() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoad(t *testing.T) {

	recipe.EnableNopLoader()

	t.Run("test testdir/for-walk", func(t *testing.T) {
		cm := NewCapabilityMap(file.NewDefaultFileSystem())
		err := cm.LoadFromDirs("testdata/for-walk/recipes1", "testdata/for-walk/recipes2")

		recipes := []string{
			"cmap-test/recipe1@0.0.1",
			"cmap-test/recipe2@0.0.1",
			"cmap-test/sub1@0.0.1",
			"cmap-test/sub2@0.0.2",
			"cmap-test/xyz@1.0.0",
			"cmap-test/xyz@2.0.0",
		}

		if err != nil {
			t.Errorf("Load() error = %v, wantErr %v", err, nil)
		}
		if cm == nil {
			t.Errorf("Load() = %v, want !nil", cm)
		}

		for _, id := range recipes {
			r, err := cm.FindRecipe(id)
			if err != nil {
				t.Errorf("cmap.FindRecipe(%q) error = %v, wantErr %v", id, err, nil)
			}
			if r == nil {
				t.Errorf("cmap.FindRecipe(%q) = nil, want !nil", id)
			}
		}
	})
}

func TestLoadFromDirs_UnrecognizedYAML(t *testing.T) {

	recipe.EnableNopLoader()

	cm := NewCapabilityMap(file.NewDefaultFileSystem())
	err := cm.LoadFromDirs("testdata/for-walk-suspects")
	if err != nil {
		t.Fatalf("LoadFromDirs() error = %v", err)
	}

	// zzz-recipe.yaml sorts after nws-catalog.yaml, so it is only ever
	// discovered via the catalog, not by its own filename — regression
	// check for the ordering bug where the catalog's removeSuspect ran
	// before the file had even been added as a suspect.
	if _, err := cm.FindRecipe("cmap-test/zzz@0.0.1"); err != nil {
		t.Errorf("expected catalog-referenced recipe to be loaded: %v", err)
	}

	suspects := cm.getUnloadedSuspects()

	for _, s := range suspects {
		switch {
		case strings.HasSuffix(s, "zzz-recipe.yaml"):
			t.Errorf("catalog-referenced recipe file must not be reported as unrecognized: %q", s)
		case strings.HasSuffix(s, "values.yaml"):
			t.Errorf("file declared under spec.assets must not be reported as unrecognized: %q", s)
		case strings.Contains(s, string(filepath.Separator)+".hidden"+string(filepath.Separator)):
			t.Errorf("files under a dot-prefixed directory must never be scanned at all: %q", s)
		case strings.HasSuffix(s, ".dotfile.yaml"):
			t.Errorf("a dot-prefixed file must never be scanned at all: %q", s)
		}
	}

	if want := 1; len(suspects) != want {
		t.Errorf("getUnloadedSuspects() = %v, want exactly %d entry (orphan.yaml)", suspects, want)
	} else if !strings.HasSuffix(suspects[0], "orphan.yaml") {
		t.Errorf("getUnloadedSuspects() = %v, want it to contain orphan.yaml", suspects)
	}
}

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

func TestCapabilityMap_LoadFromLoader(t *testing.T) {
	recipe.EnableNopLoader()

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		loader  Loader
		wantErr bool
		cm      *CapabilityMap
	}{
		{
			name:    "Empty list",
			loader:  &mockLoader{data: map[string]string{}},
			wantErr: false,
		},
		{
			name: "Success",
			loader: &mockLoader{
				data: map[string]string{
					"validRecipe": validRecipe,
				},
			},
			wantErr: false,
		},
		{
			name: "Error: failed to open",
			loader: &mockLoader{
				data: map[string]string{
					"validRecipe": validRecipe,
				},
				failOnOpen: true,
			},
			wantErr: true,
		},
		{
			name: "Error: failed to read",
			loader: &mockLoader{
				data: map[string]string{
					"validRecipe": validRecipe,
				},
				errorReader: true,
			},
			wantErr: true,
		},
		{
			name: "Error: failed to add",
			loader: &mockLoader{
				data: map[string]string{
					"validRecipe": validRecipe,
				},
				failOnOpen: true,
			},
			wantErr: true,
			cm: func() *CapabilityMap {
				// add a recipe with same FQID to cause conflict error
				cm := NewCapabilityMap(nil)
				err := cm.Add(&recipe.Recipe{
					Metadata: &recipe.Metadata{
						Id:      "test/test-recipe",
						Version: "1.0.0",
					},
				})
				if err != nil {
					t.Fatalf("Failed to add recipe: %v", err)
				}
				return cm
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := tt.cm
			if cm == nil {
				cm = NewCapabilityMap(nil)
			}
			gotErr := cm.LoadFromLoader(tt.loader)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("LoadFromLoader() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("LoadFromLoader() succeeded unexpectedly")
			}
		})
	}
}
