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

	if err := validateWorkflow(pln.ExecutionPlan.ConstructGraph); err != nil {
		return errors.Join(err, fmt.Errorf("invalid construct workflow in plan: %w", err))
	}
	if err := validateWorkflow(pln.ExecutionPlan.DestructGraph); err != nil {
		return errors.Join(err, fmt.Errorf("invalid destruct workflow in plan: %w", err))
	}

	return nil
}
