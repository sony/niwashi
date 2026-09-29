// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_barrier

import (
	"reflect"
	"testing"
)

func TestNewBarrierDagVertex(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		phase string
		want  *BarrierDagVertex
	}{
		{
			name:  "test simple",
			phase: "test",
			want: &BarrierDagVertex{
				id:    "barrier:test",
				phase: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewBarrierDagVertex(tt.phase)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBarrierDagVertex() = %v, want %v", got, tt.want)
			}
		})
	}
}
