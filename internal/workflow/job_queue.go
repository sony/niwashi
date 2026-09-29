// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow

import (
	"fmt"

	"github.com/sony/niwashi/internal/plan"
)

type JobQueue interface {
	NextJob() *Job
	OnJobComplete(j *Job)
	Empty() bool

	String() string
}

type SingleJobQueue struct {
	wf      *plan.Workflow
	cursor  int
	workJob *Job
}

func NewSingleJobQueue(wf *plan.Workflow) *SingleJobQueue {
	if wf == nil {
		panic("workflow cannot be nil")
	}

	return &SingleJobQueue{
		wf: wf,
	}
}

func (s *SingleJobQueue) Empty() bool {
	return s.workJob == nil && s.cursor >= len(s.wf.Jobs)
}

func (s *SingleJobQueue) NextJob() *Job {
	if s.workJob != nil || s.cursor >= len(s.wf.Jobs) {
		return nil
	}
	job := (*Job)(s.wf.Jobs[s.cursor])
	s.workJob = job
	s.cursor++
	return job
}

func (s *SingleJobQueue) OnJobComplete(j *Job) {
	s.workJob = nil
}

func (s *SingleJobQueue) String() string {
	return fmt.Sprintf("SingleJobQueue(cursor=%d, workJob=%v, total=%d)", s.cursor, s.workJob, len(s.wf.Jobs))
}
