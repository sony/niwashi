// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_node_sync

import (
	"fmt"

	"github.com/sony/niwashi/internal/types"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/planner"
)

func init() {
	workflow.AddPhase(workflow.PhaseNodeSync, workflow.PhaseNodeSyncPriority, nil)

	planner.RegisterPhasePlanner(workflow.PhaseNodeSync, &NodeSyncPlanner{})
}

type NodeSyncDagVertex struct {
	data types.Dict
}

func NewNodeSyncDagVertex(nodeSync types.Dict) *NodeSyncDagVertex {
	return &NodeSyncDagVertex{
		data: nodeSync,
	}
}

func (d *NodeSyncDagVertex) GetId() string {
	return workflow.PhaseNodeSync
}

func (d *NodeSyncDagVertex) GetData() any {
	return d.data
}

type NodeSyncPlanner struct{}

func (p *NodeSyncPlanner) BuildVertices(ctx planner.Context) error {
	diff := ctx.Diff()

	node_sync := types.Dict{}

	for _, name := range diff.Node {
		node := ctx.State().GetNode(name)
		if node == nil {
			return fmt.Errorf("failed to find node %q in state", name)
		}

		node_sync[name] = node.InstanceSelector
	}

	if len(node_sync) > 0 {
		ctx.AddVertex(NewNodeSyncDagVertex(types.Dict{"nodes": node_sync}))
	}

	return nil
}

func (p *NodeSyncPlanner) BuildEdges(ctx planner.Context, vertex workflow.Vertex) error {
	return nil
}

func (p *NodeSyncPlanner) CanOmit() bool {
	return false
}
