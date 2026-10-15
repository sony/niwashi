// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package plan_test

import (
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/state"
)

func TestNewPlan(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		want *plan.Plan
	}{
		{
			name: "default plan",
			want: &plan.Plan{
				Version: plan.Version,
				Mode:    plan.ModeDefault,
				States: plan.States{
					Target: func() *state.State {
						s := state.NewState()
						return s
					}(),
				},
				Runtime: func() plan.Runtime {
					r := plan.Runtime{}
					r.SetDefaults()
					return r
				}(),
				Metadata: plan.NewMetadata(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := plan.NewPlan()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPlan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newWorkflows(stages ...string) map[string]*plan.Workflow {
	m := map[string]*plan.Workflow{}
	for _, s := range stages {
		m[s] = &plan.Workflow{}
	}
	return m
}

func TestPlan_StageOrder(t *testing.T) {
	tests := []struct {
		name      string
		mode      plan.Mode
		construct map[string]*plan.Workflow
		destruct  map[string]*plan.Workflow
		want      []string
		wantErr   bool
	}{
		{
			name:      "construct without pre-stage",
			mode:      plan.ModeDefault,
			construct: newWorkflows(state.MainStageName, state.InitStageName),
			want:      []string{state.InitStageName, state.MainStageName},
		},
		{
			name:      "construct with a pre-stage",
			mode:      plan.ModeDefault,
			construct: newWorkflows(state.MainStageName, "bootstrap", state.InitStageName),
			want:      []string{state.InitStageName, "bootstrap", state.MainStageName},
		},
		{
			name:     "destroy runs the stages in reverse order",
			mode:     plan.ModeDestroy,
			destruct: newWorkflows(state.MainStageName, "bootstrap", state.InitStageName),
			want:     []string{state.MainStageName, "bootstrap", state.InitStageName},
		},
		{
			name:     "prune runs the stages in reverse order",
			mode:     plan.ModePrune,
			destruct: newWorkflows(state.MainStageName, "bootstrap", state.InitStageName),
			want:     []string{state.MainStageName, "bootstrap", state.InitStageName},
		},
		{
			name:      "construct uses only the construct graph",
			mode:      plan.ModeDefault,
			construct: newWorkflows(state.MainStageName, state.InitStageName),
			destruct:  newWorkflows("bootstrap"),
			want:      []string{state.InitStageName, state.MainStageName},
		},
		{
			name:      "destroy uses only the destruct graph",
			mode:      plan.ModeDestroy,
			construct: newWorkflows("bootstrap"),
			destruct:  newWorkflows(state.MainStageName, state.InitStageName),
			want:      []string{state.MainStageName, state.InitStageName},
		},
		{
			name: "construct with no graph",
			mode: plan.ModeDefault,
			want: []string{state.InitStageName, state.MainStageName},
		},
		{
			name: "destroy with no graph",
			mode: plan.ModeDestroy,
			want: []string{state.MainStageName, state.InitStageName},
		},
		{
			name: "stages with nil workflow",
			mode: plan.ModeDefault,
			construct: map[string]*plan.Workflow{
				state.InitStageName: nil,
				"bootstrap":         nil,
				state.MainStageName: nil,
			},
			want: []string{state.InitStageName, "bootstrap", state.MainStageName},
		},
		{
			name:      "more than one pre-stage",
			mode:      plan.ModeDefault,
			construct: newWorkflows(state.MainStageName, "bootstrap", "proxy", state.InitStageName),
			wantErr:   true,
		},
		{
			name:    "unsupported mode",
			mode:    plan.Mode("unknown"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &plan.Plan{
				Mode: tt.mode,
				ExecutionPlan: plan.ExecutionPlan{
					ConstructGraph: tt.construct,
					DestructGraph:  tt.destruct,
				},
			}

			got, err := p.StageOrder()
			if (err != nil) != tt.wantErr {
				t.Fatalf("StageOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StageOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}
