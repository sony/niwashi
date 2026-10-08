// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sony/niwashi/internal/cluster"
	"github.com/sony/niwashi/internal/instance"
	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/transport"
	"github.com/sony/niwashi/internal/transport/ssh"
	"github.com/sony/niwashi/internal/types"
	"github.com/sony/niwashi/internal/workflow/applier"
)

func TestStageView_PatchApply(t *testing.T) {

	tests := []struct {
		name    string
		fields  *state.StageView
		patch   patch.Patch
		wantErr bool
		want    *state.StageView
	}{
		{
			name: "simple",
			fields: &state.StageView{
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
			want: &state.StageView{
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
			fields: &state.StageView{
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
			want: &state.StageView{
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
			fields: &state.StageView{
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

// newStageViewFixture returns a StageView with:
//   - node-1: bound to gen-1:inst-1
//   - node-2: not bound
//   - node-3: bound to a missing instance
//   - node-4: bound to a missing generator
//   - cluster-1
func newStageViewFixture() *state.StageView {
	return &state.StageView{
		Inventory: &state.Inventory{
			Nodes: map[string]*node.Node{
				"node-1": {InstanceRef: "gen-1:inst-1"},
				"node-2": {},
				"node-3": {InstanceRef: "gen-1:missing"},
				"node-4": {InstanceRef: "missing:inst-1"},
			},
			Clusters: map[string]*cluster.Cluster{
				"cluster-1": {},
			},
		},
		Infrastructure: &state.Infrastructure{
			Generators: map[string]*instance.Generator{
				"gen-1": {
					Instances: map[string]*instance.Instance{
						"inst-1": {
							NodeRef: "node-1",
							Connection: &transport.Connection{
								Data: map[string]transport.Transport{
									ssh.TypeName: &ssh.Transport{
										Address: &ssh.Address{Host: "localhost", User: "testuser"},
									},
								},
							},
						},
					},
				},
			},
		},
		Runtime: &state.Runtime{},
	}
}

func TestStageView_SetDefaults(t *testing.T) {
	t.Run("fill nil fields", func(t *testing.T) {
		sv := &state.StageView{}
		sv.SetDefaults()

		if sv.Inventory == nil || sv.Inventory.Nodes == nil || sv.Inventory.Clusters == nil {
			t.Errorf("Inventory is not initialized: %+v", sv.Inventory)
		}
		if sv.Infrastructure == nil || sv.Infrastructure.Generators == nil {
			t.Errorf("Infrastructure is not initialized: %+v", sv.Infrastructure)
		}
		if sv.Runtime == nil || sv.Runtime.Tool == nil || sv.Runtime.Service == nil {
			t.Errorf("Runtime is not initialized: %+v", sv.Runtime)
		}
	})

	t.Run("keep existing fields", func(t *testing.T) {
		inv := &state.Inventory{}
		infra := &state.Infrastructure{}
		rt := &state.Runtime{}
		sv := &state.StageView{Inventory: inv, Infrastructure: infra, Runtime: rt}
		sv.SetDefaults()

		if sv.Inventory != inv || sv.Infrastructure != infra || sv.Runtime != rt {
			t.Errorf("existing fields must not be replaced")
		}
	})
}

func TestStageView_GetNode(t *testing.T) {
	fixture := newStageViewFixture()

	tests := []struct {
		name     string
		sv       *state.StageView
		nodeName string
		wantNil  bool
	}{
		{name: "found", sv: fixture, nodeName: "node-1"},
		{name: "not found", sv: fixture, nodeName: "unknown", wantNil: true},
		{name: "nil inventory", sv: &state.StageView{}, nodeName: "node-1", wantNil: true},
		{name: "nil receiver", sv: nil, nodeName: "node-1", wantNil: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sv.GetNode(tt.nodeName)
			if (got == nil) != tt.wantNil {
				t.Fatalf("GetNode() = %v, wantNil %v", got, tt.wantNil)
			}
			if got == nil {
				return
			}
			if got.Name != tt.nodeName {
				t.Errorf("GetNode().Name = %q, want %q", got.Name, tt.nodeName)
			}
			if got.Node != tt.sv.Inventory.Nodes[tt.nodeName] {
				t.Errorf("GetNode() must refer to the node in the StageView")
			}
		})
	}
}

func TestStageView_GetCluster(t *testing.T) {
	fixture := newStageViewFixture()

	tests := []struct {
		name        string
		sv          *state.StageView
		clusterName string
		wantNil     bool
	}{
		{name: "found", sv: fixture, clusterName: "cluster-1"},
		{name: "not found", sv: fixture, clusterName: "unknown", wantNil: true},
		{name: "nil inventory", sv: &state.StageView{}, clusterName: "cluster-1", wantNil: true},
		{name: "nil receiver", sv: nil, clusterName: "cluster-1", wantNil: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sv.GetCluster(tt.clusterName)
			if (got == nil) != tt.wantNil {
				t.Fatalf("GetCluster() = %v, wantNil %v", got, tt.wantNil)
			}
			if got == nil {
				return
			}
			if got.Name != tt.clusterName {
				t.Errorf("GetCluster().Name = %q, want %q", got.Name, tt.clusterName)
			}
			if got.Cluster != tt.sv.Inventory.Clusters[tt.clusterName] {
				t.Errorf("GetCluster() must refer to the cluster in the StageView")
			}
		})
	}
}

func TestStageView_GetGenerator(t *testing.T) {
	fixture := newStageViewFixture()

	tests := []struct {
		name          string
		sv            *state.StageView
		generatorName string
		wantNil       bool
	}{
		{name: "found", sv: fixture, generatorName: "gen-1"},
		{name: "not found", sv: fixture, generatorName: "unknown", wantNil: true},
		{name: "nil infrastructure", sv: &state.StageView{}, generatorName: "gen-1", wantNil: true},
		{name: "nil receiver", sv: nil, generatorName: "gen-1", wantNil: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sv.GetGenerator(tt.generatorName)
			if (got == nil) != tt.wantNil {
				t.Fatalf("GetGenerator() = %v, wantNil %v", got, tt.wantNil)
			}
			if got == nil {
				return
			}
			if got.Name != tt.generatorName {
				t.Errorf("GetGenerator().Name = %q, want %q", got.Name, tt.generatorName)
			}
			if got.Generator != tt.sv.Infrastructure.Generators[tt.generatorName] {
				t.Errorf("GetGenerator() must refer to the generator in the StageView")
			}
		})
	}
}

func TestStageView_GetBoundInstance(t *testing.T) {
	fixture := newStageViewFixture()
	noInfra := newStageViewFixture()
	noInfra.Infrastructure = nil

	tests := []struct {
		name     string
		sv       *state.StageView
		nodeName string
		wantNil  bool
	}{
		{name: "bound", sv: fixture, nodeName: "node-1"},
		{name: "not bound", sv: fixture, nodeName: "node-2", wantNil: true},
		{name: "missing instance", sv: fixture, nodeName: "node-3", wantNil: true},
		{name: "missing generator", sv: fixture, nodeName: "node-4", wantNil: true},
		{name: "missing node", sv: fixture, nodeName: "unknown", wantNil: true},
		{name: "nil infrastructure", sv: noInfra, nodeName: "node-1", wantNil: true},
		{name: "nil receiver", sv: nil, nodeName: "node-1", wantNil: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sv.GetBoundInstance(tt.nodeName)
			if (got == nil) != tt.wantNil {
				t.Fatalf("GetBoundInstance() = %v, wantNil %v", got, tt.wantNil)
			}
			if got == nil {
				return
			}
			if got.GeneratorName != "gen-1" || got.InstanceName != "inst-1" {
				t.Errorf("GetBoundInstance() = %s, want gen-1:inst-1", got.GetFqid())
			}
			if got.Instance != tt.sv.Infrastructure.Generators["gen-1"].Instances["inst-1"] {
				t.Errorf("GetBoundInstance() must refer to the instance in the StageView")
			}
		})
	}
}

func TestStageView_GetConnection(t *testing.T) {
	t.Run("bound node returns a cloned connection with defaults", func(t *testing.T) {
		sv := newStageViewFixture()
		orig := sv.Infrastructure.Generators["gen-1"].Instances["inst-1"].Connection

		got := sv.GetConnection("node-1")
		if got == nil {
			t.Fatalf("GetConnection() = nil, want non-nil")
		}
		if got == orig {
			t.Errorf("GetConnection() must return a clone")
		}

		gotSsh, ok := got.Data[ssh.TypeName].(*ssh.Transport)
		if !ok {
			t.Fatalf("GetConnection() does not have ssh transport: %+v", got.Data)
		}
		if gotSsh.Address.Host != "localhost" || gotSsh.Address.Port != 22 {
			t.Errorf("GetConnection() address = %+v, want host=localhost port=22", gotSsh.Address)
		}

		// defaults must not be applied to the original connection
		origSsh := orig.Data[ssh.TypeName].(*ssh.Transport)
		if origSsh.Address.Port != 0 || origSsh.Auth != nil {
			t.Errorf("original connection must not be modified: %+v", origSsh)
		}
	})

	fixture := newStageViewFixture()
	noConn := newStageViewFixture()
	noConn.Infrastructure.Generators["gen-1"].Instances["inst-1"].Connection = nil

	tests := []struct {
		name     string
		sv       *state.StageView
		nodeName string
	}{
		{name: "not bound", sv: fixture, nodeName: "node-2"},
		{name: "missing instance", sv: fixture, nodeName: "node-3"},
		{name: "missing generator", sv: fixture, nodeName: "node-4"},
		{name: "missing node", sv: fixture, nodeName: "unknown"},
		{name: "no connection", sv: noConn, nodeName: "node-1"},
		{name: "nil inventory", sv: &state.StageView{}, nodeName: "node-1"},
		{name: "nil receiver", sv: nil, nodeName: "node-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sv.GetConnection(tt.nodeName); got != nil {
				t.Errorf("GetConnection() = %+v, want nil", got)
			}
		})
	}
}

func TestStageView_UnsupportedPath(t *testing.T) {
	for _, root := range []string{"status", "template", "metadata", "version", "unknown"} {
		t.Run(root, func(t *testing.T) {
			sv := newStageViewFixture()
			want := newStageViewFixture()

			if err := sv.Add([]string{root, "foo"}, "bar"); err == nil {
				t.Errorf("Add() error = nil, want error")
			}
			if err := sv.Remove([]string{root, "foo"}); err == nil {
				t.Errorf("Remove() error = nil, want error")
			}
			if diff := cmp.Diff(want, sv); diff != "" {
				t.Errorf("StageView must not be modified (-want +got):\n%s", diff)
			}
		})
	}

	t.Run("empty path", func(t *testing.T) {
		sv := newStageViewFixture()
		if err := sv.Add([]string{}, "bar"); err == nil {
			t.Errorf("Add() error = nil, want error")
		}
		if err := sv.Remove([]string{}); err == nil {
			t.Errorf("Remove() error = nil, want error")
		}
	})
}

// assertSameFields checks that the listed fields of after still refer to those of before.
func assertSameFields(t *testing.T, before, after *state.StageView, fields []string) {
	t.Helper()
	for _, f := range fields {
		var same bool
		switch f {
		case "inventory":
			same = before.Inventory == after.Inventory
		case "infrastructure":
			same = before.Infrastructure == after.Infrastructure
		case "runtime":
			same = before.Runtime == after.Runtime
		default:
			t.Fatalf("unknown field: %s", f)
		}
		if !same {
			t.Errorf("%s must not be replaced", f)
		}
	}
}

func TestStageView_Add(t *testing.T) {
	all := []string{"inventory", "infrastructure", "runtime"}
	emptyView := func() *state.StageView { return &state.StageView{} }

	tests := []struct {
		name     string
		sv       func() *state.StageView
		path     []string
		value    any
		wantErr  bool
		want     func() *state.StageView
		wantSame []string // fields that must keep the same pointer
	}{
		{
			name:  "add node",
			sv:    newStageViewFixture,
			path:  []string{"inventory", "nodes", "node-new"},
			value: types.Dict{"instanceRef": "gen-1:inst-2"},
			want: func() *state.StageView {
				sv := newStageViewFixture()
				sv.Inventory.Nodes["node-new"] = &node.Node{InstanceRef: "gen-1:inst-2"}
				return sv
			},
			wantSame: all,
		},
		{
			name:  "add generator",
			sv:    newStageViewFixture,
			path:  []string{"infrastructure", "generators", "gen-new"},
			value: types.Dict{"provisioner": "org/provisioner"},
			want: func() *state.StageView {
				sv := newStageViewFixture()
				sv.Infrastructure.Generators["gen-new"] = &instance.Generator{Provisioner: "org/provisioner"}
				return sv
			},
			wantSame: all,
		},
		{
			name:  "replace whole inventory",
			sv:    newStageViewFixture,
			path:  []string{"inventory"},
			value: types.Dict{"nodes": types.Dict{"node-x": types.Dict{}}},
			want: func() *state.StageView {
				sv := newStageViewFixture()
				sv.Inventory = &state.Inventory{Nodes: map[string]*node.Node{"node-x": {}}}
				return sv
			},
			wantSame: []string{"infrastructure", "runtime"},
		},
		{
			name:  "create inventory from nil",
			sv:    emptyView,
			path:  []string{"inventory"},
			value: types.Dict{"nodes": types.Dict{"node-x": types.Dict{}}},
			want: func() *state.StageView {
				return &state.StageView{
					Inventory: &state.Inventory{Nodes: map[string]*node.Node{"node-x": {}}},
				}
			},
		},
		{
			name:  "create infrastructure from nil",
			sv:    emptyView,
			path:  []string{"infrastructure"},
			value: types.Dict{"generators": types.Dict{"gen-x": types.Dict{}}},
			want: func() *state.StageView {
				return &state.StageView{
					Infrastructure: &state.Infrastructure{Generators: map[string]*instance.Generator{"gen-x": {}}},
				}
			},
		},
		{
			name:  "create runtime from nil",
			sv:    emptyView,
			path:  []string{"runtime"},
			value: types.Dict{"tool": types.Dict{"custom": "xxxxx"}},
			want: func() *state.StageView {
				return &state.StageView{
					Runtime: &state.Runtime{Tool: types.Dict{"custom": "xxxxx"}},
				}
			},
		},
		{
			name:    "sub-path of nil inventory",
			sv:      emptyView,
			path:    []string{"inventory", "nodes", "node-x"},
			value:   types.Dict{},
			wantErr: true,
		},
		{
			name:    "sub-path of nil infrastructure",
			sv:      emptyView,
			path:    []string{"infrastructure", "generators", "gen-x"},
			value:   types.Dict{},
			wantErr: true,
		},
		{
			name:    "sub-path of nil runtime",
			sv:      emptyView,
			path:    []string{"runtime", "tool", "custom"},
			value:   "xxxxx",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sv := tt.sv()
			before := *sv

			err := sv.Add(tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tt.want(), sv, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Add() mismatch (-want +got):\n%s", diff)
			}
			assertSameFields(t, &before, sv, tt.wantSame)
		})
	}
}

func TestStageView_Remove(t *testing.T) {
	all := []string{"inventory", "infrastructure", "runtime"}
	emptyView := func() *state.StageView { return &state.StageView{} }

	tests := []struct {
		name     string
		sv       func() *state.StageView
		path     []string
		wantErr  bool
		want     func() *state.StageView
		wantSame []string // fields that must keep the same pointer
	}{
		{
			name: "remove node",
			sv:   newStageViewFixture,
			path: []string{"inventory", "nodes", "node-2"},
			want: func() *state.StageView {
				sv := newStageViewFixture()
				delete(sv.Inventory.Nodes, "node-2")
				return sv
			},
			wantSame: all,
		},
		{
			name: "remove generator",
			sv:   newStageViewFixture,
			path: []string{"infrastructure", "generators", "gen-1"},
			want: func() *state.StageView {
				sv := newStageViewFixture()
				delete(sv.Infrastructure.Generators, "gen-1")
				return sv
			},
			wantSame: all,
		},
		{
			name: "remove whole inventory",
			sv:   newStageViewFixture,
			path: []string{"inventory"},
			want: func() *state.StageView {
				sv := newStageViewFixture()
				sv.Inventory = nil
				return sv
			},
			wantSame: []string{"infrastructure", "runtime"},
		},
		{
			name: "remove whole infrastructure",
			sv:   newStageViewFixture,
			path: []string{"infrastructure"},
			want: func() *state.StageView {
				sv := newStageViewFixture()
				sv.Infrastructure = nil
				return sv
			},
			wantSame: []string{"inventory", "runtime"},
		},
		{
			name: "remove whole runtime",
			sv:   newStageViewFixture,
			path: []string{"runtime"},
			want: func() *state.StageView {
				sv := newStageViewFixture()
				sv.Runtime = nil
				return sv
			},
			wantSame: []string{"inventory", "infrastructure"},
		},
		{
			// Runtime.Remove is not implemented for sub-paths
			name:    "sub-path of runtime",
			sv:      newStageViewFixture,
			path:    []string{"runtime", "tool", "custom"},
			wantErr: true,
		},
		{
			name:    "sub-path of nil inventory",
			sv:      emptyView,
			path:    []string{"inventory", "nodes", "node-1"},
			wantErr: true,
		},
		{
			name:    "sub-path of nil infrastructure",
			sv:      emptyView,
			path:    []string{"infrastructure", "generators", "gen-1"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sv := tt.sv()
			before := *sv

			err := sv.Remove(tt.path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tt.want(), sv, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Remove() mismatch (-want +got):\n%s", diff)
			}
			assertSameFields(t, &before, sv, tt.wantSame)
		})
	}
}

func TestStageView_Update(t *testing.T) {
	sv := newStageViewFixture()
	if err := sv.Update(newStageViewFixture()); err == nil {
		t.Errorf("Update() error = nil, want error")
	}
}
