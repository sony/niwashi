// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package planner_test

import (
	"fmt"
	"testing"

	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/planner"
)

func TestRunPlanners(t *testing.T) {

	fs := file.NewDefaultFileSystem()

	finder, _ := cmap.NewCapabilityMapFrom(nil, fs, "./testdata/recipes")
	s, _ := state.Load("./testdata/state.yaml", fs)

	ctx := planner.NewContext(recipe.OperationConstruct, finder, s, nil, nil, nil)
	err := planner.CreateGraph(ctx)
	if err != nil {
		t.Errorf("RunPlanners failed: %v", err)
	}
	cg := ctx.Graph()

	ctx = planner.NewContext(recipe.OperationDestruct, finder, s, nil, nil, nil)
	err = planner.CreateGraph(ctx)
	if err != nil {
		t.Errorf("RunPlanners failed: %v", err)
	}
	dg := ctx.Graph()

	fmt.Printf("Construction Graph:\n%v\n", cg)
	fmt.Printf("Destruction Graph:\n%v\n", dg)

	wf, err := workflow.BuildWorkflow(cg)
	if err != nil {
		t.Errorf("BuildWorkflow failed: %v", err)
	}
	for i, j := range wf.Jobs {
		fmt.Printf("Job(%d): %s\n", i, j.Id)
	}
}
