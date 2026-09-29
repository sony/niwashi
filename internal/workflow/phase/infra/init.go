// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_infra

import (
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/planner"
	"github.com/sony/niwashi/internal/workflow/recipe_exec"
)

const MyPhase = workflow.PhaseInfra

func init() {
	workflow.AddPhase(
		MyPhase,
		workflow.PhaseInfraPriority,
		[]string{workflow.PhaseHost},
	)

	planner.RegisterPhasePlanner(
		MyPhase,
		&InfraPlanner{},
	)

	workflow.RegisterTaskGeneratorFactory(
		MyPhase,
		NewInfraTaskGenerator,
	)

	recipe_exec.RegisterOperator(
		MyPhase,
		&Operator{},
	)
}
