// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_host

import (
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/planner"
	"github.com/sony/niwashi/internal/workflow/recipe_exec"
)

const MyPhase = workflow.PhaseHost

func init() {
	workflow.AddPhase(
		MyPhase,
		workflow.PhaseHostPriority,
		[]string{},
	)

	planner.RegisterPhasePlanner(
		MyPhase,
		&HostPlanner{},
	)

	workflow.RegisterTaskGeneratorFactory(
		MyPhase,
		func(ctx workflow.ApplyContext, job *workflow.Job) (workflow.TaskGenerator, error) {
			return recipe_exec.NewRecipeTaskGenerator(ctx, job)
		},
	)

	recipe_exec.RegisterOperator(
		MyPhase,
		&Operator{},
	)
}
