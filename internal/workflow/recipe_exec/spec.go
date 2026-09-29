// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/recipe"
)

type RecipeSpec interface {
	action.RecipeSpec

	GetKind() string
	GetId() string
	GetVersion() string

	GetRuntime() *recipe.Runtime

	NewTaskSpec(src *recipe.Task) TaskSpec

	GetTasks() []*recipe.Task

	GetRequires() recipe.RequireList
}

type AdapterRecipeSpec interface {
	RecipeSpec

	GetExecutionUnit() string

	NewTaskSpecFrom(command string, condVar map[string]any) (TaskSpec, error)
}

type TaskSpec interface {
	RecipeSpec
	action.TaskSpec

	GetWhere() string
	GetStateChanges() map[string]*recipe.StateChangeOperation
}

type CallerTaskSpec interface {
	action.CallerTaskSpec
}

type recipeSpec struct {
	*metaSpec
	persistent bool
	assets     []string
	params     map[string]any
	env        map[string]string
	tasks      []*recipe.Task
	runtime    *recipe.Runtime
	requires   recipe.RequireList
}

func NewRecipeSpec(src *recipe.Recipe, op string) RecipeSpec {
	return &recipeSpec{
		metaSpec: NewMetaSpec(src),
		persistent: func() bool {
			if src.Spec.Workspace == nil || src.Spec.Workspace.Mode != recipe.ModePersistent {
				return false
			}
			return true
		}(),
		assets: src.Spec.Assets,
		params: func() map[string]any {
			if src.Spec.Defaults == nil {
				return nil
			}
			return src.Spec.Defaults.Params
		}(),
		env: func() map[string]string {
			if src.Spec.Defaults == nil {
				return nil
			}
			return src.Spec.Defaults.Environments
		}(),
		tasks:    filterTaskBy(src, op),
		runtime:  src.Spec.Runtime,
		requires: src.Spec.Requires,
	}
}

func filterTaskBy(r *recipe.Recipe, op string) []*recipe.Task {

	// no filtering
	if op == "" {
		return r.Spec.Tasks
	}

	tasks := []*recipe.Task{}
	for _, t := range r.Spec.Tasks {
		if t.Operation != op {
			continue
		}
		tasks = append(tasks, t)
	}

	sortedTasks := []*recipe.Task{}
	for len(tasks) > 0 {
		prevLen := len(sortedTasks)
		nextTasks := []*recipe.Task{}
		for _, t := range tasks {
			var canRun = true
			for _, d := range t.DependsOn {
				// check if already in sortedTasks
				if slices.ContainsFunc(sortedTasks, func(tt *recipe.Task) bool { return tt.Name == d }) {
					continue
				}
				canRun = false
				break
			}
			if canRun {
				sortedTasks = append(sortedTasks, t)
			} else {
				nextTasks = append(nextTasks, t)
			}
		}

		if len(sortedTasks) == prevLen {
			// no progress made, circular dependency?
			panic("circular dependency detected among tasks")
		}

		tasks = nextTasks
	}

	return sortedTasks
}

func (r *recipeSpec) NewTaskSpec(src *recipe.Task) TaskSpec {

	t := &taskSpec{
		RecipeSpec:   r,
		name:         src.Name,
		actionType:   src.Action.ActionType,
		actionSpec:   src.Action.Spec,
		where:        src.Where,
		stateChanges: src.StateChanges,
	}

	toolRunSpec, ok := src.Action.Spec.(*recipe.ToolRunSpec)
	if ok {
		return &callerTaskSpec{
			TaskSpec:   t,
			argv:       toolRunSpec.ArgvTemplate,
			workDir:    toolRunSpec.WorkdirTemplate,
			inheritEnv: toolRunSpec.InheritEnv,
		}
	}

	return t
}

func (r *recipeSpec) GetAssets() []string {
	return r.assets
}

func (r *recipeSpec) GetDefaultParams() map[string]any {
	return r.params
}

func (r *recipeSpec) GetDefaultEnv() map[string]string {
	return r.env
}

func (r *recipeSpec) GetTasks() []*recipe.Task {
	return r.tasks
}

func (r *recipeSpec) GetRuntime() *recipe.Runtime {
	return r.runtime
}

