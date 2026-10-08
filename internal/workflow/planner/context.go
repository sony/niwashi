// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package planner

import (
	"github.com/sony/niwashi/internal/cmap"
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
	state, baseState         *state.State
	stageView, baseStageView *state.StageView
	diff, updateDiff         *state.Difference
	operation                string
	skippedRecipe            []string
}

func NewContext(
	op string,
	f cmap.RecipeFinder,
	s *state.State,
	baseState *state.State,
	diff *state.Difference,
	updateDiff *state.Difference) (*context, error) {

	ctx := &context{
		RecipeFinder: f,
		operation:    op,
		diff:         diff,
		updateDiff:   updateDiff,
		state:        s,
		baseState:    baseState,
		DagGraph:     workflow.NewDagGraph(),
	}

	err := ctx.SetStageName(state.DefaultStageName)
	if err != nil {
		return nil, err
	}

	return ctx, nil
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
	return pc.updateDiff
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

func (pc *context) SetStageName(stageName string) error {

	// todo: switch diff

	sv, err := pc.state.Stage(stageName)
	if err != nil {
		return err
	}

	var baseSv *state.StageView
	if pc.baseState != nil {
		baseSv, err = pc.baseState.Stage(stageName)
		if err != nil {
			return err
		}
	}
	pc.stageView = sv
	pc.baseStageView = baseSv
	return nil
}
