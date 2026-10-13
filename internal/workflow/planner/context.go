// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package planner

import (
	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workflow"
)

type Context interface {
	cmap.RecipeFinder
	workflow.Graph
	Operation() string
	StageView() *state.StageView     // target state
	BaseStageView() *state.StageView // base state
	Diff() *state.Difference
	UpdateDiff() *state.Difference
	Graph() *workflow.DagGraph
	AddSkippedRecipe(phase, target, fqid string)
}

type context struct {
	cmap.RecipeFinder
	*workflow.DagGraph
	operation string
	initGraph *workflow.DagGraph

	stageView, baseStageView *state.StageView
	diff, diffUpdate         *state.Difference
	skippedRecipe            []string
}

func NewContext(op string, f cmap.RecipeFinder) *context {
	return &context{
		RecipeFinder: f,
		operation:    op,
		initGraph:    workflow.NewDagGraph(),
	}
}

func (pc *context) WithStage(stageName string, s, baseState *state.State, diffMap map[string]*state.Diff) (*context, error) {
	newCtx := &context{
		RecipeFinder: pc.RecipeFinder,
		operation:    pc.operation,
		initGraph:    pc.initGraph,
	}

	// switch stage view
	sv, err := s.StageView(stageName)
	if err != nil {
		return nil, err
	}
	var baseSv *state.StageView
	if baseState != nil {
		if _, exist := baseState.FindStage(stageName); exist {
			baseSv, err = baseState.StageView(stageName)
			if err != nil {
				return nil, err
			}
		}
	}
	newCtx.stageView = sv
	newCtx.baseStageView = baseSv

	// switch diff
	diffSpec, exists := diffMap[stageName]
	if exists && diffSpec != nil {
		if newCtx.operation == recipe.OperationDestruct {
			newCtx.diff = diffSpec.Destruct
			newCtx.diffUpdate = nil
		} else {
			newCtx.diff = diffSpec.Construct
			newCtx.diffUpdate = diffSpec.Update
		}
	}

	// switch graph
	g := workflow.NewDagGraph()
	newCtx.DagGraph = g

	return newCtx, nil
}

func (pc *context) Operation() string {
	return pc.operation
}

func (pc *context) StageView() *state.StageView {
	return pc.stageView
}

func (pc *context) BaseStageView() *state.StageView {
	return pc.baseStageView
}

func (pc *context) Diff() *state.Difference {
	return pc.diff
}

func (pc *context) UpdateDiff() *state.Difference {
	return pc.diffUpdate
}

func (pc *context) Graph() *workflow.DagGraph {
	return pc.DagGraph
}

func (pc *context) AddSkippedRecipe(phase, target, fqid string) {
	pc.skippedRecipe = append(pc.skippedRecipe, workflow.MakeId(phase, fqid, target))
}

func (pc *context) SkippedRecipe() []string {
	return pc.skippedRecipe
}
