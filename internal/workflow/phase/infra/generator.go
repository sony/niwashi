// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_infra

import (
	"fmt"

	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/recipe_exec"
)

func NewInfraTaskGenerator(ctx workflow.ApplyContext, job *workflow.Job) (workflow.TaskGenerator, error) {

	rtg, err := recipe_exec.NewRecipeTaskGenerator(ctx, job)
	if err != nil {
		return nil, err
	}

	switch ctx.Operation() {
	case recipe.OperationConstruct:
		// copy infra state before apply patches
		rtg.GetRunState().AddEnterFunc(func() {
			err := fillInfraState(ctx, ctx.TargetState(), rtg.GetRecipeWs().Target)
			if err != nil {
				rtg.HoldError(err)
			}
		},
		)

		rtg.GetRunState().AddFailureFunc(func() {
			// if error occurs during construct, remove the infra state
			err := removeInfraState(ctx, rtg.GetRecipeWs().Target)
			if err != nil {
				rtg.HoldError(err)
			}
		})

	case recipe.OperationDestruct:
		rtg.GetRunState().AddSuccessFunc(func() {
			err := removeInfraState(ctx, rtg.GetRecipeWs().Target)
			if err != nil {
				rtg.HoldError(err)
			}
		})
	}

	return rtg, nil
}

func fillInfraState(
	sm workflow.StateManager,
	targetState *state.Accessor,
	target string) error {

	// If generator info already exists, do nothing
	s := sm.State()
	if s.GetGenerator(target) != nil {
		return nil
	}

	// For infra phase
	// If some paches are created, add generator info into infrastructure
	path := fmt.Sprintf("/infrastructure/generators/%s", target)

	// copy generator info from target state
	generator := targetState.GetGenerator(target)
	if generator == nil {
		return fmt.Errorf("generator info for target %s not found in target state", target)
	}

	generatorClone := generator.Clone()
	generatorClone.SetDefaults()

	err := sm.ApplyPatches([]patch.Patch{
		patch.NewAddPatch(path, generatorClone),
	}, nil, nil)
	return err
}

func removeInfraState(sm workflow.StateManager, target string) error {

	s := sm.State()

	// If generator info not exists, do nothing
	if s.GetGenerator(target) == nil {
		return nil
	}

	err := sm.ApplyPatches([]patch.Patch{
		patch.NewRemovePatch(fmt.Sprintf("/infrastructure/generators/%s", target)),
	}, nil, nil)
	return err
}
