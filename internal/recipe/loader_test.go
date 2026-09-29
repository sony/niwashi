// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sony/niwashi/internal/action/exec_local"
	"github.com/sony/niwashi/internal/recipe"
)

func TestLoadWithHash(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		reader      io.Reader
		assetOpener func(path string) (io.ReadCloser, error)
		want        *recipe.Recipe
		wantErr     bool
	}{
		{
			name: "default",
			reader: func() io.Reader {
				f, _ := os.Open("testdata/test-recipe.yaml")
				return f
			}(),
			assetOpener: func(path string) (io.ReadCloser, error) {
				// #nosec G304 -- test code
				return os.Open(filepath.Join("testdata", path))
			},
			want: &recipe.Recipe{
				Version: "nws.recipe/v1",
				Kind:    "node",
				Metadata: &recipe.Metadata{
					Id:      "test/test-recipe",
					Version: "1.0.0",
				},
				Spec: &recipe.Spec{
					Workspace: &recipe.Workspace{
						Mode: "ephemeral",
					},
					Assets: []string{
						"assets/asset1.txt",
						"assets/asset2.txt",
					},
					Tasks: []*recipe.Task{
						{
							Name:      "task1",
							Operation: "construct",
							Action: recipe.Action{
								ActionType: "exec.local",
								Spec:       &exec_local.Spec{},
							},
						},
					},
				},
				Hash: "6a55878f1c324f5402ac95fabb8279da9f8e026031efdaa19f4781375c2dc04e",
				// cat ./testdata/test-recipe.yaml ./testdata/assets/asset1.txt ./testdata/assets/asset2.txt | sha256sum
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := recipe.LoadWithHash(tt.reader, tt.assetOpener, "")
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("LoadWithHash() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("LoadWithHash() succeeded unexpectedly")
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("LoadWithHash() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
