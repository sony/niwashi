// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package plan_test

import (
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/state"
)

func TestNewPlan(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		want *plan.Plan
	}{
		{
			name: "default plan",
			want: &plan.Plan{
				Version: "nws.plan/v1",
				Mode:    "default",
				States: plan.States{
					Target: func() *state.State {
						s := state.NewState()
						return s
					}(),
				},
				Runtime: func() plan.Runtime {
					r := plan.Runtime{}
					r.SetDefaults()
					return r
				}(),
				Metadata: plan.NewMetadata(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := plan.NewPlan()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPlan() = %v, want %v", got, tt.want)
			}
		})
	}
}
