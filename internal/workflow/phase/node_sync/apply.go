// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_node_sync

import (
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/workflow"
)

func init() {

	workflow.RegisterTaskGeneratorFactory(
		workflow.PhaseNodeSync,
		func(ctx workflow.ApplyContext, job *workflow.Job) (workflow.TaskGenerator, error) {
			return NewNodeSyncTaskGenerator(ctx, job)
		},
	)

}

const (
	nodeSyncInit int = iota
	nodeSyncBinding
	nodeSyncProbing
	nodeSyncComplete
)

type NodeSyncTaskGenerator struct {
	workflow.DefaultErrorHolder
	workflow.TaskQueue
	ctx workflow.ApplyContext
	job *workflow.Job

	stateMachine *workflow.StateMachine
}

func NewNodeSyncTaskGenerator(ctx workflow.ApplyContext, job *workflow.Job) (*NodeSyncTaskGenerator, error) {

	g := &NodeSyncTaskGenerator{
		ctx: ctx,
		job: job,
	}

	stateMachine := workflow.NewStateMachine(nodeSyncInit, nodeSyncComplete)
	stateMachine.AddState(nodeSyncBinding, &NodeBindState{g: g})
	stateMachine.AddState(nodeSyncProbing, &InstanceProbeState{g: g})

	stateMachine.AddTransitions(map[int]int{
		nodeSyncInit:    nodeSyncProbing,
		nodeSyncProbing: nodeSyncBinding,
		nodeSyncBinding: nodeSyncComplete,
	})

	g.stateMachine = stateMachine

	return g, nil
}

func (g *NodeSyncTaskGenerator) GetDryRunMode() string {
	return g.ctx.RunContext().DryRun()
}

func (g *NodeSyncTaskGenerator) GetJob() *workflow.Job {
	return g.job
}

func (g *NodeSyncTaskGenerator) GetStateManager() workflow.StateManager {
	return g.ctx
}

func (g *NodeSyncTaskGenerator) GetFileSystem() file.FileSystem {
	return g.ctx.FileSystem()
}

func (g *NodeSyncTaskGenerator) Generate() workflow.Task {
	if g.HasError() {
		return nil
	}
	err := g.stateMachine.Execute()
	if err != nil {
		g.HoldError(err)
		return nil
	}
	return g.PopTask()
}

func (g *NodeSyncTaskGenerator) OnComplete(result workflow.TaskResult) error {
	g.Complete(result.GetTask())

	// Hold error to stop generating new tasks
	g.HoldError(result.GetError())

	return nil
}

func (g *NodeSyncTaskGenerator) IsComplete() bool {
	return g.stateMachine.IsComplete()
}

type nodeProbeTaskContext struct {
	*NodeSyncTaskGenerator
	generatorName, instanceName string
}

func (c *nodeProbeTaskContext) GetInstanceInfo() (string, string) {
	return c.generatorName, c.instanceName
}

type NodeBindState struct {
	workflow.EmptyState
	g *NodeSyncTaskGenerator
}

func (s *NodeBindState) OnEnter() {
	s.CreateSyncTask()
}

func (s *NodeBindState) Satisfy() bool {
	return s.g.IsEmpty()
}

func (s *NodeBindState) CreateSyncTask() {

	t, err := buildNodeBindTask(s.g.ctx, s.g.job)
	s.g.HoldError(err)
	s.g.AddTask(t)
}

type InstanceProbeState struct {
	workflow.EmptyState
	g *NodeSyncTaskGenerator
}

func (s *InstanceProbeState) OnEnter() {
	s.CreateProbeTasks()
}

func (s *InstanceProbeState) Satisfy() bool {
	if s.g.HasError() {
		// wait until all tasks are completed to return the error
		return !s.g.HasWorkingTask()
	}
	return s.g.IsEmpty()
}

func (s *InstanceProbeState) CreateProbeTasks() {

	instances := GetUnprobedInstances(s.g.ctx.State())
	for generatorName, instanceNames := range instances {
		for _, instanceName := range instanceNames {
			t, err := NewNodeProbeTask(&nodeProbeTaskContext{
				NodeSyncTaskGenerator: s.g,
				generatorName:         generatorName,
				instanceName:          instanceName,
			})

			if err != nil {
				s.g.HoldError(err)
				return
			}
			s.g.AddTask(t)
		}
	}
}
