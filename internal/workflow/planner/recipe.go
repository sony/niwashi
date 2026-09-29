// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package planner

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/types"
	"github.com/sony/niwashi/internal/workflow"
)

type RecipeDagVertex struct {
	Id   string
	Data types.Dict
}

func (d *RecipeDagVertex) GetKind() (string, error) {
	kind, _, err := workflow.ParseId(d.GetId())
	if err != nil {
		return "", err
	}
	return kind, nil
}

func (d *RecipeDagVertex) GetId() string {
	return d.Id
}

func (d *RecipeDagVertex) GetData() any {
	return d.Data
}

func NewRecipeDagVertex(
	phase,
	target string,
	r *recipe.Recipe,
	params types.Dict,
	f cmap.RecipeFinder,
	baseVertexData map[string]any) (*RecipeDagVertex, error) {
	fqid := r.Fqid()

	data := baseVertexData
	if data == nil {
		data = make(map[string]any)
	}
	if params != nil {
		data["params"] = params
	}

	switch phase {
	case workflow.PhaseHost:
		target = ""
	case workflow.PhaseInfra, workflow.PhaseNode, workflow.PhaseCluster:
		if target == "" {
			return nil, fmt.Errorf("target is required for phase %q recipe %q", phase, fqid)
		}
	default:
		return nil, fmt.Errorf("unknown phase %q for recipe %q", phase, fqid)
	}

	return &RecipeDagVertex{
		Id:   workflow.MakeId(phase, fqid, target),
		Data: data,
	}, nil
}

func (v *RecipeDagVertex) GetRecipe(f cmap.RecipeFinder) (*recipe.Recipe, error) {
	phase, options, err := workflow.ParseId(v.GetId())
	if err != nil {
		return nil, err
	}
	switch phase {
	case workflow.PhaseHost, workflow.PhaseInfra, workflow.PhaseNode, workflow.PhaseCluster:
		return f.FindRecipe(options[0])
	default:
		return nil, fmt.Errorf("unknown vertex type: %T", v)
	}
}

func (v *RecipeDagVertex) GetTarget() string {
	phase, options, err := workflow.ParseId(v.GetId())
	if err != nil {
		return ""
	}

	switch phase {
	case workflow.PhaseHost:
		return ""
	case workflow.PhaseInfra, workflow.PhaseNode, workflow.PhaseCluster:
		return options[1]
	default:
		return ""
	}
}

func (v *RecipeDagVertex) GetRequireVertices(ctx Context, require string) ([]*RecipeDagVertex, error) {

	req, err := ctx.FindRecipe(require)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to find required recipe %q", require))
	}

	src, err := v.GetRecipe(ctx)
	if err != nil {
		return nil, err
	}
	if !workflow.CanDepends(src.Kind, req.Kind) {
		return nil, fmt.Errorf("recipe %q with scope %q cannot depend on recipe %q with scope %q", src.Fqid(), src.Kind, req.Fqid(), req.Kind)
	}

	switch req.Kind {
	case workflow.PhaseHost:
		vtx, err := NewRecipeDagVertex(workflow.PhaseHost, "", req, nil, ctx, nil)
		if err != nil {
			return nil, err
		}
		return []*RecipeDagVertex{vtx}, nil
	case workflow.PhaseNode:
		switch src.Kind {
		case workflow.PhaseNode:
			// Case1: Node recipe -> Node recipe
			// Same node target
			vtx, err := NewRecipeDagVertex(workflow.PhaseNode, v.GetTarget(), req, nil, ctx, nil)
			if err != nil {
				return nil, err
			}
			return []*RecipeDagVertex{vtx}, nil
		case workflow.PhaseCluster:
			// Case2: Cluster recipe -> Node recipe
			// when cluster recipe depends on node recipe, add for all nodes in the cluster
			t := v.GetTarget()
			cluster := ctx.State().GetCluster(t)
			if cluster == nil {
				return nil, fmt.Errorf("failed to find cluster %q in state", t)
			}

			vertices := make([]*RecipeDagVertex, 0, len(cluster.Nodes))
			for node := range cluster.Nodes {
				vtx, err := NewRecipeDagVertex(workflow.PhaseNode, node, req, nil, ctx, nil)
				if err != nil {
					return nil, err
				}

				vertices = append(vertices, vtx)
			}
			return vertices, nil
		}
	case workflow.PhaseCluster:
		vtx, err := NewRecipeDagVertex(workflow.PhaseCluster, v.GetTarget(), req, nil, ctx, nil)
		if err != nil {
			return nil, err
		}
		return []*RecipeDagVertex{vtx}, nil
	}

	return nil, fmt.Errorf("recipe %q has unknown kind %q", req.Fqid(), req.Kind)
}

