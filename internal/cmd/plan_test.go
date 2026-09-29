// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cmd_test

import (
	"testing"

	"github.com/sony/niwashi/internal/cmd"
	"github.com/sony/niwashi/internal/recipe"
)

func TestRunPlanCmd(t *testing.T) {

	tempDir := t.TempDir()

	recipe.EnableNopLoader()

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		arg *cmd.PlanData
	}{
		{
			name: "normal",
			arg: &cmd.PlanData{
				WorkDir: "./testdata",
				Recipes: []string{"./testdata/recipes"},
				Targets: []string{"./testdata/plan-case1/target.yaml"},
				Out:     tempDir + "/output_plan.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.RunPlanCmd(tt.arg)
			if err != nil {
				t.Errorf("RunPlanCmd() error = %v", err)
			}
		})
	}
}

// func TestLoadStates(t *testing.T) {
// 	tests := []struct {
// 		name string // description of this test case
// 		// Named input parameters for target function.
// 		finder  cmap.RecipeFinder
// 		arg     *cmd.PlanData
// 		want    *state.State
// 		want2   *state.State
// 		wantErr bool
// 	}{
// 		{
// 			name: "nil case",
// 			finder: func() cmap.RecipeFinder {
// 				f, _ := cmap.Load("./testdata/recipes")
// 				return f
// 			}(),
// 			arg: &cmd.PlanData{
// 				Initial: "./testdata/plan-case1/initial.yaml",
// 				Targets: []string{"./testdata/plan-case1/target.yaml"},
// 				Destroy: false,
// 			},
// 			want: &state.State{
// 				Version: "nws.state/v1",
// 			},
// 			want2: &state.State{
// 				Version: "nws.state/v1",
// 				Inventory: &state.Inventory{
// 					Nodes: map[string]*state.Node{
// 						"node1": {
// 							Capabilities: []*state.Capability{
// 								{
// 									Id: "test/node-capability@1.0.0",
// 								},
// 							},
// 						},
// 					},
// 					Clusters: map[string]*state.Cluster{
// 						"cluster1": {
// 							Capabilities: []*state.Capability{
// 								{
// 									Id: "test/cluster-capability@1.0.0",
// 								},
// 							},
// 						},
// 					},
// 				},
// 				Infrastructure: &state.Infrastructure{
// 					Instances: map[string]*state.Instance{
// 						"vm1": {
// 							Provisioner: "test/provisioner@1.0.0",
// 						},
// 					},
// 				},
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, got2, gotErr := cmd.LoadStates(tt.finder, tt.arg)
// 			// t.Errorf("got=%v, got2=%v, gotErr=%v", got, got2, gotErr)
// 			if gotErr != nil {
// 				if !tt.wantErr {
// 					t.Errorf("LoadStates() failed: %v", gotErr)
// 				}
// 				return
// 			}
// 			if tt.wantErr {
// 				t.Fatal("LoadStates() succeeded unexpectedly")
// 			}
// 			// TODO: update the condition below to compare got with tt.want.
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("LoadStates() = %v, want %v", got, tt.want)
// 			}
// 			if !reflect.DeepEqual(got2, tt.want2) {
// 				t.Errorf("LoadStates() = %v, want %v", got2, tt.want2)
// 			}
// 		})
// 	}
// }
