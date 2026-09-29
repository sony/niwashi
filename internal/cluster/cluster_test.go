// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cluster_test

import (
	"testing"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/cluster"
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
