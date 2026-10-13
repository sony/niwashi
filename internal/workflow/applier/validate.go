// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workflow"
)

func ValidateState(finder cmap.RecipeFinder, s *state.State) error {

	if err := s.Validate(); err != nil {
		return err
	}

	return nil
}

func ValidatePlan(finder cmap.RecipeFinder, pln *plan.Plan) error {

	if err := pln.Validate(); err != nil {
		return err
	}

	validateWorkflow := func(wf *plan.Workflow) error {
		if wf == nil {
			return nil
		}

		for _, job := range wf.Jobs {
			if err := workflow.ValidateGenerator(job); err != nil {
				return err
			}
		}

		return nil
	}

	for stage, wf := range pln.ExecutionPlan.ConstructGraph {
		if err := validateWorkflow(wf); err != nil {
			return errors.Join(err, fmt.Errorf("invalid construct workflow in plan for stage %s: %w", stage, err))
		}
	}
	for stage, wf := range pln.ExecutionPlan.DestructGraph {
		if err := validateWorkflow(wf); err != nil {
			return errors.Join(err, fmt.Errorf("invalid destruct workflow in plan for stage %s: %w", stage, err))
		}
	}

	return nil
}
