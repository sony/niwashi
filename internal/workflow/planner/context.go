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
	State() *state.Accessor     // target state
	BaseState() *state.Accessor // base state
	Diff() *state.Difference
	UpdateDiff() *state.Difference
	Graph() *workflow.DagGraph
	AddSkippedRecipe(phase, target, fqid string)
}

type context struct {
	cmap.RecipeFinder
	*workflow.DagGraph
	state         *state.Accessor
	baseState     *state.Accessor
	diff          *state.Difference
	updateDiff    *state.Difference
	operation     string
	skippedRecipe []string
}

func NewContext(
	op string,
	f cmap.RecipeFinder,
	s *state.State,
	baseState *state.State,
	diff *state.Difference,
	updateDiff *state.Difference) *context {

	ctx := &context{
		RecipeFinder: f,
		operation:    op,
		diff:         diff,
		updateDiff:   updateDiff,
		state:        state.NewAccessor(s),
		baseState:    state.NewAccessor(baseState),
		DagGraph:     workflow.NewDagGraph(),
	}

	return ctx
}

func (pc *context) Operation() string {
	return pc.operation
}

func (pc *context) State() *state.Accessor {
	return pc.state
}

func (pc *context) BaseState() *state.Accessor {
	return pc.baseState
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
