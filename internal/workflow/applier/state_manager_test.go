// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/instance"
	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/types"
)

func Test_stateManager_ApplyPatches(t *testing.T) {
	tests := []struct {
		name      string
		patches   []patch.Patch
		want      *state.State
		wantDirty bool
		wantErr   bool
	}{
		{
			name:    "no patches",
			patches: []patch.Patch{},
			want:    state.NewState(),
		},
		{
			name:    "nil patches",
			patches: nil,
			want:    state.NewState(),
		},
		{
			name: "normal patches",
			patches: []patch.Patch{
				patch.NewAddPatch("/runtime/tool/hoge/path", "here is path"),
			},
			want: func() *state.State {
				s := state.NewState()
				err := s.Runtime.Tool.SetPath("here is path", "hoge", "path")
				if err != nil {
					t.Fatalf("failed to set tool path: %v", err)
				}
				return s
			}(),
			wantDirty: true,
		},
		{
			// The patch replaces the field of the stage view itself,
			// so it must be written back to the state.
			name: "replace whole inventory",
			patches: []patch.Patch{
				patch.NewAddPatch("/inventory", types.Dict{
					"nodes": types.Dict{"node-1": types.Dict{}},
				}),
			},
			want: func() *state.State {
				s := state.NewState()
				s.Inventory = &state.Inventory{
					Nodes: map[string]*node.Node{"node-1": {}},
				}
				return s
			}(),
			wantDirty: true,
		},
		{
			name: "remove whole runtime",
			patches: []patch.Patch{
				patch.NewRemovePatch("/runtime"),
			},
			want: func() *state.State {
				s := state.NewState()
				s.Runtime = nil
				return s
			}(),
			wantDirty: true,
		},
		{
			// The second patch creates the field from nil after the first one removes it.
			name: "remove and recreate infrastructure",
			patches: []patch.Patch{
				patch.NewRemovePatch("/infrastructure"),
				patch.NewAddPatch("/infrastructure", types.Dict{
					"generators": types.Dict{"gen-1": types.Dict{}},
				}),
			},
			want: func() *state.State {
				s := state.NewState()
				s.Infrastructure = &state.Infrastructure{
					Generators: map[string]*instance.Generator{"gen-1": {}},
				}
				return s
			}(),
			wantDirty: true,
		},
		{
			name: "invalid patch",
			patches: []patch.Patch{
				patch.NewRemovePatch("/runtime/tool/nonexistent/path"),
			},
			wantErr: true,
		},
		{
			name: "invalid patch: invalid path",
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
			sm := NewStateManager(state.NewState())

			gotErr := sm.ApplyPatches(tt.patches, nil, nil)
			if (gotErr != nil) != tt.wantErr {
				t.Fatalf("ApplyPatches() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
			if gotErr != nil {
				return
			}

			if diff := cmp.Diff(tt.want, sm.State()); diff != "" {
				t.Errorf("ApplyPatches() mismatch (-want +got):\n%s", diff)
			}
			if got := sm.IsStateDirty(); got != tt.wantDirty {
				t.Errorf("IsStateDirty() = %v, want %v", got, tt.wantDirty)
			}
		})
	}
}

func Test_stateManager_ApplyPatches_KeepsStateOnFailure(t *testing.T) {
	tests := []struct {
		name     string
		patches  []patch.Patch
		validate func(*state.StageView) error
	}{
		{
			name: "second patch fails",
			patches: []patch.Patch{
				patch.NewAddPatch("/inventory", types.Dict{
					"nodes": types.Dict{"node-1": types.Dict{}},
				}),
				patch.NewAddPatch("/status/unknown", "value"),
			},
		},
		{
			name: "validation fails",
			patches: []patch.Patch{
				patch.NewAddPatch("/inventory", types.Dict{
					"nodes": types.Dict{"node-1": types.Dict{}},
				}),
			},
			validate: func(*state.StageView) error {
				return errors.New("validation error")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStateManager(state.NewState())
			before := sm.State()

			if err := sm.ApplyPatches(tt.patches, nil, tt.validate); err == nil {
				t.Fatal("ApplyPatches() succeeded unexpectedly")
			}

			if sm.State() != before {
				t.Errorf("state must not be replaced on failure")
			}
			if diff := cmp.Diff(state.NewState(), sm.State()); diff != "" {
				t.Errorf("state must not be modified on failure (-want +got):\n%s", diff)
			}
			if sm.IsStateDirty() {
				t.Errorf("state must not be marked dirty on failure")
			}
		})
	}
}

func Test_stateManager_SaveState_RemovesEmptyStages(t *testing.T) {
	fs := file.NewDefaultFileSystem()
	f := file.NewFile(filepath.Join(t.TempDir(), "state.json"), fs)

	s := state.NewState()
	s.Stages["bootstrap"] = &state.Stage{
		Inventory: &state.Inventory{Nodes: map[string]*node.Node{"gateway": {}}},
	}
	// the pre-stage whose nodes are all destroyed
	s.Stages["pruned"] = &state.Stage{
		Inventory: &state.Inventory{Nodes: map[string]*node.Node{}},
	}
	s.SetDefaults()

	sm := NewStateManager(s, WithStateSaving(f))
	sm.dirtyState = true

	if err := sm.SaveState(); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	r, err := f.Open()
	if err != nil {
		t.Fatalf("failed to open the saved state: %v", err)
	}
	defer func() { _ = r.Close() }()
	saved, err := state.LoadAsJson(r)
	if err != nil {
		t.Fatalf("failed to load the saved state: %v", err)
	}

	if _, ok := saved.Stages["pruned"]; ok {
		t.Errorf("saved state must not have the empty stage: %v", saved.Stages)
	}
	if _, ok := saved.Stages["bootstrap"]; !ok {
		t.Errorf("saved state must have the non-empty stage: %v", saved.Stages)
	}
}
