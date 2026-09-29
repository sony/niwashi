// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import (
	"fmt"

	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/workflow"
)

type Graph struct {
	*workflow.DagGraph
}

func NewGraph(tasks []*recipe.Task) (*Graph, error) {

	graph, err := createDagGraph(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to create DAG graph: %v", err)
	}

	return &Graph{DagGraph: graph}, nil
}

func createDagGraph(tasks []*recipe.Task) (*workflow.DagGraph, error) {
	dag := workflow.NewDagGraph()

	// add vertex
	for _, task := range tasks {
		dag.AddVertex(&Vertex{task: task})
	}

	// add edge
	for _, v := range dag.Vertices {
		for _, dep := range v.GetData().(*recipe.Task).DependsOn {
			depVtx, ok := dag.Vertices[dep]
			if !ok {
				return nil, fmt.Errorf("task not found: depends on %s", dep)
			}
			dag.AddEdge(depVtx, v)
		}
	}

	return dag, nil
}

type Vertex struct {
	task *recipe.Task
}

func (v *Vertex) GetId() string {
	return v.task.Name
}

func (v *Vertex) GetData() any {
	return v.task
}
