// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package planner_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/types"
	"github.com/sony/niwashi/internal/workflow"
	_ "github.com/sony/niwashi/internal/workflow/phase/node" // registers workflow.PhaseNode into phaseData via init()
	"github.com/sony/niwashi/internal/workflow/planner"
)

type mockFinder struct {
	adapters map[string]*recipe.Recipe
}

func (f *mockFinder) FindRecipe(fqid string) (*recipe.Recipe, error) {
	return nil, nil
}

func (f *mockFinder) FindRecipeAsAdapter(fqid string) (*recipe.Recipe, error) {
	if r, ok := f.adapters[fqid]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("adapter recipe %q not found", fqid)
}

// fakeRecipeFinder resolves recipes by both their FQID (id@version, used by
// RecipeDagVertex.GetRecipe) and their bare id (used by Require.Name).
type fakeRecipeFinder struct {
	recipes map[string]*recipe.Recipe
}

func newFakeRecipeFinder(recipes ...*recipe.Recipe) *fakeRecipeFinder {
	f := &fakeRecipeFinder{recipes: map[string]*recipe.Recipe{}}
	for _, r := range recipes {
		f.recipes[r.Fqid()] = r
		f.recipes[r.Metadata.Id] = r
	}
	return f
}

func (f *fakeRecipeFinder) FindRecipe(n string) (*recipe.Recipe, error) {
	if r, ok := f.recipes[n]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("recipe not found: %q", n)
}

func (f *fakeRecipeFinder) FindRecipeAsAdapter(n string) (*recipe.Recipe, error) {
	return nil, fmt.Errorf("not supported")
}

func newFixtureRecipe(kind, id string, requires ...string) *recipe.Recipe {
	r := &recipe.Recipe{
		Kind:     kind,
		Metadata: &recipe.Metadata{Id: id, Version: "1.0.0"},
		Spec:     &recipe.Spec{},
	}
	for _, req := range requires {
		r.Spec.Requires = append(r.Spec.Requires, &recipe.Require{Name: req})
	}
	r.SetDefaults()
	return r
}

func newEmptyState() *state.State {
	return &state.State{Inventory: &state.Inventory{Nodes: map[string]*node.Node{}}}
}

func hasEdge(g *workflow.DagGraph, fromId, toId string) bool {
	for _, e := range g.Edges {
		if e.From.GetId() == fromId && e.To.GetId() == toId {
			return true
		}
	}
	return false
}

func TestRecipeDagVertex_BuildRequiresEdge_ReusesExistingVertex(t *testing.T) {
	lib := newFixtureRecipe(recipe.KindNode, "org/lib")
	app := newFixtureRecipe(recipe.KindNode, "org/app", "org/lib")
	finder := newFakeRecipeFinder(lib, app)

	ctx := planner.NewContext(recipe.OperationConstruct, finder, newEmptyState(), newEmptyState(), nil, nil)

	// Simulate BuildVertices having already added a properly-configured
	// vertex for lib (e.g. carrying real params from the diff walk, or the
	// "update" flag). The auto-expansion in BuildRequiresEdge must not
	// clobber this with a plain, param-less vertex for the same recipe.
	existingLibVtx, err := planner.NewRecipeDagVertex(workflow.PhaseNode, "node-1", lib, types.Dict{"memory": "4096"}, finder, nil)
	if err != nil {
		t.Fatalf("failed to build fixture vertex: %v", err)
	}
	ctx.AddVertex(existingLibVtx)

	appVtx, err := planner.NewRecipeDagVertex(workflow.PhaseNode, "node-1", app, nil, finder, nil)
	if err != nil {
		t.Fatalf("failed to build fixture vertex: %v", err)
	}

	if err := appVtx.BuildRequiresEdge(ctx); err != nil {
		t.Fatalf("BuildRequiresEdge() failed: %v", err)
	}

	var libVtx workflow.Vertex
	for _, v := range ctx.Graph().GetVertices() {
		if v.GetId() == existingLibVtx.GetId() {
			libVtx = v
		}
	}
	if libVtx == nil {
		t.Fatalf("expected vertex %q to still be in the graph", existingLibVtx.GetId())
	}
	if !reflect.DeepEqual(libVtx.GetData(), existingLibVtx.GetData()) {
		t.Errorf("existing vertex was overwritten: got Data=%#v, want unchanged Data=%#v", libVtx.GetData(), existingLibVtx.GetData())
	}
	if !hasEdge(ctx.Graph(), existingLibVtx.GetId(), appVtx.GetId()) {
		t.Errorf("expected an edge from %q to %q", existingLibVtx.GetId(), appVtx.GetId())
	}
}

