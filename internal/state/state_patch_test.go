// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/types"
	"github.com/sony/niwashi/internal/workflow/applier"
)

func TestState_PatchApply(t *testing.T) {

	tests := []struct {
		name    string
		fields  *state.State
		patch   patch.Patch
		wantErr bool
		want    *state.State
	}{
		{
			name: "simple",
			fields: &state.State{
				Runtime: &state.Runtime{
					Tool: types.Dict{},
				},
			},
			patch: &applier.JsonPatch{
				Op:    recipe.OpSet,
				Path:  "/runtime/tool/custom/path",
				Value: "xxxxx",
			},
			wantErr: false,
			want: &state.State{
				Runtime: &state.Runtime{
					Tool: types.Dict{
						"custom": types.Dict{
							"path": "xxxxx",
						},
					},
				},
			},
		},
		{
			name: "simple2",
			fields: &state.State{
				Runtime: &state.Runtime{
					Tool: types.Dict{},
				},
			},
			patch: &applier.JsonPatch{
				Op:   recipe.OpSet,
				Path: "/runtime/tool/custom",
				Value: types.Dict{
					"path": "xxxxx",
					"args": "yyyyy",
				},
			},
			wantErr: false,
			want: &state.State{
				Runtime: &state.Runtime{
					Tool: types.Dict{
						"custom": types.Dict{
							"path": "xxxxx",
							"args": "yyyyy",
						},
					},
				},
			},
		},
		{
			name: "error",
			fields: &state.State{
				Runtime: &state.Runtime{
					Tool: types.Dict{},
				},
			},
			patch: &applier.JsonPatch{
				Op:   recipe.OpSet,
				Path: "/runtime/params",
				Value: types.Dict{
					"path": "xxxxx",
					"args": "yyyyy",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotErr := tt.patch.Apply(tt.fields)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("Patch.Apply() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}

			if gotErr == nil {
				if diff := cmp.Diff(tt.want, tt.fields); diff != "" {
					t.Errorf("Patch.Apply() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
