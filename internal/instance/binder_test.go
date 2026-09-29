// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package instance_test

import (
	"testing"

	"github.com/sony/niwashi/internal/instance"
)

func TestBinder_Bind(t *testing.T) {

	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		generators map[string]*instance.Generator
		// Named input parameters for target function.
		selector *instance.Selector
		want     string
		want2    string
		wantErr  bool
	}{
		{
			name: "complete match",
			generators: map[string]*instance.Generator{
				"gen1": {
					Instances: map[string]*instance.Instance{
						"inst1": {NodeRef: "node1"},
						"inst2": {},
						"inst3": {NodeRef: "node3"},
					},
				},
			},
			selector: &instance.Selector{
				Generator: "gen1",
				Instance:  "inst2",
			},
			want:    "gen1",
			want2:   "inst2",
			wantErr: false,
		},
		{
			name: "search match",
			generators: map[string]*instance.Generator{
				"gen1": {
					Instances: map[string]*instance.Instance{
						"inst1": {NodeRef: "node1"},
						"inst2": {},
						"inst3": {NodeRef: "node3"},
					},
				},
			},
			selector: &instance.Selector{
				Generator: "gen1",
			},
			want:    "gen1",
			want2:   "inst2",
			wantErr: false,
		},
		{
			name: "search available",
			generators: map[string]*instance.Generator{
				"gen1": {
					Instances: map[string]*instance.Instance{
						"inst1": {NodeRef: "node1"},
						"inst3": {NodeRef: "node2"},
					},
				},
				"gen2": {
					Instances: map[string]*instance.Instance{
						"inst1": {NodeRef: "node3"},
					},
				},
				"gen3": {
					Instances: map[string]*instance.Instance{
						"inst1": {NodeRef: ""}, // <- available instance
						"inst2": {NodeRef: "node4"},
					},
				},
			},
			selector: &instance.Selector{},
			want:     "gen3",
			want2:    "inst1",
			wantErr:  false,
		},
		{
			name: "cannot bind due to no available instances",
			generators: map[string]*instance.Generator{
				"gen1": {
					Instances: map[string]*instance.Instance{
						"inst1": {NodeRef: "node1"},
						"inst2": {NodeRef: "node2"},
					},
				},
			},
			selector: &instance.Selector{
				Generator: "gen2",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := instance.NewBinder(tt.generators)
			got, got2, gotErr := f.Bind(tt.selector)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Bind() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Bind() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("Bind() = %v, want %v", got, tt.want)
			}
			if got2 != tt.want2 {
				t.Errorf("Bind() = %v, want %v", got2, tt.want2)
			}
		})
	}
}