func TestRecipeDagVertex_BuildRequiresEdge_SkipsAlreadySatisfiedDependency(t *testing.T) {
	lib := newFixtureRecipe(recipe.KindNode, "org/lib")
	app := newFixtureRecipe(recipe.KindNode, "org/app", "org/lib")
	finder := newFakeRecipeFinder(lib, app)

	baseState := &state.State{
		Inventory: &state.Inventory{
			Nodes: map[string]*node.Node{
				"node-1": {Capabilities: cap.CapabilityList{"org/lib": &cap.Capability{}}},
			},
		},
	}

	ctx := planner.NewContext(recipe.OperationConstruct, finder, newEmptyState(), baseState, nil, nil)

	appVtx, err := planner.NewRecipeDagVertex(workflow.PhaseNode, "node-1", app, nil, finder, nil)
	if err != nil {
		t.Fatalf("failed to build fixture vertex: %v", err)
	}

	if err := appVtx.BuildRequiresEdge(ctx); err != nil {
		t.Fatalf("BuildRequiresEdge() failed: %v", err)
	}

	libId := workflow.MakeId(workflow.PhaseNode, lib.Fqid(), "node-1")
	if ctx.Graph().HasVertex(libId) {
		t.Errorf("expected no vertex to be created for a dependency already satisfied in the base state")
	}
	if len(ctx.Graph().Edges) != 0 {
		t.Errorf("expected no edges when the dependency is already satisfied, got %d", len(ctx.Graph().Edges))
	}
}

func TestRecipeDagVertex_BuildRequiresEdge_AddsGenuinelyNewDependencyRecursively(t *testing.T) {
	osRecipe := newFixtureRecipe(recipe.KindNode, "org/os")
	lib := newFixtureRecipe(recipe.KindNode, "org/lib", "org/os")
	app := newFixtureRecipe(recipe.KindNode, "org/app", "org/lib")
	finder := newFakeRecipeFinder(osRecipe, lib, app)

	ctx := planner.NewContext(recipe.OperationConstruct, finder, newEmptyState(), newEmptyState(), nil, nil)

	appVtx, err := planner.NewRecipeDagVertex(workflow.PhaseNode, "node-1", app, nil, finder, nil)
	if err != nil {
		t.Fatalf("failed to build fixture vertex: %v", err)
	}

	if err := appVtx.BuildRequiresEdge(ctx); err != nil {
		t.Fatalf("BuildRequiresEdge() failed: %v", err)
	}

	libId := workflow.MakeId(workflow.PhaseNode, lib.Fqid(), "node-1")
	osId := workflow.MakeId(workflow.PhaseNode, osRecipe.Fqid(), "node-1")

	if !ctx.Graph().HasVertex(libId) {
		t.Errorf("expected a new vertex to be added for the genuinely new dependency %q", libId)
	}
	if !ctx.Graph().HasVertex(osId) {
		t.Errorf("expected the transitive dependency %q to also be processed", osId)
	}
	if !hasEdge(ctx.Graph(), libId, appVtx.GetId()) {
		t.Errorf("expected an edge from %q to %q", libId, appVtx.GetId())
	}
	if !hasEdge(ctx.Graph(), osId, libId) {
		t.Errorf("expected an edge from %q to %q", osId, libId)
	}
}

func newDummyRecipe() *recipe.Recipe {
	r := &recipe.Recipe{}
	r.SetDefaults()

	r.Spec.Tasks = []*recipe.Task{
		{
			Name:      "task1",
			Operation: "dummy1",
		},
		{
			Name:      "task2",
			Operation: "dummy2",
		},
	}

	return r
}

