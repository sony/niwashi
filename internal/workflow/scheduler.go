// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow

import (
	"errors"
)

type Scheduler interface {
	ErrorHolder
	NextTask() Task
	OnTaskComplete(result TaskResult)
	IsEmpty() bool
}

type taskScheduler struct {
	DefaultErrorHolder
	ctx          ApplyContext
	generators   map[string]TaskGenerator
	allocator    TaskAllocator
	queue        JobQueue
	errorOccured bool
}

func NewTaskScheduler(ctx ApplyContext, queue JobQueue) *taskScheduler {
	return &taskScheduler{
		ctx:        ctx,
		generators: make(map[string]TaskGenerator),
		allocator:  NewTaskAllocator(),
		queue:      queue,
	}
}

func (s *taskScheduler) IsEmpty() bool {
	return len(s.generators) == 0 && s.allocator.Empty() && s.queue.Empty()
}

func (s *taskScheduler) NextTask() Task {

	for {
		// consume job and setup generators
		s.setupGenerators()

		// generate tasks and enqueue to allocator
		s.generateTasks()

		swept := s.sweepCompleteGenerators()
		if swept == 0 {
			break
		}
	}

	// allocate task
	return s.allocator.Allocate()
}

func (s *taskScheduler) setupGenerators() {

	if s.errorOccured {
		// don't add new generators when error has occurred
		return
	}

	// Add new task generators for new available jobs
	for {
		job := s.queue.NextJob()
		if job == nil {
			break
		}

		generator, err := NewTaskGenerator(s.ctx, job)
		if err != nil {
			s.errorOccured = true
			s.HoldError(err)
			return
		}

		s.generators[job.Id] = generator

		LogJob(JobStart, job)
	}
}

func (s *taskScheduler) generateTasks() {
	// generate tasks and enqueue to allocator

	for _, generator := range s.generators {

		for {
			task := generator.Generate()
			if task != nil {
				s.allocator.Enqueue(task)
				LogTask(TaskGenerate, task.GetJob(), task)
			} else {
				if generator.HasError() {
					s.errorOccured = true
				}
				break
			}
		}
	}
}

func (s *taskScheduler) sweepCompleteGenerators() int {

	// get completed generators and complete their jobs
	completed := []TaskGenerator{}
	for _, generator := range s.generators {

		// "no error and complete" or "error and all working tasks are completed"
		if CanComplete(generator) {
			// mark complete to remove generator
			completed = append(completed, generator)
		}
	}

	for _, generator := range completed {
		err := generator.GetError()
		job := generator.GetJob()
		delete(s.generators, job.Id)
		s.queue.OnJobComplete(job)
		LogJobComplete(job, err)
		if err != nil {
			s.errorOccured = true
			s.HoldError(err)
		}
	}
	return len(completed)
}

func (s *taskScheduler) OnTaskComplete(result TaskResult) {

	job := result.GetTask().GetJob()

	generator := s.generators[job.Id]
	if generator == nil {
		panic("TaskScheduler: generator not found for job ID: " + job.Id)
	}

	// notify generator and deallocate task
	onCompleteErr := generator.OnComplete(result)
	s.allocator.Deallocate(result.GetTask())

	taskErr := result.GetError()

	if taskErr != nil || onCompleteErr != nil {
		LogTaskError(job, result.GetTask(), errors.Join(taskErr, onCompleteErr))
		s.errorOccured = true
	}
}

type TaskHolder interface {
	HasError() bool
	IsComplete() bool
	HasWorkingTask() bool
}

func CanComplete(holder TaskHolder) bool {
	return (!holder.HasError() && holder.IsComplete()) ||
		(holder.HasError() && !holder.HasWorkingTask())
}
