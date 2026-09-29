// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow

import (
	"fmt"

	"github.com/sony/niwashi/internal/plan"
)

type TaskGenerator interface {
	ErrorHolder

	GetJob() *Job
	Generate() Task
	OnComplete(result TaskResult) error
	IsComplete() bool
	HasWorkingTask() bool
}

var taskGeneratorFactories map[string]func(ctx ApplyContext, job *Job) (TaskGenerator, error) = make(map[string]func(ctx ApplyContext, job *Job) (TaskGenerator, error))

func RegisterTaskGeneratorFactory(phase string, factory func(ctx ApplyContext, job *Job) (TaskGenerator, error)) {
	taskGeneratorFactories[phase] = factory
}

type FuncGenerateTask func(ctx ApplyContext, job *Job) (Task, error)

type BasicTaskGenerator struct {
	DefaultErrorHolder
	job         *Job
	ctx         ApplyContext
	runningTask Task
	GenTasks    []FuncGenerateTask
}

func NewBasicTaskGenerator(ctx ApplyContext, job *Job, genTasks ...FuncGenerateTask) *BasicTaskGenerator {
	return &BasicTaskGenerator{
		job:      job,
		ctx:      ctx,
		GenTasks: genTasks,
	}
}

func (g *BasicTaskGenerator) Add(genTask FuncGenerateTask) {
	g.GenTasks = append(g.GenTasks, genTask)
}

func (g *BasicTaskGenerator) GetJob() *Job {
	return g.job
}

func (g *BasicTaskGenerator) Generate() Task {
	if g.HasError() {
		return nil
	}

	if g.runningTask != nil || len(g.GenTasks) == 0 {
		return nil
	}

	task, err := g.GenTasks[0](g.ctx, g.job)
	if err != nil {
		// keep error to stop generating new tasks
		g.HoldError(err)
		return nil
	}

	g.GenTasks = g.GenTasks[1:]
	g.runningTask = task
	return task
}

func (g *BasicTaskGenerator) OnComplete(result TaskResult) error {
	g.runningTask = nil
	if result.GetError() != nil {
		g.HoldError(result.GetError())
	}
	return nil
}

func (g *BasicTaskGenerator) HasWorkingTask() bool {
	return g.runningTask != nil
}

func (g *BasicTaskGenerator) IsComplete() bool {
	return g.runningTask == nil && len(g.GenTasks) == 0
}

func NewTaskGenerator(ctx ApplyContext, job *Job) (TaskGenerator, error) {
	factory, ok := taskGeneratorFactories[job.GetPhase()]
	if !ok {
		return nil, fmt.Errorf("no task generator factory registered for phase: %s", job.GetPhase())
	}
	return factory(ctx, job)
}

func ValidateGenerator(j *plan.Job) error {
	job := (*Job)(j)
	_, ok := taskGeneratorFactories[job.GetPhase()]
	if !ok {
		return fmt.Errorf("no task generator factory registered for phase: %s", job.GetPhase())
	}

	// todo: validate job by generator
	return nil
}
