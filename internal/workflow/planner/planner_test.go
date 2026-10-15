// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package planner_test

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/planner"
	"github.com/sony/niwashi/internal/workspace"

	// register the action used by the system recipes
	_ "github.com/sony/niwashi/internal/action/exec_local"
	// register all phases used by CreateGraph
	_ "github.com/sony/niwashi/internal/workflow/phase/barrier"
	_ "github.com/sony/niwashi/internal/workflow/phase/cluster"
	_ "github.com/sony/niwashi/internal/workflow/phase/cluster_sync"
	_ "github.com/sony/niwashi/internal/workflow/phase/host"
	_ "github.com/sony/niwashi/internal/workflow/phase/infra"
	_ "github.com/sony/niwashi/internal/workflow/phase/node_sync"
)

func vertexIds(g *workflow.DagGraph) []string {
	ids := []string{}
	for _, v := range g.GetVertices() {
		ids = append(ids, v.GetId())
	}
	return ids
}

func TestRunPlanners(t *testing.T) {
	recipe.EnableNopLoader()
	fs := file.NewDefaultFileSystem()

	finder, err := cmap.NewCapabilityMapFrom(nil, fs, "./testdata/recipes")
	if err != nil {
		t.Fatalf("failed to load recipes: %v", err)
	}
	s, err := state.Load("./testdata/state.yaml", fs)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	if err := s.Normalize(finder); err != nil {
		t.Fatalf("failed to normalize state: %v", err)
	}

	// vertices built from the diff of the main stage
	wantVertices := []string{
		"infra:test/provisioner@1.0.0:vm1",
		"node:test/node-capability@1.0.0:node1",
		"cluster:test/cluster-capability@1.0.0:cluster1",
	}

	tests := []struct {
		name string
		op   string
		diff *state.Diff
	}{
		{
			name: "construct",
			op:   recipe.OperationConstruct,
			diff: &state.Diff{Construct: state.ComputeConstructDiff(nil, &s.Stage)},
		},
		{
			name: "destruct",
			op:   recipe.OperationDestruct,
			diff: &state.Diff{Destruct: state.ComputeDestructDiff(&s.Stage, nil)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffMap := map[string]*state.Diff{state.MainStageName: tt.diff}
			ctx, err := planner.NewContext(tt.op, finder).WithStage(state.MainStageName, s, nil, diffMap)
			if err != nil {
				t.Fatalf("WithStage() error = %v", err)
			}
			if err := planner.CreateGraph(ctx); err != nil {
				t.Fatalf("CreateGraph() error = %v", err)
			}

			ids := vertexIds(ctx.Graph())
			for _, want := range wantVertices {
				if !slices.Contains(ids, want) {
					t.Errorf("vertex %q not found in graph: %v", want, ids)
				}
			}

			wf, err := workflow.BuildWorkflow(ctx.Graph())
			if err != nil {
				t.Fatalf("BuildWorkflow() error = %v", err)
			}
			if len(wf.Jobs) == 0 {
				t.Errorf("BuildWorkflow() returned no jobs")
			}
		})
	}
}

func jobIds(wf *plan.Workflow) []string {
	ids := []string{}
	if wf == nil {
		return ids
	}
	for _, j := range wf.Jobs {
		ids = append(ids, j.Id)
	}
	return ids
}

// newWorkDir creates a work directory whose state.json holds the given state.
func newWorkDir(t *testing.T, s *state.State) string {
	t.Helper()
	dir := t.TempDir()
	p := workspace.NewRunWs(dir).GetStateFilePath()
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		t.Fatalf("failed to create the state directory: %v", err)
	}
	if err := s.WriteToFile(p, file.NewDefaultFileSystem()); err != nil {
		t.Fatalf("failed to write the initial state: %v", err)
	}
	return dir
}

