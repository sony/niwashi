// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/cluster"
	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/types"
)

func TestExport_Validate(t *testing.T) {
	tests := []struct {
		name    string
		export  *state.Export
		wantErr bool
	}{
		{
			name:    "empty export",
			export:  &state.Export{},
			wantErr: true,
		},
		{
			name: "nil node and cluster",
			export: &state.Export{
				Capability: "squid.proxy", Path: "ca/cert",
			},
			wantErr: true,
		},
		{
			name: "nil capability",
			export: &state.Export{
				Node: "node", Path: "ca/cert",
			},
			wantErr: true,
		},
		{
			name: "nil path",
			export: &state.Export{
				Node: "node", Capability: "squid.proxy",
			},
			wantErr: true,
		},
		{
			name: "conflict node and cluster",
			export: &state.Export{
				Node: "node", Cluster: "cluster", Capability: "squid.proxy", Path: "ca/cert",
			},
			wantErr: true,
		},
		{
			name: "successful export",
			export: &state.Export{
				Node: "gateway", Capability: "squid.proxy", Path: "ca/cert",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.export.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExport_ValidateWith(t *testing.T) {
	tests := []struct {
		name    string
		stage   func() *state.Stage
		export  *state.Export
		wantErr bool
	}{
		{
			name:  "no Inventory(node specified)",
			stage: func() *state.Stage { return &state.Stage{} },
			export: &state.Export{
				Node: "node", Capability: "squid.proxy", Path: "ca/cert",
			},
			wantErr: true,
		},
		{
			name:  "no Inventory(cluster specified)",
			stage: func() *state.Stage { return &state.Stage{} },
			export: &state.Export{
				Cluster: "cluster", Capability: "squid.proxy", Path: "ca/cert",
			},
			wantErr: true,
		},
		{
			name:  "no Nodes",
			stage: func() *state.Stage { return &state.Stage{Inventory: &state.Inventory{}} },
			export: &state.Export{
				Node: "node", Capability: "squid.proxy", Path: "ca/cert",
			},
			wantErr: true,
		},
		{
			name: "Node exists, but no Capability",
			stage: func() *state.Stage {
				return &state.Stage{Inventory: &state.Inventory{
					Nodes: map[string]*node.Node{
						"node": {},
					},
				}}
			},
			export: &state.Export{
				Node: "node", Capability: "squid.proxy", Path: "ca/cert",
			},
			wantErr: true,
		},

		{
			name:  "no Clusters",
			stage: func() *state.Stage { return &state.Stage{Inventory: &state.Inventory{}} },
			export: &state.Export{
				Cluster: "cluster", Capability: "squid.proxy", Path: "ca/cert",
			},
			wantErr: true,
		},
		{
			name: "Cluster exists, but no Capability",
			stage: func() *state.Stage {
				return &state.Stage{Inventory: &state.Inventory{
					Clusters: map[string]*cluster.Cluster{
						"cluster": {},
					},
				}}
			},
			export: &state.Export{
				Cluster: "cluster", Capability: "squid.proxy", Path: "ca/cert",
			},
			wantErr: true,
		},

		{
			name: "successful export(node)",
			stage: func() *state.Stage {
				return &state.Stage{Inventory: &state.Inventory{
					Nodes: map[string]*node.Node{
						"gateway": {
							Capabilities: map[string]*cap.Capability{
								"squid.proxy": {},
							},
						},
					},
				}}
			},
			export: &state.Export{
				Node: "gateway", Capability: "squid.proxy", Path: "ca/cert",
			},
			wantErr: false,
		},
		{
			name: "successful export(cluster)",
			stage: func() *state.Stage {
				return &state.Stage{Inventory: &state.Inventory{
					Clusters: map[string]*cluster.Cluster{
						"gateway": {
							Capabilities: map[string]*cap.Capability{
								"squid.proxy": {},
							},
						},
					},
				}}
			},
			export: &state.Export{
				Cluster: "gateway", Capability: "squid.proxy", Path: "ca/cert",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.export.ValidateWith(tt.stage(), nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWith() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// exportRecipeFinder resolves the names registered in recipes and records
// every name passed to FindRecipe.
type exportRecipeFinder struct {
	recipes   map[string]*recipe.Recipe
	requested []string
}

func newExportRecipeFinder() *exportRecipeFinder {
	r := &recipe.Recipe{Metadata: &recipe.Metadata{Id: "org/squid-proxy", Version: "1.0.0"}}
	return &exportRecipeFinder{
		recipes: map[string]*recipe.Recipe{
			"squid.proxy":     r, // alias
			"org/squid-proxy": r, // ID without version
		},
	}
}

func (f *exportRecipeFinder) FindRecipe(n string) (*recipe.Recipe, error) {
	f.requested = append(f.requested, n)
	r, ok := f.recipes[n]
	if !ok {
		return nil, fmt.Errorf("recipe not found: %q", n)
	}
	return r, nil
}

func (f *exportRecipeFinder) FindRecipeAsAdapter(n string) (*recipe.Recipe, error) {
	return f.FindRecipe(n)
}

func TestExport_Normalize(t *testing.T) {
	tests := []struct {
		name          string
		capability    string
		want          string
		wantRequested string
		wantErr       bool
	}{
		{
			name:          "alias",
			capability:    "squid.proxy",
			want:          "org/squid-proxy",
			wantRequested: "squid.proxy",
		},
		{
			name:          "ID without version",
			capability:    "org/squid-proxy",
			want:          "org/squid-proxy",
			wantRequested: "org/squid-proxy",
		},
		{
			// the version is ignored, so it does not have to exist in the catalog
			name:          "FQID",
			capability:    "org/squid-proxy@9.9.9",
			want:          "org/squid-proxy",
			wantRequested: "org/squid-proxy",
		},
		{
			name:       "invalid FQID",
			capability: "squid.proxy@1.0.0",
			wantErr:    true,
		},
		{
			name:          "unknown capability",
			capability:    "unknown.capability",
			wantRequested: "unknown.capability",
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newExportRecipeFinder()
			e := &state.Export{Node: "gateway", Capability: tt.capability, Path: "ca/cert"}

			err := e.Normalize(f)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Normalize() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantRequested == "" {
				if len(f.requested) != 0 {
					t.Errorf("FindRecipe() called with %v, want no call", f.requested)
				}
			} else if len(f.requested) != 1 || f.requested[0] != tt.wantRequested {
				t.Errorf("FindRecipe() called with %v, want [%q]", f.requested, tt.wantRequested)
			}

			if tt.wantErr {
				return
			}
			if e.Capability != tt.want {
				t.Errorf("Capability = %q, want %q", e.Capability, tt.want)
			}
			if e.Node != "gateway" || e.Path != "ca/cert" {
				t.Errorf("Normalize() changed other fields: %+v", e)
			}
		})
	}
}

func TestExport_Normalize_Nil(t *testing.T) {
	var e *state.Export
	if err := e.Normalize(newExportRecipeFinder()); err != nil {
		t.Errorf("Normalize() error = %v, want nil", err)
	}
}

func TestExports_Normalize(t *testing.T) {
	t.Run("all exports", func(t *testing.T) {
		e := state.Exports{
			"proxy_ca_cert":  &state.Export{Node: "gateway", Capability: "squid.proxy", Path: "ca/cert"},
			"proxy_endpoint": &state.Export{Node: "gateway", Capability: "org/squid-proxy@1.0.0", Path: "endpoint"},
		}
		if err := e.Normalize(newExportRecipeFinder()); err != nil {
			t.Fatalf("Normalize() error = %v", err)
		}
		for k, v := range e {
			if v.Capability != "org/squid-proxy" {
				t.Errorf("export %q: Capability = %q, want %q", k, v.Capability, "org/squid-proxy")
			}
		}
	})

	t.Run("error names the export", func(t *testing.T) {
		e := state.Exports{
			"proxy_ca_cert": &state.Export{Node: "gateway", Capability: "unknown.capability", Path: "ca/cert"},
		}
		err := e.Normalize(newExportRecipeFinder())
		if err == nil {
			t.Fatal("Normalize() succeeded unexpectedly")
		}
		if !strings.Contains(err.Error(), `"proxy_ca_cert"`) {
			t.Errorf("Normalize() error = %v, want the export name in it", err)
		}
	})

	t.Run("nil", func(t *testing.T) {
		var e state.Exports
		if err := e.Normalize(newExportRecipeFinder()); err != nil {
			t.Errorf("Normalize() error = %v, want nil", err)
		}
	})
}

// exportStage returns a stage that has the given capabilities on the node
// "gateway" and the cluster "k8s". A nil list omits the node or the cluster.
func exportStage(nodeCaps, clusterCaps []string) *state.Stage {
	toList := func(ids []string) cap.CapabilityList {
		l := cap.CapabilityList{}
		for _, id := range ids {
			l[id] = &cap.Capability{}
		}
		return l
	}

	s := &state.Stage{Inventory: state.NewInventory()}
	if nodeCaps != nil {
		s.Inventory.Nodes["gateway"] = &node.Node{Capabilities: toList(nodeCaps)}
	}
	if clusterCaps != nil {
		s.Inventory.Clusters["k8s"] = &cluster.Cluster{Capabilities: toList(clusterCaps)}
	}
	return s
}

func TestExport_ValidateWith_MultipleStages(t *testing.T) {
	const id = "org/squid-proxy"
	nodeExport := &state.Export{Node: "gateway", Capability: id, Path: "ca/cert"}
	clusterExport := &state.Export{Cluster: "k8s", Capability: id, Path: "ca/cert"}

	tests := []struct {
		name    string
		export  *state.Export
		stages  []*state.Stage
		wantErr bool
	}{
		{
			name:    "node in both stages",
			export:  nodeExport,
			stages:  []*state.Stage{exportStage([]string{id}, nil), exportStage([]string{id}, nil)},
			wantErr: false,
		},
		{
			name:    "node only in the first stage",
			export:  nodeExport,
			stages:  []*state.Stage{exportStage([]string{id}, nil), exportStage(nil, nil)},
			wantErr: false,
		},
		{
			name:    "node only in the second stage",
			export:  nodeExport,
			stages:  []*state.Stage{exportStage(nil, nil), exportStage([]string{id}, nil)},
			wantErr: false,
		},
		{
			name:    "node in neither stage",
			export:  nodeExport,
			stages:  []*state.Stage{exportStage(nil, nil), exportStage(nil, nil)},
			wantErr: true,
		},
		{
			// the node exists in both, but only one of them has the capability
			name:    "capability only in the second stage",
			export:  nodeExport,
			stages:  []*state.Stage{exportStage([]string{"org/other"}, nil), exportStage([]string{id}, nil)},
			wantErr: false,
		},
		{
			name:    "capability in neither stage",
			export:  nodeExport,
			stages:  []*state.Stage{exportStage([]string{"org/other"}, nil), exportStage([]string{}, nil)},
			wantErr: true,
		},
		{
			name:    "cluster only in the second stage",
			export:  clusterExport,
			stages:  []*state.Stage{exportStage(nil, nil), exportStage(nil, []string{id})},
			wantErr: false,
		},
		{
			name:    "cluster in neither stage",
			export:  clusterExport,
			stages:  []*state.Stage{exportStage(nil, []string{"org/other"}), exportStage(nil, nil)},
			wantErr: true,
		},
		{
			name:    "second stage is nil, found in the first",
			export:  nodeExport,
			stages:  []*state.Stage{exportStage([]string{id}, nil), nil},
			wantErr: false,
		},
		{
			name:    "second stage is nil, not found in the first",
			export:  nodeExport,
			stages:  []*state.Stage{exportStage(nil, nil), nil},
			wantErr: true,
		},
		{
			name:    "first stage is nil, found in the second",
			export:  nodeExport,
			stages:  []*state.Stage{nil, exportStage([]string{id}, nil)},
			wantErr: false,
		},
		{
			name:    "all stages are nil",
			export:  nodeExport,
			stages:  []*state.Stage{nil, nil},
			wantErr: true,
		},
		{
			name:    "no stage",
			export:  nodeExport,
			stages:  nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.export.ValidateWith(tt.stages...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWith() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_ValidateExportReferences(t *testing.T) {
	const id = "org/squid-proxy"

	// newState returns a State whose pre-stage "bootstrap" is built by exportStage.
	// withExport adds an export of the capability on the node "gateway".
	newState := func(nodeCaps []string, withExport bool) *state.State {
		s := state.NewState()
		stage := exportStage(nodeCaps, nil)
		if withExport {
			stage.Export = state.Exports{"proxy_ca_cert": &state.Export{
				Node: "gateway", Capability: id, Path: "ca/cert",
			}}
		}
		s.Stages["bootstrap"] = stage
		return s
	}

	tests := []struct {
		name    string
		target  *state.State
		base    *state.State
		wantErr bool
	}{
		{
			name:    "first plan (empty initial state)",
			target:  newState([]string{id}, true),
			base:    state.NewState(),
			wantErr: false,
		},
		{
			name:    "node already applied and kept in the target",
			target:  newState([]string{id}, true),
			base:    newState([]string{id}, true),
			wantErr: false,
		},
		{
			// default plan keeps what is absent from the target
			name:    "node only in the initial state",
			target:  newState(nil, true),
			base:    newState([]string{id}, true),
			wantErr: false,
		},
		{
			name:    "node in neither state",
			target:  newState(nil, true),
			base:    newState(nil, false),
			wantErr: true,
		},
		{
			// the stage is looked up by name
			name:   "node only in another stage of the initial state",
			target: newState(nil, true),
			base: func() *state.State {
				s := state.NewState()
				s.Stages["other"] = exportStage([]string{id}, nil)
				return s
			}(),
			wantErr: true,
		},
		{
			name:    "nil initial state, node in the target",
			target:  newState([]string{id}, true),
			base:    nil,
			wantErr: false,
		},
		{
			name:    "nil initial state, node not in the target",
			target:  newState(nil, true),
			base:    nil,
			wantErr: true,
		},
		{
			name:    "no export",
			target:  newState(nil, false),
			base:    nil,
			wantErr: false,
		},
		{
			name:    "nil target",
			target:  nil,
			base:    newState([]string{id}, true),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.target.ValidateExportReferences(tt.base)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExportReferences() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Run("error names the stage and the export", func(t *testing.T) {
		err := newState(nil, true).ValidateExportReferences(nil)
		if err == nil {
			t.Fatal("ValidateExportReferences() succeeded unexpectedly")
		}
		for _, want := range []string{`stage "bootstrap"`, `export "proxy_ca_cert"`} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("ValidateExportReferences() error = %v, want %s in it", err, want)
			}
		}
	})
}

// exportTestStage returns a stage whose nodes have the given capabilities,
// with the given exports.
func exportTestStage(nodes map[string][]string, exports state.Exports) *state.Stage {
	s := &state.Stage{Inventory: state.NewInventory(), Export: exports}
	for name, ids := range nodes {
		l := cap.CapabilityList{}
		for _, id := range ids {
			l[id] = &cap.Capability{}
		}
		s.Inventory.Nodes[name] = &node.Node{Capabilities: l}
	}
	return s
}

// exportTestState returns a State that has the given pre-stages.
func exportTestState(stages map[string]*state.Stage) *state.State {
	s := state.NewState()
	for name, stage := range stages {
		s.Stages[name] = stage
	}
	return s
}

func TestState_ValidateExportReferences_EffectiveExports(t *testing.T) {
	const id = "org/squid-proxy"
	export := func(nodeName string) *state.Export {
		return &state.Export{Node: nodeName, Capability: id, Path: "ca/cert"}
	}
	gateway := map[string][]string{"gateway": {id}}

	tests := []struct {
		name    string
		target  *state.State
		initial *state.State
		wantErr bool
	}{
		{
			// the stage "a" is kept by the default plan, so its export remains
			name: "same export name in a stage only in the initial state",
			target: exportTestState(map[string]*state.Stage{
				"b": exportTestStage(gateway, state.Exports{"e1": export("gateway")}),
			}),
			initial: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(gateway, state.Exports{"e1": export("gateway")}),
			}),
			wantErr: true,
		},
		{
			// the target of "a" omits "e1", but the default plan keeps it
			name: "same export name kept only in the initial state of a stage",
			target: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(gateway, nil),
				"b": exportTestStage(gateway, state.Exports{"e1": export("gateway")}),
			}),
			initial: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(gateway, state.Exports{"e1": export("gateway")}),
			}),
			wantErr: true,
		},
		{
			name: "export redefined in the same stage",
			target: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(gateway, state.Exports{"e1": export("gateway")}),
			}),
			initial: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(gateway, state.Exports{"e1": export("old")}),
			}),
			wantErr: false,
		},
		{
			name:   "export only in the initial state, source kept",
			target: exportTestState(map[string]*state.Stage{"a": exportTestStage(nil, nil)}),
			initial: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(gateway, state.Exports{"e1": export("gateway")}),
			}),
			wantErr: false,
		},
		{
			// the target definition replaces the initial one as a whole; merging
			// field by field would keep the node "missing" and fail
			name: "target definition replaces the initial one as a whole",
			target: exportTestState(map[string]*state.Stage{
				"a": func() *state.Stage {
					s := exportTestStage(nil, state.Exports{"e1": {Cluster: "k8s", Capability: id, Path: "ca/cert"}})
					s.Inventory.Clusters["k8s"] = &cluster.Cluster{Capabilities: cap.CapabilityList{id: {}}}
					return s
				}(),
			}),
			initial: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(nil, state.Exports{"e1": export("missing")}),
			}),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.target.ValidateExportReferences(tt.initial)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExportReferences() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Run("duplicate error names both stages", func(t *testing.T) {
		err := tests[0].target.ValidateExportReferences(tests[0].initial)
		if err == nil {
			t.Fatal("ValidateExportReferences() succeeded unexpectedly")
		}
		for _, want := range []string{`export "e1"`, `stage "a"`, `stage "b"`} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("ValidateExportReferences() error = %v, want %s in it", err, want)
			}
		}
	})
}

