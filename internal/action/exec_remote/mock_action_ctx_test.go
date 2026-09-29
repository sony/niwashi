// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_remote

import (
	"github.com/sony/niwashi/internal/action"
)

type DefaultActionContext struct {
	runContext action.RunContext
	task       action.TaskContext
	localEnv   action.Workspace
	subject    action.Subject
	tplParams  *action.TemplateParams
}

func NewDefaultActionContext(opts ...func(*DefaultActionContext)) *DefaultActionContext {
	ctx := &DefaultActionContext{
		runContext: NewDefaultRunContext(),
	}
	ctx.Set(opts...)
	return ctx
}

func (ac *DefaultActionContext) Set(opts ...func(*DefaultActionContext)) {
	for _, opt := range opts {
		opt(ac)
	}
}

func WithRunContext(r action.RunContext) func(*DefaultActionContext) {
	return func(ctx *DefaultActionContext) {
		ctx.runContext = r
	}
}

func WithTask(t action.TaskContext) func(*DefaultActionContext) {
	return func(ctx *DefaultActionContext) {
		ctx.task = t
	}
}

func WithWorkspace(e action.Workspace) func(*DefaultActionContext) {
	return func(ctx *DefaultActionContext) {
		ctx.localEnv = e
	}
}

func WithSubject(subject action.Subject) func(*DefaultActionContext) {
	return func(ctx *DefaultActionContext) {
		ctx.subject = subject
	}
}

func WithTemplateParams(params *action.TemplateParams) func(*DefaultActionContext) {
	return func(ctx *DefaultActionContext) {
		ctx.tplParams = params
	}
}

func (a *DefaultActionContext) RunContext() action.RunContext {
	return a.runContext
}

func (a *DefaultActionContext) Task() action.TaskContext {
	return a.task
}

func (a *DefaultActionContext) Workspace() action.Workspace {
	return a.localEnv
}

func (a *DefaultActionContext) Subject() action.Subject {
	return a.subject
}

func (a *DefaultActionContext) TemplateParams() *action.TemplateParams {
	return a.tplParams
}
