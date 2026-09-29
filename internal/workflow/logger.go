// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow

import (
	"context"

	"github.com/sony/niwashi/internal/logger"
)

type Event int

const (
	JobEvent Event = iota // job any event
	JobStart
	JobComplete
	TaskEvent
	TaskGenerate
	TaskStart
	TaskRun
	TaskDone
	TaskComplete
	PhaseStart
	PlanStart
	PlanComplete
)

func (e Event) String() string {
	switch e {
	case JobEvent:
		return "JobEvent"
	case JobStart:
		return "JobStart"
	case JobComplete:
		return "JobComplete"

	case TaskEvent:
		return "TaskEvent"
	case TaskGenerate:
		return "TaskGenerate"
	case TaskStart:
		return "TaskStart"
	case TaskRun:
		return "TaskRun"
	case TaskDone:
		return "TaskDone"
	case TaskComplete:
		return "TaskComplete"

	case PhaseStart:
		return "PhaseStart"

	case PlanStart:
		return "PlanStart"
	case PlanComplete:
		return "PlanComplete"

	default:
		return "UnknownEvent"

	}
}

func LogJob(event Event, job *Job, args ...any) {
	LogJobContext(context.Background(), event, job, nil, args...)
}

func LogJobComplete(job *Job, err error, args ...any) {
	LogJobContext(context.Background(), JobComplete, job, err, args...)
}

func LogJobContext(ctx context.Context, event Event, job *Job, err error, args ...any) {
	a := Compact(job, nil, err, args)
	logger.ProgressContext(ctx, event.String(), a...)
}

func LogTask(event Event, job *Job, task Task, args ...any) {
	LogTaskContext(context.Background(), event, job, task, nil, args...)
}

func LogTaskComplete(job *Job, task Task, err error, args ...any) {
	LogTaskContext(context.Background(), TaskComplete, job, task, err, args...)
}

func LogTaskContext(ctx context.Context, event Event, job *Job, task Task, err error, args ...any) {
	a := Compact(job, task, err, args)
	logger.ProgressContext(ctx, event.String(), a...)
}

func LogTaskError(job *Job, task Task, err error) {
	logger.Error("Task error", "job", job.Id, "task", task.GetName(), "subject", task.GetSubject(), "error", err)
}

func LogProgress(event Event, args ...any) {
	logger.Progress(event.String(), args...)
}

func Compact(job *Job, task Task, err error, args []any) []any {
	a := []any{}
	if job != nil {
		a = append(a, "job", job.Id)
	}
	if task != nil {
		a = append(a, "task", task.GetName(), "subject", task.GetSubject())
	}
	if err != nil {
		a = append(a, "error", err)
	}
	return append(a, args...)
}
