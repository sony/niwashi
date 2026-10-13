// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier

import (
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/workflow"
)

const testValidatePhase = "test_validate_phase"

func init() {
	workflow.RegisterTaskGeneratorFactory(testValidatePhase,
		func(ctx workflow.ApplyContext, job *workflow.Job) (workflow.TaskGenerator, error) {
			return nil, nil
		})
}

func TestValidatePlan_Stages(t *testing.T) {
	validJob := &plan.Job{Id: workflow.MakeId(testValidatePhase, "x")}
	invalidJob := &plan.Job{Id: workflow.MakeId("unknown_phase", "x")}

	tests := []struct {
		name      string
		construct map[string]*plan.Workflow
		destruct  map[string]*plan.Workflow
		wantErr   string // empty means no error
	}{
		{
			name: "valid jobs in all stages",
			construct: map[string]*plan.Workflow{
				"main":      {Jobs: []*plan.Job{validJob}},
				"bootstrap": {Jobs: []*plan.Job{validJob}},
			},
			destruct: map[string]*plan.Workflow{
				"main": {Jobs: []*plan.Job{validJob}},
			},
		},
		{
			name: "nil workflow of a stage",
			construct: map[string]*plan.Workflow{
				"main": nil,
			},
		},
		{
			name: "invalid job in a construct workflow",
			construct: map[string]*plan.Workflow{
				"main":      {Jobs: []*plan.Job{validJob}},
				"bootstrap": {Jobs: []*plan.Job{invalidJob}},
			},
			wantErr: `construct workflow in plan for stage bootstrap`,
		},
		{
			name: "invalid job in a destruct workflow",
			destruct: map[string]*plan.Workflow{
				"main": {Jobs: []*plan.Job{invalidJob}},
			},
			wantErr: `destruct workflow in plan for stage main`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pln := &plan.Plan{
				ExecutionPlan: plan.ExecutionPlan{
					ConstructGraph: tt.construct,
					DestructGraph:  tt.destruct,
				},
			}

			err := ValidatePlan(nil, pln)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("ValidatePlan() error = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ValidatePlan() error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
