// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import (
	"fmt"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/runtime"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/types"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workspace"
)

const (
	RecipeStateInitial = iota
	RecipeStateRunning
	RecipeStateDone
)

type RecipeContext interface {
	GetApplyContext() workflow.ApplyContext
	GetRecipeWs() *workspace.RecipeWorkspace
	GetJob() *workflow.Job
	GetRuntimeState() *runtime.State
	GetStateFileName() string
	GetRecipe() RecipeSpec
	GetStoreData() map[string]any
}

type TaskGenerator interface {
	workflow.TaskGenerator
	RecipeContext

	GetRunState() RunState
	GetOperation() string
}

type taskGenerator struct {
	workflow.DefaultErrorHolder
	ApplyContext      workflow.ApplyContext
	job               *workflow.Job
	recipe            RecipeSpec
	recipeWs          *workspace.RecipeWorkspace
	runtimeState      *runtime.State
	dirtyState        bool
	stateFileCoundter int
	graph             *Graph
	producers         map[string]*Producer
	stateMachine      *workflow.StateMachine
	runState          *runState
	operation         string
}

func (g *taskGenerator) GetApplyContext() workflow.ApplyContext {
	return g.ApplyContext
}

func (g *taskGenerator) GetRecipeWs() *workspace.RecipeWorkspace {
	return g.recipeWs
}

func (g *taskGenerator) GetJob() *workflow.Job {
	return g.job
}

func (g *taskGenerator) GetRunState() RunState {
	return g.runState
}

func (g *taskGenerator) GetOperation() string {
	return g.operation
}

func (g *taskGenerator) GetRuntimeState() *runtime.State {
	if g.dirtyState {
		// Save runtime state file for the action
		s := g.ApplyContext.State()
		phase := g.recipeWs.Phase
		target := g.recipeWs.Target
		runtimeState, err := createRuntimeState(s, phase, target)
		if err != nil {
			logger.Error("Error creating runtime state", "error", err)
			return nil
		}
		g.runtimeState = runtimeState
		g.dirtyState = false
		g.stateFileCoundter++
	}
	return g.runtimeState
}

func (g *taskGenerator) GetStoreData() map[string]any {
	s := g.ApplyContext.State()
	r := g.recipe

	return getStoreData(s, r, g.recipeWs.Target)
}

func getStoreData(s *state.Accessor, r RecipeSpec, target string) map[string]any {

	var store map[string]any

	switch r.GetKind() {
	case recipe.KindHost:
		store = getHostStoreData(s, r)
	case recipe.KindInfra:
		// Get data from infrastructure.generators
		gen := s.GetGenerator(target)
		if gen != nil && gen.Store != nil {
			store = gen.Store.Clone()
		}
	case recipe.KindNode:
		node := s.GetNode(target)
		if node != nil {
			cap := node.Capabilities.Find(r.GetFqid())
			if cap != nil {
				store = cap.Store.Clone()
			}
		}
	case recipe.KindCluster:
		cluster := s.GetCluster(target)
		if cluster != nil {
			cap := cluster.Capabilities.Find(r.GetFqid())
			if cap != nil {
				store = cap.Store.Clone()
			}
		}
	}

	if store == nil {
		store = make(map[string]any)
	}

	return store
}

func getHostStoreData(s *state.Accessor, r RecipeSpec) map[string]any {
	var store map[string]any

	// Get data from runtime
	rr := r.GetRuntime()
	if rr != nil {
		switch rr.Type {
		case "tool":
			a, exist := s.Runtime.Tool.Get(rr.Name, "store")
			if !exist {
				break
			}

			data, ok := a.(map[string]any)
			if !ok {
				logger.Warn("store data in tool runtime is not a map[string]any, ignoring store data")
			} else {
				d := types.Dict(data)
				store = d.Clone()
			}
		case "service":
			a, exist := s.Runtime.Service.Get(rr.Name, "store")
			if !exist {
				break
			}

			data, ok := a.(map[string]any)
			if !ok {
				logger.Warn("store data in service runtime is not a map[string]any, ignoring store data")
			} else {
				d := types.Dict(data)
				store = d.Clone()
			}
		default:
			logger.Warn("unsupported runtime type for store data", "type", rr.Type)
		}
	}
	return store
}

func (g *taskGenerator) GetStateFileName() string {
	return fmt.Sprintf("state-%d.json", g.stateFileCoundter)
}

func (g *taskGenerator) GetRecipe() RecipeSpec {
	return g.recipe
}

func NewRecipeTaskGenerator(ctx workflow.ApplyContext, job *workflow.Job) (TaskGenerator, error) {
	g, err := _newRecipeTaskGenerator(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("failed to create TaskGenerator for job %q: %s", job.Id, err)
	}
	return g, nil
}

