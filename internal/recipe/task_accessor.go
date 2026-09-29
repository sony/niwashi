// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import "path/filepath"

type TaskAccessor struct {
	Recipe *Recipe
	Task   *Task
}

func NewTaskAccessor(r *Recipe, t *Task) *TaskAccessor {
	return &TaskAccessor{
		Recipe: r,
		Task:   t,
	}
}

func (t *TaskAccessor) RecipeDir() string {
	if t != nil && t.Recipe != nil {
		return filepath.Dir(t.Recipe.FilePath)
	}
	return ""
}

func (t *TaskAccessor) GetAdapter() *Adapter {
	if t != nil {
		return t.Recipe.GetAdapter()
	}
	return nil
}
