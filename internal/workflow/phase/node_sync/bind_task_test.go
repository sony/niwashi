// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_node_sync

import (
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/instance"
)

func Test_sort(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		originalNodes []*SelectorInfo
		want          []*SelectorInfo
	}{
		{
			name: "normal test",
			originalNodes: []*SelectorInfo{
				{NodeName: "node1", Selector: &instance.Selector{Generator: "gen1"}},
				{NodeName: "node2", Selector: &instance.Selector{Generator: "gen2"}},
				{NodeName: "node3", Selector: &instance.Selector{Generator: "gen1"}},
				{NodeName: "node4", Selector: &instance.Selector{Generator: "gen1", Instance: "inst1"}},
				{NodeName: "node5", Selector: &instance.Selector{}},
				{NodeName: "node6", Selector: &instance.Selector{Generator: "gen2"}},
				{NodeName: "node7", Selector: &instance.Selector{Generator: "gen1", Instance: "inst1"}},
				{NodeName: "node8", Selector: &instance.Selector{}},
				{NodeName: "node9", Selector: &instance.Selector{Instance: "unknown"}},
			},
			want: []*SelectorInfo{
				{NodeName: "node4", Selector: &instance.Selector{Generator: "gen1", Instance: "inst1"}},
				{NodeName: "node7", Selector: &instance.Selector{Generator: "gen1", Instance: "inst1"}},
				{NodeName: "node1", Selector: &instance.Selector{Generator: "gen1"}},
				{NodeName: "node2", Selector: &instance.Selector{Generator: "gen2"}},
				{NodeName: "node3", Selector: &instance.Selector{Generator: "gen1"}},
				{NodeName: "node6", Selector: &instance.Selector{Generator: "gen2"}},
				{NodeName: "node5", Selector: &instance.Selector{}},
				{NodeName: "node8", Selector: &instance.Selector{}},
				{NodeName: "node9", Selector: &instance.Selector{Instance: "unknown"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sort(tt.originalNodes)
			// TODO: update the condition below to compare got with tt.want.
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sort() = %v, want %v", got, tt.want)
			}
		})
	}
}