type dummyActionSpec struct{}

func (s *dummyActionSpec) GetEnvTpl() map[string]string {
	return nil
}
func (s *dummyActionSpec) Validate() error {
	return nil
}

func TestResolveToolAlias(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		operation string
		finder    cmap.RecipeFinder
		r         *recipe.Recipe
		wantErr   bool
	}{
		{
			name:      "no tasks",
			operation: "construct",
			finder:    &mockFinder{},
			r:         newDummyRecipe(),
			wantErr:   false,
		},
		{
			name:      "tool.run with nil spec",
			operation: "construct",
			finder:    &mockFinder{},
			r: func() *recipe.Recipe {
				r := newDummyRecipe()
				r.Spec.Tasks = append(r.Spec.Tasks, &recipe.Task{
					Name:      "task3",
					Operation: "construct",
					Action: recipe.Action{
						ActionType: recipe.ToolRunActionType,
						Spec:       nil,
					},
				})
				return r
			}(),
			wantErr: true,
		},
		{
			name:      "tool.run with invalid spec",
			operation: "construct",
			finder:    &mockFinder{},
			r: func() *recipe.Recipe {
				r := newDummyRecipe()
				r.Spec.Tasks = append(r.Spec.Tasks, &recipe.Task{
					Name:      "task3",
					Operation: "construct",
					Action: recipe.Action{
						ActionType: recipe.ToolRunActionType,
						Spec:       &dummyActionSpec{},
					},
				})
				return r
			}(),
			wantErr: true,
		},
		{
			name:      "find error in FindRecipeAsAdapter",
			operation: "construct",
			finder:    &mockFinder{},
			r: func() *recipe.Recipe {
				r := newDummyRecipe()
				r.Spec.Tasks = append(r.Spec.Tasks, &recipe.Task{
					Name:      "task3",
					Operation: "construct",
					Action: recipe.Action{
						ActionType: recipe.ToolRunActionType,
						Spec: &recipe.ToolRunSpec{
							ToolRef: "nonexistent_adapter",
						},
					},
				})
				return r
			}(),
			wantErr: true,
		},
		{
			name:      "find error in FindTaskByCommand",
			operation: "construct",
			finder: &mockFinder{
				adapters: map[string]*recipe.Recipe{
					"existent_adapter": func() *recipe.Recipe {
						r := &recipe.Recipe{
							Kind: "adapter",
						}
						r.SetDefaults()
						return r
					}(),
				},
			},
			r: func() *recipe.Recipe {
				r := newDummyRecipe()
				r.Spec.Tasks = append(r.Spec.Tasks, &recipe.Task{
					Name:      "task3",
					Operation: "construct",
					Action: recipe.Action{
						ActionType: recipe.ToolRunActionType,
						Spec: &recipe.ToolRunSpec{
							ToolRef: "existent_adapter",
							Command: "nonexistent_command",
						},
					},
				})
				return r
			}(),
			wantErr: true,
		},
		{
			name:      "succeed to resolve tool alias",
			operation: "construct",
			finder: &mockFinder{
				adapters: map[string]*recipe.Recipe{
					"existent_adapter": func() *recipe.Recipe {
						r := &recipe.Recipe{
							Kind: "adapter",
						}
						r.SetDefaults()

						r.Spec.Commands = map[string]*recipe.Command{
							"existent_command": {
								Task: "existent_task",
							},
						}
						r.Spec.Tasks = []*recipe.Task{
							{
								Name: "existent_task",
							},
						}
						return r
					}(),
				},
			},
			r: func() *recipe.Recipe {
				r := newDummyRecipe()
				r.Spec.Tasks = append(r.Spec.Tasks, &recipe.Task{
					Name:      "task3",
					Operation: "construct",
					Action: recipe.Action{
						ActionType: recipe.ToolRunActionType,
						Spec: &recipe.ToolRunSpec{
							ToolRef: "existent_adapter",
							Command: "existent_command",
						},
					},
				})
				return r
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := planner.ResolveToolAlias(tt.operation, tt.finder, tt.r)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ResolveToolAlias() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ResolveToolAlias() succeeded unexpectedly")
			}
		})
	}
}
