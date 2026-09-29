// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_cluster

import (
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/planner"
	"github.com/sony/niwashi/internal/workflow/recipe_exec"
)

const MyPhase = workflow.PhaseCluster

func init() {

	workflow.AddPhase(
		MyPhase,
		workflow.PhaseClusterPriority,
		[]string{
			workflow.PhaseHost,
			workflow.PhaseNode,
		},
	)

	planner.RegisterPhasePlanner(
		MyPhase,
		&ClusterCapPlanner{},
	)

	workflow.RegisterTaskGeneratorFactory(
		MyPhase,
		NewClusterTaskGenerator,
	)

	recipe_exec.RegisterOperator(
		MyPhase,
		&Operator{},
	)
}
