// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_node

import (
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/planner"
	"github.com/sony/niwashi/internal/workflow/recipe_exec"
)

const MyPhase = workflow.PhaseNode

func init() {
	workflow.AddPhase(
		MyPhase,
		workflow.PhaseNodePriority,
		[]string{workflow.PhaseHost},
	)

	planner.RegisterPhasePlanner(
		MyPhase,
		&NodeCapPlanner{},
	)

	workflow.RegisterTaskGeneratorFactory(
		MyPhase,
		NewNodeTaskGenerator,
	)

	recipe_exec.RegisterOperator(
		MyPhase,
		&Operator{},
	)
}
