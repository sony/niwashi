// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_cluster

import (
	"fmt"

	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/planner"
)

type ClusterCapPlanner struct{}

func (p *ClusterCapPlanner) BuildVertices(ctx planner.Context) error {
	diff := ctx.Diff()

	op := ctx.Operation()

	// Cluster capabilities
	for clusterName, c := range diff.Capability.Cluster {

		cluster := ctx.State().GetCluster(clusterName)
		if cluster == nil {
			return fmt.Errorf("failed to find cluster %q in state", clusterName)
		}

		for _, capId := range c {

			cap := cluster.Capabilities.Find(capId)
			if cap == nil {
				return fmt.Errorf("failed to find capability %q in cluster %q", capId, clusterName)
			}

			r, err := ctx.FindRecipe(cap.SearchId(capId))
			if err != nil {
				return err
			}

			// if no tasks are generated, skip this Vertex
			if op == recipe.OperationUpdate &&
				planner.GetTaskCount(ctx, r) == 0 {
				logger.Warn("capability changed but its recipe has no update tasks; the change is not applied", "capability", capId, "cluster", clusterName)
				ctx.AddSkippedRecipe(workflow.PhaseCluster, clusterName, r.Fqid())
				continue
			}

			err = planner.ResolveToolAlias(ctx.Operation(), ctx, r)
			if err != nil {
				return err
			}

			vtx, err := planner.NewRecipeDagVertex(
				workflow.PhaseCluster, clusterName, r, cap.Params, ctx, diff.BaseVertexData)
			if err != nil {
				return err
			}

			ctx.AddVertex(vtx)
		}

	}

	return nil
}

func (p *ClusterCapPlanner) BuildEdges(ctx planner.Context, v workflow.Vertex) error {
	vtx := v.(*planner.RecipeDagVertex)
	return vtx.BuildRequiresEdge(ctx)
}

func (p *ClusterCapPlanner) CanOmit() bool {
	return false
}
