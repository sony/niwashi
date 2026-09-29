// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_cluster

import (
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/recipe_exec"
)

func NewClusterTaskGenerator(ctx workflow.ApplyContext, job *workflow.Job) (workflow.TaskGenerator, error) {

	rtg, err := recipe_exec.NewRecipeTaskGenerator(ctx, job)
	if err != nil {
		return nil, err
	}

	// Add listener to add/remove capability when all tasks done
	prefix := "/inventory/clusters"
	switch rtg.GetOperation() {
	case recipe.OperationConstruct:
		rtg.GetRunState().AddEnterFunc(func() {
			err := recipe_exec.AddCapability(
				prefix,
				ctx,
				rtg.GetRecipe(),
				rtg.GetRecipeWs().Target,
				job)

			if err != nil {
				rtg.HoldError(err)
			}
		})

		// remove capability on exit if error occurs
		rtg.GetRunState().AddFailureFunc(func() {
			err := recipe_exec.RemoveCapability(
				prefix,
				ctx,
				rtg.GetRecipe(),
				rtg.GetRecipeWs().Target)
			if err != nil {
				rtg.HoldError(err)
			}
		})
	case recipe.OperationUpdate:
		rtg.GetRunState().AddSuccessFunc(func() {
			err := recipe_exec.UpdateCapability(
				prefix,
				ctx,
				rtg.GetRecipe(),
				rtg.GetRecipeWs().Target,
				job)
			if err != nil {
				rtg.HoldError(err)
			}
		})
	case recipe.OperationDestruct:
		rtg.GetRunState().AddSuccessFunc(func() {
			err := recipe_exec.RemoveCapability(
				prefix,
				ctx,
				rtg.GetRecipe(),
				rtg.GetRecipeWs().Target)
			if err != nil {
				rtg.HoldError(err)
			}
		})
	}

	return rtg, nil
}
