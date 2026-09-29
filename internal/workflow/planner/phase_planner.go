// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package planner

import "github.com/sony/niwashi/internal/workflow"

type PhasePlanner interface {
	BuildVertices(ctx Context) error
	BuildEdges(ctx Context, vertex workflow.Vertex) error
	CanOmit() bool
}

var phasePlanners = map[string]PhasePlanner{}

func RegisterPhasePlanner(phase string, p PhasePlanner) {
	phasePlanners[phase] = p
}

func GetPhasePlanners() map[string]PhasePlanner {
	return phasePlanners
}

func AddPhasePlanner(phase string, planner PhasePlanner) {
	phasePlanners[phase] = planner
}
