// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/cluster"
	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/types"
)

func Test_capKeys(t *testing.T) {
	type args struct {
		m cap.CapabilityList
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := capKeys(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("capKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_differenceCapChange(t *testing.T) {
	newNode := func(caps cap.CapabilityList) *node.Node {
		return &node.Node{Capabilities: caps}
	}
	getNodeCaps := func(n *node.Node) cap.CapabilityList { return n.Capabilities }

	tests := []struct {
		name    string
		a       map[string]*node.Node
		b       map[string]*node.Node
		want    map[string][]string
		wantErr string // non-empty: error expected, matched as a substring
	}{
		{
			name: "no nodes",
			a:    nil,
			b:    nil,
			want: nil,
		},
		{
			name: "node missing from target is ignored",
			a: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{
					"cap1": {Version: "1.0.0"},
				}),
			},
			b:    nil,
			want: nil,
		},
		{
			name: "capability missing from target node is ignored",
			a: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{
					"cap1": {Version: "1.0.0"},
				}),
			},
			b: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{}),
			},
			want: nil,
		},
		{
			name: "same version and same params is not a change",
			a: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{
					"cap1": {Version: "1.0.0", Params: types.Params{"memory": "1024"}},
				}),
			},
			b: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{
					"cap1": {Version: "1.0.0", Params: types.Params{"memory": "1024"}},
				}),
			},
			want: nil,
		},
		{
			name: "same version and different params is a change",
			a: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{
					"cap1": {Version: "1.0.0", Params: types.Params{"memory": "1024"}},
				}),
			},
			b: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{
					"cap1": {Version: "1.0.0", Params: types.Params{"memory": "2048"}},
				}),
			},
			want: map[string][]string{"node1": {"cap1"}},
		},
		{
			name: "upgraded version is a change regardless of params",
			a: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{
					"cap1": {Version: "1.0.0", Params: types.Params{"memory": "1024"}},
				}),
			},
			b: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{
					"cap1": {Version: "1.1.0", Params: types.Params{"memory": "1024"}},
				}),
			},
			want: map[string][]string{"node1": {"cap1"}},
		},
		{
			name: "downgraded version is an error",
			a: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{
					"cap1": {Version: "1.1.0"},
				}),
			},
			b: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{
					"cap1": {Version: "1.0.0"},
				}),
			},
			wantErr: "capability cap1 of node1 has older version in target",
		},
		{
			name: "invalid version in initial state is an error",
			a: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{
					"cap1": {Version: "not-a-version"},
				}),
			},
			b: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{
					"cap1": {Version: "1.0.0"},
				}),
			},
			wantErr: "invalid semantic version",
		},
		{
			name: "invalid version in target state is an error",
			a: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{
					"cap1": {Version: "1.0.0"},
				}),
			},
			b: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{
					"cap1": {Version: "not-a-version"},
				}),
			},
			wantErr: "invalid semantic version",
		},
		{
			// Node iteration order and capability-map iteration order are both
			// randomized by Go, so this case deliberately mixes an unchanged,
			// a changed, and a removed capability on the same node to make sure
			// each is evaluated independently rather than the loop short-circuiting.
			name: "multiple capabilities on the same node are evaluated independently",
			a: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{
					"unchanged": {Version: "1.0.0", Params: types.Params{"a": "1"}},
					"changed":   {Version: "1.0.0", Params: types.Params{"a": "1"}},
					"removed":   {Version: "1.0.0"},
				}),
			},
			b: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{
					"unchanged": {Version: "1.0.0", Params: types.Params{"a": "1"}},
					"changed":   {Version: "1.0.0", Params: types.Params{"a": "2"}},
				}),
			},
			want: map[string][]string{"node1": {"changed"}},
		},
		{
			name: "multiple nodes are evaluated independently",
			a: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{"cap1": {Version: "1.0.0"}}),
				"node2": newNode(cap.CapabilityList{"cap1": {Version: "1.0.0"}}),
			},
			b: map[string]*node.Node{
				"node1": newNode(cap.CapabilityList{"cap1": {Version: "1.0.0"}}),
				"node2": newNode(cap.CapabilityList{"cap1": {Version: "2.0.0"}}),
			},
			want: map[string][]string{"node2": {"cap1"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := differenceCapChange(tt.a, tt.b, getNodeCaps)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// capability-map iteration order is randomized; sort before comparing.
			for _, ids := range got {
				sort.Strings(ids)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("differenceCapChange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_differenceCapChange_cluster(t *testing.T) {
	// Sanity check that the generic function behaves the same way when
	// instantiated with *cluster.Cluster instead of *node.Node.
	getClusterCaps := func(c *cluster.Cluster) cap.CapabilityList { return c.Capabilities }

	a := map[string]*cluster.Cluster{
		"cluster1": {Capabilities: cap.CapabilityList{"cap1": {Version: "1.0.0"}}},
	}
	b := map[string]*cluster.Cluster{
		"cluster1": {Capabilities: cap.CapabilityList{"cap1": {Version: "1.1.0"}}},
	}

	got, err := differenceCapChange(a, b, getClusterCaps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string][]string{"cluster1": {"cap1"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("differenceCapChange() = %v, want %v", got, want)
	}
}
