// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package state

import "time"

type Status struct {
	ActiveApply ActiveApply `json:"activeApply,omitempty" yaml:"activeApply,omitempty"`
}

func (s *Status) Clone() *Status {
	if s == nil {
		return nil
	}

	clone := &Status{}
	clone.ActiveApply = s.ActiveApply.Clone()

	return clone
}

func (s *Status) SetDefaults() {
	// no-op
	s.ActiveApply.SetDefaults()
}

type ActiveApply struct {
	PlanHash       string              `json:"planHash,omitempty" yaml:"planHash,omitempty"`
	RunId          string              `json:"runId,omitempty" yaml:"runId,omitempty"`
	StateHash      string              `json:"stateHash,omitempty" yaml:"stateHash,omitempty"`
	StartedAt      time.Time           `json:"startedAt,omitempty" yaml:"startedAt,omitempty"`
	CompletedTasks map[string][]string `json:"completedTasks,omitempty" yaml:"completedTasks,omitempty"`
	FailedTasks    []*FailedTask       `json:"failedTasks,omitempty" yaml:"failedTasks,omitempty"`
}

func (a *ActiveApply) ResetTasks() {
	a.CompletedTasks = make(map[string][]string)
	a.FailedTasks = []*FailedTask{}
}

func (a *ActiveApply) Set(startedAt time.Time, runId string, planHash string) {
	a.StartedAt = startedAt
	a.RunId = runId
	a.PlanHash = planHash
}

func (a *ActiveApply) Clone() ActiveApply {
	cloned := ActiveApply{}
	cloned.PlanHash = a.PlanHash
	cloned.RunId = a.RunId
	cloned.StateHash = a.StateHash
	cloned.StartedAt = a.StartedAt
	cloned.CompletedTasks = make(map[string][]string, len(a.CompletedTasks))
	for k, v := range a.CompletedTasks {
		cloned.CompletedTasks[k] = append([]string(nil), v...)
	}
	cloned.FailedTasks = make([]*FailedTask, len(a.FailedTasks))
	for i, ft := range a.FailedTasks {
		cloned.FailedTasks[i] = ft.Clone()
	}
	return cloned
}

func (a *ActiveApply) SetDefaults() {
	if a.CompletedTasks == nil {
		a.CompletedTasks = make(map[string][]string)
	}
	if a.FailedTasks == nil {
		a.FailedTasks = []*FailedTask{}
	}
	for _, ft := range a.FailedTasks {
		ft.SetDefaults()
	}
}

func (a *ActiveApply) AddCompleteTask(op, id string) {
	a.CompletedTasks[op] = append(a.CompletedTasks[op], id)
}

func (a *ActiveApply) AddErrorTask(op, id string, err error) {
	a.FailedTasks = append(a.FailedTasks, &FailedTask{
		TaskId:    id,
		Operation: op,
		FailedAt:  time.Now(),
		Error:     err.Error(),
	})
}

type FailedTask struct {
	TaskId    string    `json:"taskId,omitempty" yaml:"taskId,omitempty"`
	Operation string    `json:"operation,omitempty" yaml:"operation,omitempty"`
	FailedAt  time.Time `json:"failedAt,omitempty" yaml:"failedAt,omitempty"`
	Error     string    `json:"error,omitempty" yaml:"error,omitempty"`
}

func (f *FailedTask) Clone() *FailedTask {
	if f == nil {
		return nil
	}

	clone := &FailedTask{}
	clone.TaskId = f.TaskId
	clone.Operation = f.Operation
	clone.FailedAt = f.FailedAt
	clone.Error = f.Error
	return clone
}

func (f *FailedTask) SetDefaults() {
	// no-op
}
