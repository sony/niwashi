// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import (
	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/workspace"
)

type actionWorkspace struct {
	*workspace.TaskWorkspace
	output        action.Output
	baseCtx       *baseActionContext
	recipeTaskCtx TaskContext
}

func NewWorkspace(
	baseCtx *baseActionContext,
	recipeTaskCtx TaskContext,
	output action.Output,
) *actionWorkspace {
	return &actionWorkspace{
		TaskWorkspace: baseCtx.taskWs,
		output:        output,
		baseCtx:       baseCtx,
		recipeTaskCtx: recipeTaskCtx,
	}
}

func (w *actionWorkspace) GetOutput() action.Output {
	return w.output
}

func (w *actionWorkspace) GetFileSystem() file.FileSystem {
	return w.baseCtx.FileSystem()
}

func (w *actionWorkspace) GetRuntime() action.Runtime {
	return w.baseCtx.runtimeState.Runtime
}

func (w *actionWorkspace) GetTaskWorkspace() *workspace.TaskWorkspace {
	return w.TaskWorkspace
}

func (w *actionWorkspace) GetStateFileName() string {
	return w.baseCtx.recipeContext.GetStateFileName()
}

func (w *actionWorkspace) GetParamsFileName() string {
	return w.recipeTaskCtx.GetParamsFileName()
}

func (w *actionWorkspace) Clone() action.Workspace {
	return &actionWorkspace{
		TaskWorkspace: w.TaskWorkspace.Clone(),
		output:        w.output,
		baseCtx:       w.baseCtx,
		recipeTaskCtx: w.recipeTaskCtx,
	}
}
