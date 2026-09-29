// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow

import (
	"context"
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/logger"
)

type Engine struct {
	maxConcurrency uint
	listener       Listener
	taskScheduler  Scheduler
}

const (
	MinConcurrency     uint = 1
	DefaultConcurrency uint = 5
	MaxConcurrency     uint = 1000
)

type Listener interface {
	OnTaskComplete(caller *Engine, task TaskResult)
}

type TaskResult interface {
	GetTask() Task
	GetError() error
}

type taskResult struct {
	Task Task
	Err  error
}

func (r *taskResult) GetTask() Task {
	return r.Task
}

func (r *taskResult) GetError() error {
	return r.Err
}

func NewEngine(ops ...func(*Engine)) *Engine {
	e := &Engine{
		maxConcurrency: DefaultConcurrency,
	}

	for _, op := range ops {
		op(e)
	}

	return e
}

func WithScheduler(s Scheduler) func(*Engine) {
	return func(e *Engine) {
		e.taskScheduler = s
	}
}

func WithListener(l Listener) func(*Engine) {
	return func(e *Engine) {
		e.listener = l
	}
}

func WithMaxConcurrency(max uint) func(*Engine) {
	if max < MinConcurrency || max > MaxConcurrency {
		panic(fmt.Sprintf("incorrect max value: %d", max))
	}

	return func(e *Engine) {
		e.maxConcurrency = max
	}
}

func (e *Engine) RaiseError(err error) {
	e.taskScheduler.HoldError(err)
}

func (e *Engine) Validate() error {
	if e.maxConcurrency <= 0 || e.maxConcurrency > 1000 {
		return fmt.Errorf("incorrect max value: %d", e.maxConcurrency)
	}
	if e.taskScheduler == nil {
		return fmt.Errorf("taskScheduler cannot be nil")
	}
	return nil
}

func (e *Engine) Run(ctx context.Context) error {

	if err := e.Validate(); err != nil {
		return err
	}

	ch := make(chan taskResult, e.maxConcurrency)

	var working uint = 0
	// Loop until all tasks are done
	for {
		// Run as many tasks as possible
		for working < e.maxConcurrency {
			if e.postTask(ctx, ch) {
				working++
			} else {
				break
			}
		}

		if working == 0 {
			if e.taskScheduler.HasError() {
				return e.taskScheduler.GetError()
			}

			if !e.taskScheduler.IsEmpty() {
				return fmt.Errorf("deadlock detected: no tasks running but scheduler is not empty")
			}

			// no erros and no tasks left
			// all tasks completed
			logger.Debug("All tasks completed")
			return nil

		} else if working > 0 {
			select {
			case res := <-ch:
				e.taskDone(&res)
				working--
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func (e *Engine) postTask(ctx context.Context, ch chan taskResult) bool {

	t := e.taskScheduler.NextTask()
	if t == nil {
		return false
	}

	LogTask(TaskStart, t.GetJob(), t)
	e.goTask(ctx, ch, t)
	return true
}

func (e *Engine) taskDone(res TaskResult) {

	runErr := res.GetError()

	// Invoke AfterRun first with the run error
	err := res.GetTask().AfterRun(runErr)

	var finalErr error
	if runErr != nil || err != nil {
		finalErr = errors.Join(runErr, err)
	}

	// consume the task result with Run() and AfterRun() errors
	e.consume(&taskResult{Task: res.GetTask(), Err: finalErr})
}

func (e *Engine) consume(res TaskResult) {

	// notify task completion to scheduler
	e.taskScheduler.OnTaskComplete(res)

	t := res.GetTask()
	LogTaskComplete(t.GetJob(), t, res.GetError())

	if e.listener != nil {
		e.listener.OnTaskComplete(e, res)
	}
}

func (e *Engine) goTask(ctx context.Context, ch chan taskResult, t Task) {
	go func(task Task) {
		LogTaskContext(ctx, TaskRun, task.GetJob(), task, nil)
		err := task.Run(ctx)
		LogTaskContext(ctx, TaskDone, task.GetJob(), task, err)
		ch <- taskResult{Task: task, Err: err}
	}(t)
}
