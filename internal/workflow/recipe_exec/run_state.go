// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import "github.com/sony/niwashi/internal/workflow"

type RunState interface {
	AddEnterFunc(f func())
	AddSuccessFunc(f func())
	AddFailureFunc(f func())
}

type runState struct {
	g            *taskGenerator
	t            workflow.Task
	enterFuncs   []func()
	successFuncs []func()
	failureFuncs []func()
}

func (s *runState) AddEnterFunc(f func()) {
	s.enterFuncs = append(s.enterFuncs, f)
}

func (s *runState) AddSuccessFunc(f func()) {
	s.successFuncs = append(s.successFuncs, f)
}

func (s *runState) AddFailureFunc(f func()) {
	s.failureFuncs = append(s.failureFuncs, f)
}

func (s *runState) OnEnter() {
	for _, f := range s.enterFuncs {
		f()
	}
}

func (s *runState) Satisfy() bool {
	return s.g.canTransiteToFinal()
}

func (s *runState) OnExecute() {
	t := s.g.generate()
	if t != nil {
		s.t = t
	}
}

func (s *runState) OnExit() {
	funcs := s.successFuncs
	if s.g.HasError() {
		funcs = s.failureFuncs
	}
	for _, f := range funcs {
		f()
	}

	err := s.g.ApplyContext.SaveState()
	s.g.HoldError(err)
}

func (s *runState) Pop() workflow.Task {
	if s.t == nil {
		return nil
	}
	t := s.t
	s.t = nil
	return t
}