func TestState_ValidatePrunedExportReferences(t *testing.T) {
	const id = "org/squid-proxy"
	export := func(nodeName string) *state.Export {
		return &state.Export{Node: nodeName, Capability: id, Path: "ca/cert"}
	}

	tests := []struct {
		name    string
		target  *state.State
		initial *state.State
		wantErr bool
	}{
		{
			name: "source kept in the target",
			target: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(map[string][]string{"A": {id}}, state.Exports{"x": export("A")}),
			}),
			initial: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(map[string][]string{"A": {id}}, state.Exports{"x": export("A")}),
			}),
			wantErr: false,
		},
		{
			// the example in 029 7-2: prune keeps the initial definition (node A),
			// but removes node A
			name: "initial definition refers to a node removed by prune",
			target: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(map[string][]string{"B": {id}}, state.Exports{"x": export("B")}),
			}),
			initial: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(map[string][]string{"A": {id}}, state.Exports{"x": export("A")}),
			}),
			wantErr: true,
		},
		{
			// the initial definition is checked, not the target one
			name: "target definition refers to a missing node, initial one is kept",
			target: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(map[string][]string{"A": {id}}, state.Exports{"x": export("C")}),
			}),
			initial: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(map[string][]string{"A": {id}}, state.Exports{"x": export("A")}),
			}),
			wantErr: false,
		},
		{
			name: "source capability removed by prune",
			target: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(map[string][]string{"A": {}}, state.Exports{"x": export("A")}),
			}),
			initial: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(map[string][]string{"A": {id}}, state.Exports{"x": export("A")}),
			}),
			wantErr: true,
		},
		{
			name: "export removed together with its source",
			target: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(nil, nil),
			}),
			initial: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(map[string][]string{"A": {id}}, state.Exports{"x": export("A")}),
			}),
			wantErr: false,
		},
		{
			name:   "stage removed by prune",
			target: exportTestState(nil),
			initial: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(map[string][]string{"A": {id}}, state.Exports{"x": export("A")}),
			}),
			wantErr: false,
		},
		{
			// prune does not add definitions, so a target-only export is not checked
			name: "export only in the target",
			target: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(nil, state.Exports{"x": export("C")}),
			}),
			initial: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(nil, nil),
			}),
			wantErr: false,
		},
		{
			name:    "nil initial state",
			target:  exportTestState(map[string]*state.Stage{"a": exportTestStage(nil, state.Exports{"x": export("C")})}),
			initial: nil,
			wantErr: false,
		},
		{
			name:   "nil target",
			target: nil,
			initial: exportTestState(map[string]*state.Stage{
				"a": exportTestStage(map[string][]string{"A": {id}}, state.Exports{"x": export("A")}),
			}),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.target.ValidatePrunedExportReferences(tt.initial)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePrunedExportReferences() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	t.Run("error names the stage and the export", func(t *testing.T) {
		err := tests[1].target.ValidatePrunedExportReferences(tests[1].initial)
		if err == nil {
			t.Fatal("ValidatePrunedExportReferences() succeeded unexpectedly")
		}
		for _, want := range []string{`stage "a"`, `export "x"`} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("ValidatePrunedExportReferences() error = %v, want %s in it", err, want)
			}
		}
	})
}

