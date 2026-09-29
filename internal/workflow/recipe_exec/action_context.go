// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import (
	"fmt"
	"io"
	"os"

	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/runtime"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/types"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workspace"
)

type baseActionContext struct {
	workflow.ApplyContext // for Executors(), RunContext()
	recipeContext         RecipeContext
	task                  TaskSpec
	adapterRecipe         RecipeSpec
	params                map[string]any
	store                 map[string]any
	tplParams             *action.TemplateParams // base templateParams, in some case, this will be replaced
	taskWs                *workspace.TaskWorkspace
	runtimeState          *runtime.State
}

func NewBaseActionContext(g RecipeContext, task TaskSpec, adapterRecipe RecipeSpec) (*baseActionContext, error) {

	// Merge order:
	// - No adapter case: user params > task
	// - Adapter case: user params > caller(task) > callee(adapterTask)
	params := action.GetFinalParams(createParams(g.GetJob()), task, adapterRecipe)

	store := g.GetStoreData()

	taskWs := workspace.NewTaskWorkspace(
		g.GetRecipeWs(),
		workspace.Task{
			Name: task.GetName(),
		},
	)

	tplParams := action.NewTemplateParams(
		action.WithTaskDirs(taskWs),
		action.WithParams(params),
		action.WithStore(store),
		action.WithTarget(g.GetRecipeWs().Target),
		action.WithRuntime(g.GetRuntimeState().Runtime),
	)

	return &baseActionContext{
		ApplyContext:  g.GetApplyContext(),
		recipeContext: g,
		task:          task,
		adapterRecipe: adapterRecipe,
		params:        params,
		tplParams:     tplParams,
		store:         store,
		taskWs:        taskWs,
		runtimeState:  g.GetRuntimeState(),
	}, nil
}

func createParams(job *workflow.Job) map[string]any {
	params := map[string]any{}

	// "data" field in workflow.Job as params, which includes the followings:
	//    - params in profile by user
	//    - params in capability by user
	// These compositions are done in plan phase.
	d := job.Data
	if d != nil {
		if data, found := d.(map[string]any)["params"]; found {
			params = data.(map[string]any)
		}
	}

	return params
}

type actionContext struct {
	*baseActionContext
	taskContext    *taskContext
	subject        *Subject
	actionWs       *actionWorkspace
	logPath        string
	affectSubjects []*SubjectId
	tplParams      *action.TemplateParams
}

func (c *actionContext) Close() error {
	if c.actionWs.output != nil {
		return c.actionWs.output.Close()
	}
	return nil
}

func (c *actionContext) Subject() action.Subject {
	return c.subject
}

func (c *actionContext) Task() action.TaskContext {
	return c.taskContext
}

func (c *actionContext) Workspace() action.Workspace {
	return c.actionWs
}

func (c *actionContext) TemplateParams() *action.TemplateParams {
	return c.tplParams
}

func (b *baseActionContext) NewActionContext(ctx TaskContext, subTask *FilteredTask) *actionContext {

	subject := NewSubject(b.runtimeState, subTask)

	// create output and set to workspace
	logPath := b.taskWs.Join(
		b.taskWs.GetLogDirPath(),
		file.SanitizeFilename(b.task.GetName()+"_"+subject.GetId()+".log"),
	)
	o := newOutput(logPath, b.FileSystem())

	// create action workspace
	actionWs := NewWorkspace(b, ctx, o)

	tplParams := b.tplParams
	stores := b.newStores(subject)
	if len(stores) > 0 {
		// Clone to avoid mutating the shared base params.
		c := b.tplParams.Clone()
		c.Set(action.WithStores(stores))
		tplParams = c
	}

	taskContext := &taskContext{b, subTask.Adapter, stores}

	return &actionContext{
		baseActionContext: b,
		affectSubjects:    ctx.GetAffectSubjects(),
		taskContext:       taskContext,
		subject:           subject,
		actionWs:          actionWs,
		logPath:           logPath,
		tplParams:         tplParams,
	}
}

func (b *baseActionContext) newStores(subject *Subject) map[string]types.Dict {
	stores := map[string]types.Dict{}
	requires := b.recipeContext.GetRecipe().GetRequires()

	for _, req := range requires {
		if req.As == "" {
			continue
		}

		r, err := b.ApplyContext.RecipeFinder().FindRecipe(req.Name)
		if err != nil {
			logger.Error("Failed to find required recipe", "name", req.Name, "error", err)
			continue
		}

		rs := NewRecipeSpec(r, b.Operation())

		stores[req.As] = b.findStore(b.State(), subject, rs)
	}

	return stores
}

func (b *baseActionContext) findStore(s *state.Accessor, subject *Subject, r RecipeSpec) types.Dict {
	// host store is identified via spec.runtime, not by subject ID.
	if r.GetKind() == recipe.KindHost {
		return getStoreData(s, r, "")
	}

	// When the required recipe's kind matches the subject scope, use the subject ID directly.
	if r.GetKind() == subject.GetScope() {
		return getStoreData(s, r, subject.GetId())
	}

	// A cluster recipe may decompose into per-node tasks (via where).
	// In that case the subject ID is a node ID, so use taskWs.Target (the cluster ID) instead.
	if r.GetKind() == recipe.KindCluster && b.recipeContext.GetRecipe().GetKind() == recipe.KindCluster {
		return getStoreData(s, r, b.taskWs.Target)
	}

	// All other combinations are rejected at plan time by requires validation.
	return nil
}

type taskContext struct {
	baseCtx     *baseActionContext
	adapterTask TaskSpec
	stores      map[string]types.Dict
}

func (t *taskContext) Caller() action.CallerTaskSpec {
	if t.adapterTask != nil {
		return t.baseCtx.task.(CallerTaskSpec)
	}
	return nil
}

func (t *taskContext) Callee() action.TaskSpec {
	if t.adapterTask != nil {
		return t.adapterTask
	}
	return t.baseCtx.task
}

func (t *taskContext) Params() map[string]any {
	return t.baseCtx.params
}

func (t *taskContext) Store() map[string]any {
	return t.baseCtx.store
}

func (t *taskContext) Stores() map[string]types.Dict {
	return t.stores
}

type output struct {
	o io.WriteCloser
}

func (o *output) Stdout() io.Writer {
	return o.o
}

func (o *output) Stderr() io.Writer {
	return o.o
}

func (o *output) Close() error {
	return o.o.Close()
}

func newOutput(path string, fs file.FileSystem) action.Output {
	logOut, err := fs.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create stdout log file: %v\n", err)
		logOut = os.Stdout
	}

	return &output{
		o: logOut,
	}
}
