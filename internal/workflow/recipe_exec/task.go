// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import (
	"context"
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/applier"
)

type Task struct {
	actionContext *actionContext
}

type TaskContext interface {
	GetTask() TaskSpec
	GetAffectSubjects() []*SubjectId
	GetParamsFileName() string
}

func NewTask(actionContext *actionContext) *Task {
	return &Task{
		actionContext: actionContext,
	}
}

func (rt *Task) GetName() string {
	return rt.actionContext.task.GetName()
}

func (rt *Task) GetJob() *workflow.Job {
	return rt.actionContext.recipeContext.GetJob()
}

func (rt *Task) GetSubject() string {
	s := rt.actionContext.subject
	return applier.MakeSubjectFqid(s.GetScope(), s.GetId())
}

func (rt *Task) GetAffectSubjects() []string {

	ret := []string{}
	subjects := rt.actionContext.affectSubjects
	for _, s := range subjects {
		ret = append(ret, applier.MakeSubjectFqid(s.Scope, s.Name))
	}
	return ret
}

func (rt *Task) Run(ctx context.Context) error {
	a, err := rt.actionContext.recipeContext.GetApplyContext().BuildAction(rt.actionContext)
	if err != nil {
		return err
	}

	err = a.Execute(ctx)
	if err != nil {
		logger.Error("Failed to execute action", "err", err, "log", rt.actionContext.logPath)
		return errors.Join(err, fmt.Errorf("failed to execute action, see log: %s", rt.actionContext.logPath))
	}

	return nil
}

func (rt *Task) AfterRun(runErr error) error {
	return nil
}
