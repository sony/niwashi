// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_cluster_sync

import (
	"context"
	"fmt"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/applier"
)

func init() {
	workflow.RegisterTaskGeneratorFactory(
		workflow.PhaseClusterSync,
		func(ctx workflow.ApplyContext, job *workflow.Job) (workflow.TaskGenerator, error) {
			return workflow.NewBasicTaskGenerator(
				ctx,
				job,
				buildClusterSyncTask,
			), nil
		},
	)
}

func buildClusterSyncTask(ctx workflow.ApplyContext, job *workflow.Job) (workflow.Task, error) {
	rawData, ok := job.GetFromData("clusters")
	if !ok {
		return nil, fmt.Errorf("invalid data in ClusterSyncTask: %v", job.Data)
	}
	rawSlice, ok := rawData.([]any)
	if !ok {
		return nil, fmt.Errorf("no string arrays in ClusterSyncTask: %v", rawData)
	}

	clusters := make([]string, 0, len(rawSlice))
	for i, v := range rawSlice {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("invalid cluster name at index %d: %v", i, v)
		}
		clusters = append(clusters, str)
	}

	switch ctx.Operation() {
	case recipe.OperationConstruct:
		return &ClusterConstructSyncTask{
			SystemTask: applier.NewSystemTask("sync clusters", job),
			ctx:        ctx,
			Clusters:   clusters,
		}, nil
	case recipe.OperationDestruct:
		return &ClusterDestructSyncTask{
			SystemTask: applier.NewSystemTask("sync clusters", job),
			ctx:        ctx,
			Clusters:   clusters,
		}, nil
	}
	return nil, fmt.Errorf("unsupported operation for ClusterTask: %s", ctx.Operation())
}

type ClusterConstructSyncTask struct {
	applier.SystemTask
	ctx      workflow.ApplyContext
	Clusters []string
}

func (ct *ClusterConstructSyncTask) Run(_ context.Context) error {
	// do nothing
	return nil
}

func (ct *ClusterConstructSyncTask) AfterRun(runErr error) error {
	if runErr != nil {
		return nil
	}

	// create patches to sync cluster info
	s := ct.ctx.TargetState()
	patches := []patch.Patch{}

	for _, clusterName := range ct.Clusters {
		c := s.GetCluster(clusterName)
		if c == nil {
			return fmt.Errorf("failed to find cluster %q in target state", clusterName)
		}

		if ct.ctx.State().GetCluster(clusterName) != nil {
			continue // skip if cluster already exists in current state
		}

		cc := c.Clone()
		cc.Capabilities = cap.CapabilityList{} // don't copy capabilities

		patches = append(
			patches,
			patch.NewAddPatch("/inventory/clusters/"+clusterName, cc),
		)
	}

	return ct.ctx.ApplyPatches(patches, nil, nil)
}

type ClusterDestructSyncTask struct {
	applier.SystemTask
	ctx      workflow.ApplyContext
	Clusters []string
}

func (ct *ClusterDestructSyncTask) Run(_ context.Context) error {
	// do nothing
	return nil
}

func (ct *ClusterDestructSyncTask) AfterRun(runErr error) error {
	if runErr != nil {
		return nil
	}

	patches := []patch.Patch{}

	// Remove cluster info without capabilities
	for _, clusterName := range ct.Clusters {
		if ct.ctx.State().GetCluster(clusterName) == nil {
			continue // skip if cluster does not exist in current state
		}

		patches = append(
			patches,
			patch.NewRemovePatch("/inventory/clusters/"+clusterName),
		)
	}

	return ct.ctx.ApplyPatches(patches, nil, nil)
}
