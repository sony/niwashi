// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package runtime_test

import (
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/cluster"
	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/platform"
	"github.com/sony/niwashi/internal/runtime"
	"github.com/sony/niwashi/internal/state"
)

func TestNewStateFrom(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s        *state.State
		nodes    []string
		clusters []string
		want     *runtime.State
		wantErr  bool
	}{
		{
			name: "no selected nodes or clusters",
			s: &state.State{
				Inventory: &state.Inventory{
					Nodes: map[string]*node.Node{
						"node1": {},
						"node2": {},
					},
					Clusters: map[string]*cluster.Cluster{
						"cluster1": {},
						"cluster2": {},
					},
				},
			},
			nodes:    []string{},
			clusters: []string{},
			want: &runtime.State{
				Host: &runtime.Node{
					Os:   platform.HostOs,
					Arch: platform.HostArch,
				},
				Inventory: &runtime.Inventory{
					Nodes:    map[string]*runtime.Node{},
					Clusters: map[string]*runtime.Cluster{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := runtime.NewStateFrom(state.NewAccessor(tt.s), tt.nodes, tt.clusters)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("NewStateFrom() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("NewStateFrom() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStateFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}
