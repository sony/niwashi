// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_cluster

import (
	"fmt"

	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workflow/recipe_exec"
)

func filterClusterTask(data *recipe_exec.FilterArg) ([]*recipe_exec.FilteredTask, []*recipe_exec.SubjectId, error) {

	var separate bool
	a := data.Adapter
	if data.Task.GetWhere() != "" {
		// task requires separate, but adapter doesn't support node-level execution
		if a != nil && a.Recipe.GetExecutionUnit() == "default" {
			// error
			return nil, nil, fmt.Errorf("where condition is not supported for cluster-level task with adapter that does not support node-level execution unit")
		}
		separate = true
	} else {
		if a != nil && a.Recipe.GetExecutionUnit() == "node" {
			separate = true
		} else {
			separate = false
		}
	}

	if separate {
		return separateClusterTask(data)
	} else {
		s, err := data.FilterDefault(func() map[string]any {
			return map[string]any{
				"os":   data.State.Host.Os,
				"arch": data.State.Host.Arch,
			}
		})

		var subSubjects []*recipe_exec.SubjectId
		if err == nil {
			// make cluster sub subjects from nodes in the cluster
			cluster := data.State.Inventory.Clusters[data.Target]
			for nodeName := range cluster.Nodes {
				subSubjects = append(subSubjects, &recipe_exec.SubjectId{
					Scope: workflow.PhaseNode,
					Name:  nodeName,
				})
			}
		}

		return s, subSubjects, err
	}
}

func separateClusterTask(data *recipe_exec.FilterArg) ([]*recipe_exec.FilteredTask, []*recipe_exec.SubjectId, error) {

	where := data.Task.GetWhere()
	if where == "" {
		where = "true" // pass all nodes by default
	}

	cond, err := recipe_exec.NewCondition(where)
	if err != nil {
		return nil, nil, err
	}

	subTasks := []*recipe_exec.FilteredTask{}

	// filer nodes by where condition
	c := data.State.Inventory.Clusters[data.Target]
	for nodeName := range c.Nodes {

		// runtime state already merged labels into node's
		v := data.GetNodeVar(nodeName)

		accept, err := cond.Eval(v)
		if err != nil {
			return nil, nil, err
		}

		if !accept {
			continue
		}

		// find proper adapter task with node variables
		adapterTask, err := data.FindAdapterTask(v)
		if err != nil {
			return nil, nil, err
		}

		// add node task info
		subTasks = append(subTasks, &recipe_exec.FilteredTask{
			SubjectId: recipe_exec.SubjectId{
				Scope: workflow.PhaseNode,
				Name:  nodeName,
			},
			Adapter: adapterTask,
		})
	}

	return subTasks, nil, nil
}
