// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_infra

import (
	"fmt"

	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/planner"
)

type InfraPlanner struct{}

func (p *InfraPlanner) BuildVertices(ctx planner.Context) error {
	diff := ctx.Diff()
	s := ctx.State()

	for _, name := range diff.Instance {

		generator := s.GetGenerator(name)
		if generator == nil {
			return fmt.Errorf("failed to find generator %q in state", name)
		}
		if generator.Provisioner == "" {
			return fmt.Errorf("generator %q has no provisioner", name)
		}

		r, err := ctx.FindRecipe(generator.Provisioner)
		if err != nil {
			return fmt.Errorf("failed to find recipe %q: %w", generator.Provisioner, err)
		}

		vtx, err := planner.NewRecipeDagVertex(workflow.PhaseInfra, name, r, generator.Params, ctx, diff.BaseVertexData)
		if err != nil {
			return err
		}

		ctx.AddVertex(vtx)
	}

	return nil
}

func (p *InfraPlanner) BuildEdges(ctx planner.Context, v workflow.Vertex) error {
	vtx := v.(*planner.RecipeDagVertex)
	return vtx.BuildRequiresEdge(ctx)
}

func (p *InfraPlanner) CanOmit() bool {
	return false
}