func TestExports_ValidateImports(t *testing.T) {
	// the param name and the export name differ so that a check looking up
	// the wrong one fails
	exports := state.Exports{"proxy_ca_cert": {Node: "gateway", Capability: "org/squid-proxy", Path: "ca/cert"}}
	imports := func(from string, optional bool) types.Imports {
		return types.Imports{"cert": {From: from, Optional: optional}}
	}

	tests := []struct {
		name    string
		exports state.Exports
		params  types.Params
		imports types.Imports
		wantErr string // empty means no error
	}{
		{name: "export exists", exports: exports, imports: imports("proxy_ca_cert", false)},
		{
			name:    "export not found",
			exports: exports,
			imports: imports("proxy_endpoint", false),
			wantErr: `param "cert": export "proxy_endpoint" not found`,
		},
		{
			name:    "only an export named after the param",
			exports: state.Exports{"cert": {Node: "gateway", Capability: "org/squid-proxy", Path: "ca/cert"}},
			imports: imports("proxy_ca_cert", false),
			wantErr: `export "proxy_ca_cert" not found`,
		},
		{name: "optional, export not found", exports: state.Exports{}, imports: imports("proxy_ca_cert", true)},
		{
			name:    "param also set by params",
			exports: exports,
			params:  types.Params{"cert": "/etc/ca.pem"},
			imports: imports("proxy_ca_cert", false),
			wantErr: `param "cert" is set by both params and import (from "proxy_ca_cert")`,
		},
		{
			name:    "params has a key named after the export",
			exports: exports,
			params:  types.Params{"proxy_ca_cert": "/etc/ca.pem"},
			imports: imports("proxy_ca_cert", false),
		},
		{
			// params is the fallback for an optional import (029 3-3)
			name:    "optional, param also set by params",
			exports: exports,
			params:  types.Params{"cert": "/etc/ca.pem"},
			imports: imports("proxy_ca_cert", true),
		},
		{name: "no import", exports: state.Exports{}, params: types.Params{"cert": "/etc/ca.pem"}},
		{
			name:    "nil exports",
			exports: nil,
			imports: imports("proxy_ca_cert", false),
			wantErr: `export "proxy_ca_cert" not found`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.exports.ValidateImports(tt.params, tt.imports)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateImports() error = %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("ValidateImports() succeeded unexpectedly")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ValidateImports() error = %v, want %q in it", err, tt.wantErr)
			}
		})
	}
}

