// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_cluster_sync

import (
	"fmt"

	"github.com/sony/niwashi/internal/types"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/planner"
)

func init() {
	workflow.AddPhase(workflow.PhaseClusterSync, workflow.PhaseClusterSyncPriority, nil)
	planner.RegisterPhasePlanner(workflow.PhaseClusterSync, &ClusterSyncPlanner{})
}

type ClusterSyncDagVertex struct {
	data types.Dict
}

func NewClusterSyncDagVertex(clusterSync types.Dict) *ClusterSyncDagVertex {
	return &ClusterSyncDagVertex{
		data: clusterSync,
	}
}

func (d *ClusterSyncDagVertex) GetId() string {
	return "cluster_sync"
}

func (d *ClusterSyncDagVertex) GetData() any {
	return d.data
}

type ClusterSyncPlanner struct{}

func (p *ClusterSyncPlanner) BuildVertices(ctx planner.Context) error {
	diff := ctx.Diff()

	cluster_sync := []string{}
	for _, name := range diff.Cluster {
		c := ctx.State().GetCluster(name)
		if c == nil {
			return fmt.Errorf("failed to find cluster %q in state", name)
		}

		cluster_sync = append(cluster_sync, name)
	}

	if len(cluster_sync) > 0 {
		ctx.AddVertex(NewClusterSyncDagVertex(types.Dict{"clusters": cluster_sync}))
	}

	return nil
}

func (p *ClusterSyncPlanner) BuildEdges(ctx planner.Context, vertex workflow.Vertex) error {
	return nil
}

func (p *ClusterSyncPlanner) CanOmit() bool {
	return false
}