func (v *RecipeDagVertex) BuildRequiresEdge(ctx Context) error {
	recipe, err := v.GetRecipe(ctx)
	if err != nil {
		return errors.Join(err, fmt.Errorf("recipe is nil for vertex %q", v.GetId()))
	}

	worklist := make([]*RecipeDagVertex, 0, len(recipe.Spec.Requires))
	worklist = append(worklist, v)
	for len(worklist) > 0 {
		// pop
		current := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]

		r, err := current.GetRecipe(ctx)
		if err != nil {
			return errors.Join(err, fmt.Errorf("recipe is nil for vertex %q", current.GetId()))
		}

		for _, req := range r.Spec.Requires {
			// Create Vertex at first
			reqVertices, err := current.GetRequireVertices(ctx, req.Name)
			if err != nil {
				return err
			}

			for _, reqVtx := range reqVertices {
				// no vertex exists in the current DagGraph
				if !ctx.HasVertex(reqVtx.GetId()) {

					// Recipe is already processed, so we don't need to handle it anymore.
					if reqVtx.isContainedIn(ctx) {
						continue
					}

					ctx.AddVertex(reqVtx)
					worklist = append(worklist, reqVtx)
				}
				ctx.AddEdge(reqVtx, current)
			}
		}
	}

	return nil
}

func (v *RecipeDagVertex) isContainedIn(ctx Context) bool {
	kind, err := v.GetKind()
	if err != nil {
		logger.Error("failed to get kind for vertex", "vertex", v.GetId(), "error", err)
		return false
	}

	switch kind {
	case workflow.PhaseNode:
		return v.isContainedInNodeCap(ctx)
	case workflow.PhaseCluster:
		return v.isContainedInClusterCap(ctx)
	default:
		// workflow.PhaseHost and workflow.PhaseInfra are not supported now
		return false
	}
}

func (v *RecipeDagVertex) isContainedInNodeCap(ctx Context) bool {
	baseState := ctx.BaseState()
	if baseState == nil {
		return false
	}

	r, err := v.GetRecipe(ctx)
	if err != nil {
		return false
	}

	node := baseState.GetNode(v.GetTarget())
	if node == nil {
		return false
	}

	_, ok := node.Capabilities[r.Metadata.Id]
	return ok
}

func (v *RecipeDagVertex) isContainedInClusterCap(ctx Context) bool {
	baseState := ctx.BaseState()
	if baseState == nil {
		return false
	}

	r, err := v.GetRecipe(ctx)
	if err != nil {
		return false
	}

	cluster := baseState.GetCluster(v.GetTarget())
	if cluster == nil {
		return false
	}

	_, ok := cluster.Capabilities[r.Metadata.Id]
	return ok
}

func ResolveToolAlias(operation string, finder cmap.RecipeFinder, r *recipe.Recipe) error {

	tasks := r.FilterTasks(func(t *recipe.Task) bool {
		return t.Operation == operation && t.Action.ActionType == recipe.ToolRunActionType
	})

	var err error

	// tool alias resolution
	for _, t := range tasks {

		if t.Action.Spec == nil {
			err = errors.Join(err, fmt.Errorf("tool.run is nil for task %q in recipe %q", t.Name, r.Fqid()))
			continue
		}

		spec, ok := t.Action.Spec.(*recipe.ToolRunSpec)
		if !ok {
			err = errors.Join(err, fmt.Errorf("invalid tool run action spec for task %q in recipe %q", t.Name, r.Fqid()))
			continue
		}

		// Validation
		ar, e := finder.FindRecipeAsAdapter(spec.ToolRef)
		if e != nil {
			err = errors.Join(err, fmt.Errorf("failed to find adapter recipe %q for task %q in recipe %q: %w", spec.ToolRef, t.Name, r.Fqid(), e))
			continue
		}

		_, err = ar.GetAdapter().FindTaskByCommand(spec.Command)
		if err != nil {
			err = errors.Join(err, fmt.Errorf("failed to find task by command %q in adapter recipe %q for task %q in recipe %q: %w", spec.Command, spec.ToolRef, t.Name, r.Fqid(), err))
			continue
		}
	}

	return err
}

func GetTaskCount(ctx Context, r *recipe.Recipe) int {
	count := 0

	for _, t := range r.Spec.Tasks {
		if t.Operation == ctx.Operation() {
			count++
		}
	}

	return count
}
