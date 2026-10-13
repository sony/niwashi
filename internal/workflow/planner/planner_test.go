// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package planner_test

import (
	"slices"
	"testing"

	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/planner"

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
