// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/state"
)

func Test_applyContext_ApplyPatches(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		ctx     *applyContext
		patches []patch.Patch
		want    *state.State
		wantErr bool
	}{
		{
			name:    "no patches",
			ctx:     &applyContext{},
			patches: []patch.Patch{},
			wantErr: false,
		},
		{
			name:    "nil patches",
			ctx:     &applyContext{},
			patches: nil,
			wantErr: false,
		},
		{
			name: "normal patches",
			ctx: &applyContext{
				stateManager: NewStateManager(state.NewAccessor(state.NewState())),
			},
			patches: []patch.Patch{
				patch.NewAddPatch("/runtime/tool/hoge/path", "here is path"),
			},
			wantErr: false,
			want: func() *state.State {
				s := state.NewState()
				err := s.Runtime.Tool.SetPath("here is path", "hoge", "path")
				if err != nil {
					t.Fatalf("failed to set tool path: %v", err)
				}
				return s
			}(),
		},
		{
			name: "invalid patch",
			ctx: &applyContext{
				stateManager: NewStateManager(state.NewAccessor(state.NewState())),
			},
			patches: []patch.Patch{
				patch.NewRemovePatch("/runtime/tool/nonexistent/path"),
			},
			wantErr: true,
		},
		{
			name: "invalid patch: invalid path",
			ctx: &applyContext{
				stateManager: NewStateManager(state.NewAccessor(state.NewState())),
			},
			patches: []patch.Patch{
				&JsonPatch{
					Op:    "remove",
					Path:  "/runtime/unknownfield/path",
					Value: "xxxx",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			gotErr := tt.ctx.ApplyPatches(tt.patches, nil, nil)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ApplyPatches() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ApplyPatches() succeeded unexpectedly")
			} else {
				var s *state.State
				if tt.ctx.stateManager != nil {
					s = tt.ctx.stateManager.State().State
				}
				if diff := cmp.Diff(tt.want, s); diff != "" {
					t.Errorf("Merge() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
