// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state_test

import (
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/cluster"
	"github.com/sony/niwashi/internal/instance"
	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/types"
)

func TestState_Validate(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		state   *state.State
		wantErr bool
	}{
		{
			name:    "empty",
			state:   state.NewState(),
			wantErr: false,
		},
		{
			name: "invalid node capability params",
			state: &state.State{
				Stage: state.Stage{
					Inventory: &state.Inventory{
						Nodes: map[string]*node.Node{
							"node1": {
								Capabilities: cap.CapabilityList{
									"cap1": &cap.Capability{
										Params: map[string]any{
											"invalid.name": "value", // ng
											"invalid name": "value", // ng
											"0":            "value", // ng
											".":            "value", // ng
											"a":            "value", // ok
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid node capability params 2",
			state: &state.State{
				Stage: state.Stage{
					Inventory: &state.Inventory{
						Nodes: map[string]*node.Node{
							"node1": {
								Capabilities: cap.CapabilityList{
									"cap1": &cap.Capability{
										Params: map[string]any{
											"a": map[string]any{
												"invalid.name": "value", // ng
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid cluster capability params",
			state: &state.State{
				Stage: state.Stage{
					Inventory: &state.Inventory{
						Clusters: map[string]*cluster.Cluster{
							"cluster1": {
								Capabilities: cap.CapabilityList{
									"cap1": &cap.Capability{
										Params: map[string]any{
											"invalid.name": "value", // ng
											"invalid name": "value", // ng
											"0":            "value", // ng
											".":            "value", // ng
											"a":            "value", // ok
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid cluster params",
			state: &state.State{
				Stage: state.Stage{
					Inventory: &state.Inventory{
						Clusters: map[string]*cluster.Cluster{
							"cluster1": {
								Params: map[string]any{
									"invalid.name": "value", // ng
									"invalid name": "value", // ng
									"0":            "value", // ng
									".":            "value", // ng
									"a":            "value", // ok
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid generator params",
			state: &state.State{
				Stage: state.Stage{
					Infrastructure: &state.Infrastructure{
						Generators: map[string]*instance.Generator{
							"generator1": {
								Provisioner: "dummy provisioner",
								Params: map[string]any{
									"invalid.name": "value", // ng
									"invalid name": "value", // ng
									"0":            "value", // ng
									".":            "value", // ng
									"a":            "value", // ok
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid template node cap params",
			state: &state.State{
				Template: &state.Template{
					Node: map[string]*node.Template{
						"node1": {
							Capabilities: cap.CapabilityList{
								"cap1": &cap.Capability{
									Params: map[string]any{
										"invalid.name": "value", // ng
										"invalid name": "value", // ng
										"0":            "value", // ng
										".":            "value", // ng
										"a":            "value", // ok
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid template cluster cap params",
			state: &state.State{
				Template: &state.Template{
					Cluster: map[string]*cluster.Template{
						"cluster1": {
							Capabilities: cap.CapabilityList{
								"cap1": &cap.Capability{
									Params: map[string]any{
										"invalid.name": "value", // ng
										"invalid name": "value", // ng
										"0":            "value", // ng
										".":            "value", // ng
										"a":            "value", // ok
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid template cluster params",
			state: &state.State{
				Template: &state.Template{
					Cluster: map[string]*cluster.Template{
						"cluster1": {
							Params: map[string]any{
								"invalid.name": "value", // ng
								"invalid name": "value", // ng
								"0":            "value", // ng
								".":            "value", // ng
								"a":            "value", // ok
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid template generator params",
			state: &state.State{
				Template: &state.Template{
					Generator: map[string]*instance.Template{
						"generator1": {
							Params: map[string]any{
								"invalid.name": "value", // ng
								"invalid name": "value", // ng
								"0":            "value", // ng
								".":            "value", // ng
								"a":            "value", // ok
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid runtime tool",
			state: &state.State{
				Runtime: &state.Runtime{
					Tool: map[string]any{
						"invalid.name": "value", // ng
						"invalid name": "value", // ng
						"0":            "value", // ng
						".":            "value", // ng
						"a":            "value", // ok
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid runtime service",
			state: &state.State{
				Runtime: &state.Runtime{
					Service: map[string]any{
						"invalid.name": "value", // ng
						"invalid name": "value", // ng
						"0":            "value", // ng
						".":            "value", // ng
						"a":            "value", // ok
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.state
			gotErr := s.Validate()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Validate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Validate() succeeded unexpectedly")
			}
		})
	}
}

func TestState_Validate_Export(t *testing.T) {
	tests := []struct {
		name    string
		state   func() *state.State
		wantErr bool
	}{
		{
			name: "export in the main stage",
			state: func() *state.State {
				s := state.NewState()
				s.Export = types.Dict{"proxy_ca_cert": types.Dict{}}
				return s
			},
			wantErr: true,
		},
		{
			name: "empty export in the main stage",
			state: func() *state.State {
				s := state.NewState()
				s.Export = types.Dict{}
				return s
			},
			wantErr: true,
		},
		{
			name: "export in a pre-stage",
			state: func() *state.State {
				s := state.NewState()
				s.Stages["bootstrap"] = &state.Stage{
					Export: types.Dict{"proxy_ca_cert": types.Dict{}},
				}
				s.SetDefaults()
				return s
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.state().Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_Validate_StageNameInError(t *testing.T) {
	s := state.NewState()
	s.Stages["bootstrap"] = &state.Stage{
		Inventory: &state.Inventory{
			Nodes: map[string]*node.Node{
				"node1": {
					Capabilities: cap.CapabilityList{
						"cap1": &cap.Capability{
							Params: map[string]any{"invalid.name": "value"},
						},
					},
				},
			},
		},
	}
	s.SetDefaults()

	err := s.Validate()
	if err == nil {
		t.Fatal("Validate() succeeded unexpectedly")
	}
	if !strings.Contains(err.Error(), `stage "bootstrap"`) {
		t.Errorf("Validate() error = %q, want it to contain the stage name", err)
	}
}
