// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workspace

import (
	"github.com/sony/niwashi/internal/file"
)

var (
	logDirPath    = plist(ephemeralWorkDirPath, pvalue("logs"))
	inputDirPath  = plist(ephemeralWorkDirPath, pvalue("inputs"))
	outputDirPath = plist(ephemeralWorkDirPath, pvalue("outputs"))
)

type TaskWorkspace struct {
	*RecipeWorkspace
	Task Task
}

type Task struct {
	Name string
}

func NewTaskWorkspace(recipeWs *RecipeWorkspace, task Task) *TaskWorkspace {
	w := &TaskWorkspace{
		RecipeWorkspace: recipeWs,
		Task:            task,
	}

	return w
}

func NewSimplePathResolver(rootdir, runsid, scope, target, recipeFqid string) *TaskWorkspace {
	return &TaskWorkspace{
		RecipeWorkspace: &RecipeWorkspace{
			RunWorkspace: &RunWorkspace{
				RootWorkspace: NewRootWorkspace(rootdir),
				RunsId:        runsid,
			},
			Phase:      scope,
			Target:     target,
			RecipeFqid: recipeFqid,
		},
		Task: Task{
			Name: "",
		},
	}
}

func (r *TaskWorkspace) Find(key string) string {
	return r.RecipeWorkspace.Find(key)
}

func (r *TaskWorkspace) Join(paths ...string) string {
	return r.RecipeWorkspace.Join(paths...)
}

func (r *TaskWorkspace) Render(t string) string {
	return render(t, r)
}

func (r *TaskWorkspace) Clone() *TaskWorkspace {
	return &TaskWorkspace{
		RecipeWorkspace: r.RecipeWorkspace.Clone(),
		Task:            r.Task,
	}
}

func (r *TaskWorkspace) GetInputDirPath() string {
	return inputDirPath.ToString(r)
}

func (r *TaskWorkspace) GetOutputDirPath() string {
	return outputDirPath.ToString(r)
}

func (r *TaskWorkspace) GetLogDirPath() string {
	return logDirPath.ToString(r)
}

func (w *TaskWorkspace) GetScriptFilePath() string {
	return w.Join(
		w.GetLogDirPath(),
		file.SanitizeFilename(w.Task.Name+".action"),
	)
}
