// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow

import (
	"github.com/sony/niwashi/internal/logger"
)

type TaskQueue struct {
	queue   []Task
	working []Task
}

func (q *TaskQueue) AddTask(task Task) {
	q.queue = append(q.queue, task)
}

func (q *TaskQueue) PopTask() Task {
	if len(q.queue) == 0 {
		return nil
	}
	task := q.queue[0]
	q.queue = q.queue[1:]
	q.working = append(q.working, task)
	return task
}

func (q *TaskQueue) Complete(task Task) {
	for i, t := range q.working {
		if t == task {
			q.working = append(q.working[:i], q.working[i+1:]...)
			return
		}
	}
	logger.Error("completed task not found in working queue", "task", task.GetName())
}

func (q *TaskQueue) HasWorkingTask() bool {
	return len(q.working) > 0
}

func (q *TaskQueue) IsEmpty() bool {
	return len(q.queue) == 0 && len(q.working) == 0
}
