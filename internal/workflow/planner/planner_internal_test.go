// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package planner

import (
	"testing"

	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/state"
)

func newStateWithNodes(names ...string) *state.State {
	s := state.NewState()
	for _, name := range names {
		s.Inventory.Nodes[name] = &node.Node{}
	}
	return s
}

func TestPlanner_computeDiff(t *testing.T) {
	tests := []struct {
		name          string
		mode          plan.Mode
		initial       *state.State
		target        *state.State
		wantConstruct []string
		wantDestruct  []string
		wantErr       bool
	}{
		{
			name:          "default mode computes construct and update diffs",
			mode:          plan.ModeDefault,
			initial:       newStateWithNodes("cp"),
			target:        newStateWithNodes("cp", "worker"),
			wantConstruct: []string{"worker"},
		},
		{
			name:         "destroy mode computes destruct diff",
			mode:         plan.ModeDestroy,
			initial:      newStateWithNodes("cp"),
			target:       state.NewState(),
			wantDestruct: []string{"cp"},
		},
		{
			name:         "prune mode computes destruct diff",
			mode:         plan.ModePrune,
			initial:      newStateWithNodes("cp", "worker"),
			target:       newStateWithNodes("cp"),
			wantDestruct: []string{"worker"},
		},
		{
			name:    "invalid mode",
			mode:    plan.Mode("unknown"),
			initial: state.NewState(),
			target:  state.NewState(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Planner{mode: tt.mode, initial: tt.initial, target: tt.target}

			got, err := p.computeDiff([]string{state.MainStageName})
			if (err != nil) != tt.wantErr {
				t.Fatalf("computeDiff() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			d, ok := got[state.MainStageName]
			if !ok || d == nil {
				t.Fatalf("computeDiff() = %v, want the diff of the main stage", got)
			}
			if _, ok := got[state.InitStageName]; ok {
				t.Errorf("computeDiff() must not have the diff of the %s stage", state.InitStageName)
			}

			switch tt.mode {
			case plan.ModeDefault:
				if d.Construct == nil || d.Update == nil || d.Destruct != nil {
					t.Fatalf("diff = %+v, want construct and update only", d)
				}
				assertNodes(t, "Construct.Node", d.Construct.Node, tt.wantConstruct)
			default:
				if d.Destruct == nil || d.Construct != nil || d.Update != nil {
					t.Fatalf("diff = %+v, want destruct only", d)
				}
				assertNodes(t, "Destruct.Node", d.Destruct.Node, tt.wantDestruct)
			}
		})
	}
}

func assertNodes(t *testing.T, name string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s = %v, want %v", name, got, want)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s = %v, want %v", name, got, want)
			return
		}
	}
}
