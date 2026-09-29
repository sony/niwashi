// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_host

import (
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/planner"
)

type HostPlanner struct{}

func (p *HostPlanner) BuildVertices(ctx planner.Context) error {
	return nil
}

func (p *HostPlanner) BuildEdges(ctx planner.Context, v workflow.Vertex) error {
	vtx := v.(*planner.RecipeDagVertex)
	return vtx.BuildRequiresEdge(ctx)
}

func (p *HostPlanner) CanOmit() bool {
	return false
}