// exportsFixture returns the exports of the initial state (I) and the target
// state (T). "proxy_ca_cert" is in both with different definitions,
// "registry_url" only in I and "dns_server" only in T.
func exportsFixture() (initial, target state.Exports) {
	initial = state.Exports{
		"proxy_ca_cert": {Node: "gateway", Capability: "org/squid-proxy", Path: "ca/cert"},
		"registry_url":  {Node: "gateway", Capability: "org/registry", Path: "url"},
	}
	target = state.Exports{
		"proxy_ca_cert": {Cluster: "edge", Capability: "org/squid-proxy-ha", Path: "ca/pem"},
		"dns_server":    {Node: "resolver", Capability: "org/dns", Path: "address"},
	}
	return initial, target
}

func TestExports_Merge(t *testing.T) {
	initial, target := exportsFixture()

	tests := []struct {
		name  string
		e     state.Exports
		other state.Exports
		want  state.Exports
	}{
		{name: "both nil", want: nil},
		{name: "nil receiver", other: target, want: target},
		{name: "nil other", e: initial, want: initial},
		{
			// the same name is replaced with the whole definition of other
			name:  "other wins on the same name",
			e:     initial,
			other: target,
			want: state.Exports{
				"proxy_ca_cert": target["proxy_ca_cert"],
				"registry_url":  initial["registry_url"],
				"dns_server":    target["dns_server"],
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.e.Merge(tt.other)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Merge() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("does not share entries with the inputs", func(t *testing.T) {
		initial, target := exportsFixture()
		got := initial.Merge(target)
		got["proxy_ca_cert"].Path = "changed"
		got["registry_url"].Path = "changed"
		if target["proxy_ca_cert"].Path != "ca/pem" || initial["registry_url"].Path != "url" {
			t.Errorf("Merge() result shares entries with the inputs")
		}
	})
}

func TestExports_RestrictTo(t *testing.T) {
	initial, target := exportsFixture()

	tests := []struct {
		name  string
		e     state.Exports
		other state.Exports
		want  state.Exports
	}{
		{name: "both nil", want: nil},
		{name: "nil receiver", other: target, want: nil},
		{name: "nil other", e: initial, want: nil},
		{
			// the names come from other, the definitions from the receiver
			name:  "keeps the definitions of the receiver",
			e:     initial,
			other: target,
			want:  state.Exports{"proxy_ca_cert": initial["proxy_ca_cert"]},
		},
		{
			name:  "no common name",
			e:     state.Exports{"registry_url": initial["registry_url"]},
			other: state.Exports{"dns_server": target["dns_server"]},
			want:  state.Exports{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.e.RestrictTo(tt.other)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RestrictTo() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("does not modify or share entries with the receiver", func(t *testing.T) {
		initial, target := exportsFixture()
		got := initial.RestrictTo(target)
		got["proxy_ca_cert"].Path = "changed"
		if len(initial) != 2 || initial["proxy_ca_cert"].Path != "ca/cert" {
			t.Errorf("RestrictTo() modified the receiver: %v", initial)
		}
	})
}

func TestStage_ExportMerge_ExportRestrictTo_Nil(t *testing.T) {
	initial, target := exportsFixture()
	i := &state.Stage{Export: initial}
	tg := &state.Stage{Export: target}
	var none *state.Stage

	if got := none.ExportMerge(tg); !reflect.DeepEqual(got, target) {
		t.Errorf("nil.ExportMerge() = %v, want %v", got, target)
	}
	if got := i.ExportMerge(none); !reflect.DeepEqual(got, initial) {
		t.Errorf("ExportMerge(nil) = %v, want %v", got, initial)
	}
	if got := none.ExportRestrictTo(tg); got != nil {
		t.Errorf("nil.ExportRestrictTo() = %v, want nil", got)
	}
	// the stage is absent in the target: nothing remains
	if got := i.ExportRestrictTo(none); got != nil {
		t.Errorf("ExportRestrictTo(nil) = %v, want nil", got)
	}
}
