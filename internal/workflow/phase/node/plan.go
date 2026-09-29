// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_node

import (
	"fmt"

	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/planner"
)

type NodeCapPlanner struct{}

func (p *NodeCapPlanner) BuildVertices(ctx planner.Context) error {
	diff := ctx.Diff()

	op := ctx.Operation()

	// Node capabilities
	for nodeName, caps := range diff.Capability.Node {
		node := ctx.State().GetNode(nodeName)
		if node == nil {
			return fmt.Errorf("failed to find node %q in state", nodeName)
		}

		for _, capId := range caps {

			cap := node.Capabilities.Find(capId)
			if cap == nil {
				return fmt.Errorf("failed to find capability %q in node %q", capId, nodeName)
			}

			r, err := ctx.FindRecipe(cap.SearchId(capId))
			if err != nil {
				return err
			}

			// if no tasks are generated, skip this Vertex
			if op == recipe.OperationUpdate &&
				planner.GetTaskCount(ctx, r) == 0 {
				logger.Warn("capability changed but its recipe has no update tasks; the change is not applied", "capability", capId, "node", nodeName)
				ctx.AddSkippedRecipe(workflow.PhaseNode, nodeName, r.Fqid())
				continue
			}

			err = planner.ResolveToolAlias(ctx.Operation(), ctx, r)
			if err != nil {
				return err
			}

			vtx, err := planner.NewRecipeDagVertex(workflow.PhaseNode, nodeName, r, cap.Params, ctx, diff.BaseVertexData)
			if err != nil {
				return err
			}

			ctx.AddVertex(vtx)
		}
	}

	return nil
}

func (p *NodeCapPlanner) BuildEdges(ctx planner.Context, v workflow.Vertex) error {
	vtx := v.(*planner.RecipeDagVertex)
	return vtx.BuildRequiresEdge(ctx)
}

func (p *NodeCapPlanner) CanOmit() bool {
	return false
}