func (r *recipeSpec) GetRequires() recipe.RequireList {
	return r.requires
}

type adapterRecipeSpec struct {
	RecipeSpec

	executionUnit string
	adapter       *recipe.Adapter
}

func NewAdapterRecipeSpec(src *recipe.Recipe) AdapterRecipeSpec {
	if src.Kind != recipe.KindAdapter {
		panic("Internal error: NewAdapterRecipeSpec called with non-adapter recipe")
	}

	return &adapterRecipeSpec{
		RecipeSpec:    NewRecipeSpec(src, ""), // no filtering
		executionUnit: src.Spec.ExecutionUnit,
		adapter:       src.GetAdapter(),
	}
}

func (r *adapterRecipeSpec) GetExecutionUnit() string {
	return r.executionUnit
}

func (r *adapterRecipeSpec) NewTaskSpecFrom(command string, condVar map[string]any) (TaskSpec, error) {

	cmd, ok := r.adapter.Commands[command]
	if !ok {
		panic("Internal error: command not found in adapter recipe: " + command)
	}

	// task specified by command
	if cmd.Task != "" {
		t := r.adapter.FindTask(cmd.Task) // check if task exists
		if t == nil {
			return nil, fmt.Errorf("task %q not found in adapter recipe %q", cmd.Task, r.GetFqid())
		}
		return r.NewTaskSpec(t), nil
	}

	// select task by evaluating where condition for each candidate
	for _, candidate := range cmd.Candidates {
		accept := false
		if candidate.Where != "" {
			cond, err := NewCondition(candidate.Where)
			if err != nil {
				return nil, fmt.Errorf("failed to create condition for candidate task %q in adapter recipe %q: %w", candidate.Task, r.GetFqid(), err)
			}

			accept, err = cond.Eval(condVar)
			if err != nil {
				return nil, fmt.Errorf("failed to evaluate condition for candidate task %q in adapter recipe %q: %w", candidate.Task, r.GetFqid(), err)
			}
		} else {
			accept = true
		}

		if accept {
			t := r.adapter.FindTask(candidate.Task) // check if task exists
			if t == nil {
				return nil, fmt.Errorf("task %q not found in adapter recipe %q", candidate.Task, r.GetFqid())
			}
			return r.NewTaskSpec(t), nil
		}
	}

	return nil, fmt.Errorf("no task specified for command %q in adapter recipe %q for tool run action in task %q in recipe %q", command, r.GetFqid(), command, r.GetFqid())
}

type metaSpec struct {
	kind, id, version, dir string
}

func NewMetaSpec(r *recipe.Recipe) *metaSpec {
	return &metaSpec{
		id:      r.Metadata.Id,
		version: r.Metadata.Version,
		dir:     filepath.Dir(r.FilePath),
		kind:    r.Kind,
	}
}

func (m *metaSpec) GetKind() string {
	return m.kind
}

func (m *metaSpec) GetId() string {
	return m.id
}

func (m *metaSpec) GetVersion() string {
	return m.version
}

func (m *metaSpec) GetFqid() string {
	return m.id + "@" + m.version
}

func (m *metaSpec) GetDir() string {
	return m.dir
}

type taskSpec struct {
	RecipeSpec
	name         string
	actionType   string
	actionSpec   action.ActionSpec
	where        string
	stateChanges map[string]*recipe.StateChangeOperation
}

func (t *taskSpec) GetRecipe() action.RecipeSpec {
	return t.RecipeSpec
}

func (t *taskSpec) GetName() string {
	return t.name
}

func (t *taskSpec) GetActionType() string {
	return t.actionType
}

func (t *taskSpec) GetActionSpec() action.ActionSpec {
	return t.actionSpec
}

func (t *taskSpec) GetWhere() string {
	return t.where
}

func (t *taskSpec) GetStateChanges() map[string]*recipe.StateChangeOperation {
	return t.stateChanges
}

type callerTaskSpec struct {
	TaskSpec
	argv       []string
	workDir    string
	inheritEnv []string
}

func (c *callerTaskSpec) GetArgvTpl() []string {
	return c.argv
}

func (c *callerTaskSpec) GetWorkDirTpl() string {
	return c.workDir
}

func (c *callerTaskSpec) GetInheritEnv() []string {
	return c.inheritEnv
}
