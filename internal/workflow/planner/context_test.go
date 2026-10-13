// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package planner_test

import (
	"testing"

	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workflow/planner"
)

func newStateWithPreStage() *state.State {
	s := state.NewState()
	s.Inventory.Nodes["cp"] = &node.Node{}
	s.Stages["bootstrap"] = &state.Stage{
		Inventory: &state.Inventory{Nodes: map[string]*node.Node{"gateway": {}}},
	}
	s.SetDefaults()
	return s
}

func newDiffMap() map[string]*state.Diff {
	return map[string]*state.Diff{
		state.MainStageName: {
			Construct: &state.Difference{Node: []string{"cp"}},
			Update:    &state.Difference{Node: []string{"cp-update"}},
			Destruct:  &state.Difference{Node: []string{"cp-destruct"}},
		},
		"bootstrap": {
			Construct: &state.Difference{Node: []string{"gateway"}},
		},
	}
}

func Test_context_WithStage_StageView(t *testing.T) {
	target := newStateWithPreStage()
	base := newStateWithPreStage()
	root := planner.NewContext(recipe.OperationConstruct, newFakeRecipeFinder())

	t.Run("main stage", func(t *testing.T) {
		ctx, err := root.WithStage(state.MainStageName, target, base, newDiffMap())
		if err != nil {
			t.Fatalf("WithStage() error = %v", err)
		}
		if ctx.StageView().Inventory != target.Inventory {
			t.Errorf("StageView() must refer to the main stage of the target state")
		}
		if ctx.BaseStageView().Inventory != base.Inventory {
			t.Errorf("BaseStageView() must refer to the main stage of the base state")
		}
	})

	t.Run("pre-stage", func(t *testing.T) {
		ctx, err := root.WithStage("bootstrap", target, base, newDiffMap())
		if err != nil {
			t.Fatalf("WithStage() error = %v", err)
		}
		if ctx.StageView().Inventory != target.Stages["bootstrap"].Inventory {
			t.Errorf("StageView() must refer to the pre-stage of the target state")
		}
		if ctx.BaseStageView().Inventory != base.Stages["bootstrap"].Inventory {
			t.Errorf("BaseStageView() must refer to the pre-stage of the base state")
		}
		if ctx.StageView().GetNode("gateway") == nil || ctx.StageView().GetNode("cp") != nil {
			t.Errorf("StageView() must see only the nodes of the pre-stage")
		}
	})

	t.Run("nil base state", func(t *testing.T) {
		ctx, err := root.WithStage(state.MainStageName, target, nil, newDiffMap())
		if err != nil {
			t.Fatalf("WithStage() error = %v", err)
		}
		if ctx.BaseStageView() != nil {
			t.Errorf("BaseStageView() = %+v, want nil", ctx.BaseStageView())
		}
	})

	t.Run("unknown stage", func(t *testing.T) {
		if _, err := root.WithStage("unknown", target, base, newDiffMap()); err == nil {
			t.Errorf("WithStage() error = nil, want error")
		}
	})
}

func Test_context_WithStage_Diff(t *testing.T) {
	target := newStateWithPreStage()

	tests := []struct {
		name           string
		op             string
		stage          string
		diffMap        map[string]*state.Diff
		wantDiff       []string
		wantUpdateDiff []string
	}{
		{
			name:           "construct uses construct and update diffs",
			op:             recipe.OperationConstruct,
			stage:          state.MainStageName,
			diffMap:        newDiffMap(),
			wantDiff:       []string{"cp"},
			wantUpdateDiff: []string{"cp-update"},
		},
		{
			name:     "destruct uses only destruct diff",
			op:       recipe.OperationDestruct,
			stage:    state.MainStageName,
			diffMap:  newDiffMap(),
			wantDiff: []string{"cp-destruct"},
		},
		{
			name:     "diff of the selected stage",
			op:       recipe.OperationConstruct,
			stage:    "bootstrap",
			diffMap:  newDiffMap(),
			wantDiff: []string{"gateway"},
		},
		{
			// _init has no diff
			name:    "stage without diff",
			op:      recipe.OperationConstruct,
			stage:   state.InitStageName,
			diffMap: newDiffMap(),
		},
		{
			name:    "nil diff map",
			op:      recipe.OperationConstruct,
			stage:   state.MainStageName,
			diffMap: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := planner.NewContext(tt.op, newFakeRecipeFinder())
			ctx, err := root.WithStage(tt.stage, target, nil, tt.diffMap)
			if err != nil {
				t.Fatalf("WithStage() error = %v", err)
			}

			assertDiffNodes(t, "Diff()", ctx.Diff(), tt.wantDiff)
			assertDiffNodes(t, "UpdateDiff()", ctx.UpdateDiff(), tt.wantUpdateDiff)
		})
	}
}

func assertDiffNodes(t *testing.T, name string, d *state.Difference, want []string) {
	t.Helper()
	if want == nil {
		if d != nil {
			t.Errorf("%s = %+v, want nil", name, d)
		}
		return
	}
	if d == nil || len(d.Node) != len(want) || d.Node[0] != want[0] {
		t.Errorf("%s = %+v, want nodes %v", name, d, want)
	}
}

func Test_context_WithStage_Graph(t *testing.T) {
	target := newStateWithPreStage()
	root := planner.NewContext(recipe.OperationConstruct, newFakeRecipeFinder())

	mainCtx, err := root.WithStage(state.MainStageName, target, nil, newDiffMap())
	if err != nil {
		t.Fatalf("WithStage() error = %v", err)
	}
	preCtx, err := root.WithStage("bootstrap", target, nil, newDiffMap())
	if err != nil {
		t.Fatalf("WithStage() error = %v", err)
	}

	if mainCtx.Graph() == nil || preCtx.Graph() == nil {
		t.Fatalf("WithStage() must create a graph for the stage")
	}
	if mainCtx.Graph() == preCtx.Graph() {
		t.Errorf("WithStage() must create a separate graph for each stage")
	}
	if root.Graph() != nil {
		t.Errorf("WithStage() must not set a graph to the root context")
	}
	if mainCtx.Operation() != recipe.OperationConstruct {
		t.Errorf("Operation() = %q, want %q", mainCtx.Operation(), recipe.OperationConstruct)
	}
}