// newAppliedState returns the normalized state of the target files,
// which is used as the initial state of the destroy mode.
func newAppliedState(t *testing.T, paths ...string) *state.State {
	t.Helper()
	cfg := planner.NewConfig(
		planner.WithRecipeDirs("./testdata/recipes"),
		planner.WithTargetFiles(paths...),
	)
	finder, err := cfg.LoadRecipeFinder()
	if err != nil {
		t.Fatalf("failed to load recipes: %v", err)
	}
	s, err := cfg.LoadTargetState()
	if err != nil {
		t.Fatalf("failed to load target state: %v", err)
	}
	if err := s.Normalize(finder); err != nil {
		t.Fatalf("failed to normalize state: %v", err)
	}
	return s
}

// assertStageJobs checks that the stage has the wanted jobs but not the host job.
func assertStageJobs(t *testing.T, stage string, wf *plan.Workflow, want []string, hostJob string) {
	t.Helper()
	ids := jobIds(wf)
	for _, w := range want {
		if !slices.Contains(ids, w) {
			t.Errorf("job %q not found in stage %q: %v", w, stage, ids)
		}
	}
	if slices.Contains(ids, hostJob) {
		t.Errorf("stage %q must not have the host job: %v", stage, ids)
	}
}

func TestPlanner_MakePlan_InitStage(t *testing.T) {
	recipe.EnableNopLoader()

	// test/node-capability requires test/host-tool
	hostJob := workflow.MakeId(workflow.PhaseHost, "test/host-tool@1.0.0", "")
	nodeJob := func(node string) string {
		return workflow.MakeId(workflow.PhaseNode, "test/node-capability@1.0.0", node)
	}

	tests := []struct {
		name       string
		initial    *state.State
		opts       []func(*planner.Config)
		wantStages []string
		// jobs expected in each stage other than _init
		wantJobs map[string][]string
	}{
		{
			name:    "construct main stage only",
			initial: state.NewState(),
			opts: []func(*planner.Config){
				planner.WithTargetFiles("./testdata/state.yaml"),
			},
			wantStages: []string{state.InitStageName, state.MainStageName},
			wantJobs: map[string][]string{
				state.MainStageName: {nodeJob("node1")},
			},
		},
		{
			name:    "construct with a pre-stage",
			initial: state.NewState(),
			opts: []func(*planner.Config){
				planner.WithTargetFiles("./testdata/state.yaml", "./testdata/state-stage.yaml"),
			},
			wantStages: []string{state.InitStageName, "bootstrap", state.MainStageName},
			wantJobs: map[string][]string{
				"bootstrap":         {nodeJob("gateway")},
				state.MainStageName: {nodeJob("node1")},
			},
		},
		{
			name:    "destroy with a pre-stage",
			initial: newAppliedState(t, "./testdata/state.yaml", "./testdata/state-stage.yaml"),
			opts: []func(*planner.Config){
				planner.WithDestroy(),
			},
			wantStages: []string{state.InitStageName, "bootstrap", state.MainStageName},
			wantJobs: map[string][]string{
				"bootstrap":         {nodeJob("gateway")},
				state.MainStageName: {nodeJob("node1")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := append([]func(*planner.Config){
				planner.WithWorkDir(newWorkDir(t, tt.initial)),
				planner.WithRecipeDirs("./testdata/recipes"),
			}, tt.opts...)
			p, err := planner.NewPlanner(planner.NewConfig(opts...))
			if err != nil {
				t.Fatalf("NewPlanner() error = %v", err)
			}

			got, err := p.MakePlan()
			if err != nil {
				t.Fatalf("MakePlan() error = %v", err)
			}

			graphs := got.ExecutionPlan.ConstructGraph
			if got.Mode != plan.ModeDefault {
				graphs = got.ExecutionPlan.DestructGraph
			}
			stages := slices.Sorted(maps.Keys(graphs))
			if !slices.Equal(stages, slices.Sorted(slices.Values(tt.wantStages))) {
				t.Fatalf("stages = %v, want %v", stages, tt.wantStages)
			}

			// the host recipe required by several stages is planned only once in _init
			if ids := jobIds(graphs[state.InitStageName]); !slices.Equal(ids, []string{hostJob}) {
				t.Errorf("jobs of %s = %v, want [%s]", state.InitStageName, ids, hostJob)
			}

			for stage, want := range tt.wantJobs {
				assertStageJobs(t, stage, graphs[stage], want, hostJob)
			}
		})
	}
}
