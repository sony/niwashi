// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cluster_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/cluster"
	"github.com/sony/niwashi/internal/types"
)

func TestCluster_Validate(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		cluster *cluster.Cluster
		wantErr bool
	}{
		{
			name:    "empty cluster",
			cluster: &cluster.Cluster{},
			wantErr: false,
		},
		// { // todo: add test case for id
		// 	name: "invalid capability id",
		// 	cluster: &cluster.Cluster{
		// 		Capabilities: cap.CapabilityList{
		// 			"invalid-id": &cap.Capability{},
		// 		},
		// 	},
		// 	wantErr: true,
		// },
		{
			name: "invalid param name",
			cluster: &cluster.Cluster{
				Params: map[string]any{
					"invalid-name": "value", // ng
					"invalid.name": "value", // ng
					"invalid name": "value", // ng
					"0":            "value", // ng
					".":            "value", // ng
					"a":            "value", // ok
				},
			},
			wantErr: true,
		},
		{
			name: "invalid capability store",
			cluster: &cluster.Cluster{
				Capabilities: cap.CapabilityList{
					"cap1": &cap.Capability{
						Store: map[string]any{
							"invalid-name": "value", // ng
							"invalid.name": "value", // ng
							"invalid name": "value", // ng
							"0":            "value", // ng
							".":            "value", // ng
							"a":            "value", // ok
						},
					},
				},
			},
			wantErr: true,
		},
		{ // Remove this test case after deprecating cluster labels.
			name: "invalid label key",
			cluster: &cluster.Cluster{
				Labels: map[string]string{
					"invalid-key": "value", // ng
					"invalid.key": "value", // ng
					"invalid key": "value", // ng
					"0":           "value", // ng
					".":           "value", // ng
					"a":           "value", // ok
				},
			},
			wantErr: true,
		},
		{
			name: "invalid node label key",
			cluster: &cluster.Cluster{
				Nodes: cluster.NodeList{
					"node-1": &cluster.Node{
						Labels: map[string]string{
							"invalid-key": "value", // ng
							"invalid.key": "value", // ng
							"invalid key": "value", // ng
							"0":           "value", // ng
							".":           "value", // ng
							"a":           "value", // ok
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.cluster.Validate()
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

// TestCluster_ApplyTemplate_Precedence follows the user guide (templates.md):
// the definition of the cluster overrides its templates, and a later template
// overrides an earlier one.
func TestCluster_ApplyTemplate_Precedence(t *testing.T) {
	tpl := map[string]*cluster.Template{
		"template-a": {
			Version: "1.27",
			Params:  types.Params{"network": "calico", "replicas": 1, "region": "us-east"},
			Capabilities: cap.CapabilityList{
				"org/k8s": {Params: types.Params{"cni": "calico", "proxy": "iptables"}},
			},
		},
		"template-b": {
			Version: "1.28",
			Params:  types.Params{"network": "cilium", "replicas": 3},
			Capabilities: cap.CapabilityList{
				"org/k8s": {Params: types.Params{"cni": "cilium"}},
			},
		},
	}

	t.Run("cluster overrides the templates", func(t *testing.T) {
		c := &cluster.Cluster{
			Templates: []string{"template-a", "template-b"},
			Version:   "1.29",
			Params:    types.Params{"replicas": 5},
			Capabilities: cap.CapabilityList{
				"org/k8s": {Params: types.Params{"cni": "flannel"}},
			},
		}
		if err := c.ApplyTemplate(tpl); err != nil {
			t.Fatalf("ApplyTemplate() error = %v", err)
		}
		if c.Version != "1.29" {
			t.Errorf("version = %q, want %q (the cluster overrides the templates)", c.Version, "1.29")
		}
		wantParams := types.Params{
			"replicas": 5,         // the cluster overrides both templates
			"network":  "cilium",  // template-b overrides template-a
			"region":   "us-east", // only in template-a
		}
		if diff := cmp.Diff(wantParams, c.Params); diff != "" {
			t.Errorf("params mismatch (-want +got):\n%s", diff)
		}
		wantCapParams := types.Params{
			"cni":   "flannel",  // the cluster overrides both templates
			"proxy": "iptables", // only in template-a
		}
		if diff := cmp.Diff(wantCapParams, c.Capabilities["org/k8s"].Params); diff != "" {
			t.Errorf("capability params mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("later template overrides an earlier one", func(t *testing.T) {
		c := &cluster.Cluster{Templates: []string{"template-a", "template-b"}}
		if err := c.ApplyTemplate(tpl); err != nil {
			t.Fatalf("ApplyTemplate() error = %v", err)
		}
		if c.Version != "1.28" {
			t.Errorf("version = %q, want %q (template-b overrides template-a)", c.Version, "1.28")
		}
		if got := c.Capabilities["org/k8s"].Params["cni"]; got != "cilium" {
			t.Errorf("capability param cni = %v, want %q (template-b overrides template-a)", got, "cilium")
		}
	})
}