func _newRecipeTaskGenerator(ctx workflow.ApplyContext, job *workflow.Job) (*taskGenerator, error) {
	if ctx == nil {
		return nil, fmt.Errorf("invalid task context")
	}

	fqid, phase, target := parseOptions(job)

	rawRecipe, err := ctx.RecipeFinder().FindRecipe(fqid)
	if err != nil {
		return nil, err
	}

	op := ctx.Operation()
	if update, ok := job.GetFromData("update"); ok {
		if updateBool, ok := update.(bool); ok && updateBool {
			op = recipe.OperationUpdate
		}
	}

	r := NewRecipeSpec(rawRecipe, op)

	recipeWs := workspace.NewRecipeWorkspace(
		ctx.RunWorkspace(),
		phase,
		target,
		fqid,
		rawRecipe.Spec.Workspace.IsPersistent(),
	)

	gen := &taskGenerator{
		stateMachine: workflow.NewStateMachine(RecipeStateInitial, RecipeStateDone),
		ApplyContext: ctx,
		recipe:       r,
		job:          job,
		operation:    op,
		recipeWs:     recipeWs,
		producers:    make(map[string]*Producer),
		dirtyState:   true, // mark dirty to create runtime state in the first GetRuntimeState call
	}

	gen.stateMachine.AddTransitions(map[int]int{
		RecipeStateInitial: RecipeStateRunning,
		RecipeStateRunning: RecipeStateDone,
	})

	gen.runState = &runState{g: gen}
	gen.stateMachine.AddState(RecipeStateRunning, gen.runState)

	if len(r.GetTasks()) != 0 {

		// Available tasks exist, create data for action execution

		// Create workspace directories
		err := file.MakeDirsAll(ctx.FileSystem(), recipeWs.GetWorkDirPath())
		if err != nil {
			return nil, err
		}

		graph, err := NewGraph(r.GetTasks())
		if err != nil {
			return nil, err
		}

		gen.graph = graph
	}

	return gen, nil
}

func createRuntimeState(s *state.Accessor, phase, target string) (*runtime.State, error) {
	nodes := []string{}
	clusters := []string{}
	switch phase {
	case workflow.PhaseNode:
		nodes = append(nodes, target)
	case workflow.PhaseCluster:
		c := s.Inventory.Clusters[target]
		for nodeName := range c.Nodes {
			nodes = append(nodes, nodeName)
		}
		clusters = append(clusters, target)
	default:
		// do nothing
	}
	runtimeState, err := runtime.NewStateFrom(s, nodes, clusters)
	if err != nil {
		return nil, err
	}

	return runtimeState, nil
}

func parseOptions(job *workflow.Job) (fqid string, phase string, target string) {
	options := job.GetOptions()
	fqid = options[0]
	if len(options) >= 2 {
		target = options[1]
	} else {
		target = ""
	}
	phase = job.GetPhase()

	return fqid, phase, target
}

func (g *taskGenerator) setupProducer() error {

	if g.graph == nil {
		// no tasks to do, return nil to indicate completion
		return nil
	}

	// if there is an error in the generator, it will stop generating new
	// tasks and wait until all working tasks are completed
	if g.HasError() {
		return nil
	}

	// register all available producers
	// loop until no more changes to the graph (e.g. tasks with no subjects may unblock downstream tasks)
	for {
		changed := false
		for _, v := range g.graph.FindVerticesWithoutIncomingEdges() {

			task := v.GetData().(*recipe.Task)
			if g.producers[task.Name] != nil {
				// producers for the task already exists, skip
				continue
			}

			producer, err := NewProducer(g, g.recipe.NewTaskSpec(task))
			if err != nil {
				return fmt.Errorf("failed to create TaskProducer: %s", err)
			}

			if producer == nil || workflow.CanComplete(producer) {
				// no subjects to produce tasks, remove from graph so downstream tasks can proceed
				g.graph.RemoveVertex(task.Name)
				changed = true
				continue
			}

			// save files for task
			if err := producer.StoreFilesForTask(); err != nil {
				return fmt.Errorf("failed to store files for task: %w", err)
			}

			g.producers[producer.TaskName()] = producer
		}
		if !changed {
			break
		}
	}

	return nil
}

func (g *taskGenerator) Generate() workflow.Task {
	err := g.stateMachine.Execute()
	if err != nil {
		// keep error to stop generating new tasks
		g.HoldError(err)
		return nil
	}
	return g.runState.Pop()
}

func (g *taskGenerator) generate() workflow.Task {

	// empty graph
	if g.graph == nil {
		return nil
	}

	err := g.setupProducer()
	g.HoldError(err)

	if g.HasError() {
		return nil
	}

	// generate task from producer
	for _, producer := range g.producers {
		if !producer.HasError() {
			t := producer.Produce()
			if t == nil {
				return nil
			}
			return t
		}
	}

	return nil
}

func (g *taskGenerator) OnComplete(result workflow.TaskResult) error {

	task, ok := result.GetTask().(*Task)
	if ok {
		// remove the completed task from the graph
		taskName := task.GetName()
		producer, ok := g.producers[taskName]
		if ok {

			taskErr := result.GetError()

			// invoke OnComplete for the producer to update its state
			producer.OnComplete(task, taskErr)

			var finalizeErr error
			// if the producer is completed,
			// - remove it from the producers map and the graph
			// - finalize: apply patches to the runtime state
			if workflow.CanComplete(producer) {
				delete(g.producers, taskName)
				g.graph.RemoveVertex(taskName)

				finalizeErr = producer.Finalize(task)

				// check state dirty after applying patches
				if finalizeErr == nil && g.ApplyContext.IsStateDirty() {
					// mark dirty to update runtime state in the next GetRuntimeState call
					g.dirtyState = true
				}
			}

			g.HoldError(taskErr, finalizeErr)
			return finalizeErr
		}
	}
	return nil
}

func (g *taskGenerator) HasWorkingTask() bool {
	for _, p := range g.producers {
		if p.HasWorkingTask() {
			return true
		}
	}
	return false
}

func (g *taskGenerator) IsComplete() bool {
	return g.stateMachine.IsComplete()
}

func (g *taskGenerator) canTransiteToFinal() bool {
	if g.graph == nil {
		return true
	}

	if g.HasError() {
		// no working tasks when error occurs
		return !g.HasWorkingTask()
	} else {
		return len(g.producers) == 0 && g.graph.VerticesSize() == 0
	}
}
