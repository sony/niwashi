// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow_test

import (
	"testing"

	"github.com/sony/niwashi/internal/workflow"
)

func TestDagGraph_TopologicalSort(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		want    []workflow.Vertex
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := workflow.NewDagGraph()
			got, gotErr := d.TopologicalSort()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("TopologicalSort() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("TopologicalSort() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("TopologicalSort() = %v, want %v", got, tt.want)
			}
		})
	}
}
