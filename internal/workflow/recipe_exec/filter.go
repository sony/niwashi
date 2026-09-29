// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import (
	"fmt"

	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/runtime"
)

type FilteredTask struct {
	SubjectId
	Adapter TaskSpec
}

type SubjectId struct {
	Scope string
	Name  string
}

func FilterTask(c RecipeContext, task TaskSpec) ([]*FilteredTask, []*SubjectId, error) {

	adapterInfo, err := getAdapterInfo(c.GetApplyContext().RecipeFinder(), task)
	if err != nil {
		return nil, nil, err
	}

	filterData := &FilterArg{
		State:   c.GetRuntimeState(),
		Task:    task,
		Target:  c.GetRecipeWs().Target,
		Adapter: adapterInfo,
	}

	var subjects []*FilteredTask
	var subSubjects []*SubjectId
	subjects, subSubjects, err = filterTask(task.GetKind(), filterData)
	if err != nil {
		return nil, nil, err
	}

	return subjects, subSubjects, nil
}

type FilterArg struct {
	State   *runtime.State
	Task    TaskSpec
	Target  string
	Adapter *AdapterInfo
}

type AdapterInfo struct {
	Recipe      AdapterRecipeSpec
	CommandName string
}

func getAdapterInfo(finder cmap.RecipeFinder, task TaskSpec) (*AdapterInfo, error) {
	toolRunSpec, ok := task.GetActionSpec().(*recipe.ToolRunSpec)
	if !ok || toolRunSpec == nil {
		// no adapter
		return nil, nil
	}

	r, err := finder.FindRecipeAsAdapter(toolRunSpec.ToolRef)
	if err != nil {
		return nil, err
	}

	if r.Spec.AllowedScope != "*" && r.Spec.AllowedScope != task.GetKind() {
		return nil, fmt.Errorf("task %q with kind %q is not allowed",
			task.GetName(), task.GetKind())
	}

	return &AdapterInfo{
		Recipe:      NewAdapterRecipeSpec(r),
		CommandName: toolRunSpec.Command,
	}, nil
}

func (data *FilterArg) FilterDefault(getVar func() map[string]any) ([]*FilteredTask, error) {

	var accept bool
	var adapterTask TaskSpec
	target := data.Target
	where := data.Task.GetWhere()
	if where == "" {
		where = "true"
	}

	// evaluate where condition
	cond, err := NewCondition(where)
	if err != nil {
		return nil, err
	}

	v := getVar()

	accept, err = cond.Eval(v)
	if err != nil {
		return nil, err
	} else if !accept {
		// skip task
		return nil, nil
	}

	adapterTask, err = data.FindAdapterTask(v)
	if err != nil {
		return nil, err
	}

	return []*FilteredTask{
		{
			SubjectId: SubjectId{
				Scope: data.Task.GetKind(),
				Name:  target,
			},
			Adapter: adapterTask,
		},
	}, nil
}

func (data *FilterArg) GetNodeVar(nodeName string) map[string]any {
	nd := data.State.Inventory.Nodes[nodeName]
	return map[string]any{
		"os":     nd.Os,
		"arch":   nd.Arch,
		"labels": nd.Labels,
	}
}
func (data *FilterArg) FindAdapterTask(v map[string]any) (TaskSpec, error) {
	a := data.Adapter
	if a == nil {
		return nil, nil
	}

	return a.Recipe.NewTaskSpecFrom(a.CommandName, v)
}
