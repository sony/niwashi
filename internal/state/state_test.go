// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state_test

import (
	"testing"

	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/state"
)

// newStateWithStage returns a default State that has a pre-stage named "bootstrap".
func newStateWithStage(nodes map[string]*node.Node) *state.State {
	s := state.NewState()
	s.Stages["bootstrap"] = &state.Stage{
		Inventory: &state.Inventory{Nodes: nodes},
	}
	s.SetDefaults()
	return s
}

func TestState_Stage(t *testing.T) {
	t.Run("default stage refers to the fields of State", func(t *testing.T) {
		s := state.NewState()
		got, err := s.StageView(state.DefaultStageName)
		if err != nil {
			t.Fatalf("Stage() error = %v, want nil", err)
		}
		if got == nil {
			t.Fatalf("Stage() = nil, want non-nil")
		}
		if got.Inventory != s.Inventory || got.Infrastructure != s.Infrastructure || got.Runtime != s.Runtime {
			t.Errorf("Stage() must refer to the fields of State")
		}
	})

	t.Run("pre-stage refers to the stage and the shared runtime", func(t *testing.T) {
		s := newStateWithStage(map[string]*node.Node{"gateway": {}})
		got, err := s.StageView("bootstrap")
		if err != nil {
			t.Fatalf("StageView() error = %v, want nil", err)
		}
		bs := s.Stages["bootstrap"]
		if got.Inventory != bs.Inventory || got.Infrastructure != bs.Infrastructure {
			t.Errorf("StageView() must refer to the fields of the pre-stage")
		}
		if got.Runtime != s.Runtime {
			t.Errorf("StageView() must refer to the runtime shared by all stages")
		}
		if got.GetNode("gateway") == nil {
			t.Errorf("StageView() must see the nodes of the pre-stage")
		}
	})

	t.Run("unsupported stage name", func(t *testing.T) {
		s := state.NewState()
		if got, err := s.StageView("bootstrap"); got != nil || err == nil {
			t.Errorf("Stage() = %+v, err = %v, want nil, want error", got, err)
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		var s *state.State
		if got, err := s.StageView(state.DefaultStageName); got != nil || err == nil {
			t.Errorf("Stage() = %+v, err = %v, want nil, want error", got, err)
		}
	})
}

func TestState_SetDefaults_FillsNilStage(t *testing.T) {
	s := state.NewState()
	s.Stages["bootstrap"] = nil
	s.SetDefaults()

	bs := s.Stages["bootstrap"]
	if bs == nil || bs.Inventory == nil || bs.Infrastructure == nil {
		t.Fatalf("Stages[bootstrap] = %+v, want a stage with defaults", bs)
	}
	if bs.Export != nil {
		t.Errorf("Export = %v, want nil (Export must not be initialized by default)", bs.Export)
	}
	if _, err := s.StageView("bootstrap"); err != nil {
		t.Errorf("StageView() error = %v, want nil", err)
	}
}

func TestState_Clone_Stages(t *testing.T) {
	s := newStateWithStage(map[string]*node.Node{"gateway": {}})
	c := s.Clone()

	if c.Stages["bootstrap"] == s.Stages["bootstrap"] ||
		c.Stages["bootstrap"].Inventory == s.Stages["bootstrap"].Inventory {
		t.Errorf("Clone() must deep copy the stages")
	}
	if c.Stages["bootstrap"].Inventory.Nodes["gateway"] == nil {
		t.Errorf("Clone() must keep the nodes of the stages")
	}
}

func TestState_Merge_Stages(t *testing.T) {
	a := newStateWithStage(map[string]*node.Node{"gateway": {}})
	a.Inventory.Nodes["cp"] = &node.Node{}

	b := newStateWithStage(map[string]*node.Node{"pxe": {}})
	b.Stages["proxy"] = &state.Stage{
		Inventory: &state.Inventory{Nodes: map[string]*node.Node{"squid": {}}},
	}
	b.SetDefaults()

	got, err := a.Merge(b)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// the same stage name is merged
	bs := got.Stages["bootstrap"]
	if bs == nil || bs.Inventory.Nodes["gateway"] == nil || bs.Inventory.Nodes["pxe"] == nil {
		t.Errorf("Stages[bootstrap] = %+v, want gateway and pxe", bs)
	}
	// a stage only in the other is added
	if ps := got.Stages["proxy"]; ps == nil || ps.Inventory.Nodes["squid"] == nil {
		t.Errorf("Stages[proxy] = %+v, want squid", ps)
	}
	// the main stage is merged separately from the pre-stages
	if got.Inventory.Nodes["cp"] == nil || got.Inventory.Nodes["gateway"] != nil {
		t.Errorf("main stage nodes = %v, want only cp", got.Inventory.Nodes)
	}
	// the inputs are not modified
	if a.Stages["bootstrap"].Inventory.Nodes["pxe"] != nil {
		t.Errorf("Merge() must not modify the receiver")
	}
}
