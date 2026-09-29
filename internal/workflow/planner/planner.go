// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package planner

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/profile"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workflow"
)

type Planner struct {
	finder          *cmap.NameMapper
	profile         *profile.Profile
	initial, target *state.State
	mode            plan.Mode
	filesystem      file.FileSystem
}

func NewPlanner(cfg *Config) (*Planner, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}

	finder, err := cfg.LoadRecipeFinder()
	if err != nil {
		return nil, err
	}

	prof, err := cfg.LoadProfile()
	if err != nil {
		return nil, err
	}

	init, err := cfg.LoadInitialState()
	if err != nil {
		return nil, err
	}

	target, err := cfg.LoadTargetState()
	if err != nil {
		return nil, err
	}

	// setup name mapper with binding and alias of profile to create final map data

	return &Planner{
		finder: cmap.NewNameMapper(
			prof.CapabilityBinding,
			prof.ToolAlias,
			finder),
		profile:    prof,
		initial:    init,
		target:     target,
		mode:       cfg.GetMode(),
		filesystem: cfg.fs,
	}, nil
}

func (p *Planner) MakePlan() (*plan.Plan, error) {

	// Resolve target params/env
	err := p.resolveTargetState()
	if err != nil {
		return nil, err
	}

	planData := &plan.Plan{
		Version: plan.Version,
		Mode:    p.mode,
		States: plan.States{
			Initial: plan.InitState{Hash: p.initial.Hash},
			Target:  p.target,
		},
		Runtime: plan.Runtime{
			Bindings:  p.finder.CapBounds,
			ToolAlias: p.finder.ToolBounds,
			Params:    p.profile.Params,
		},
		Metadata: plan.NewMetadata(),
	}

	var diff *state.Difference
	var updateDiff *state.Difference
	var operation string
	var targetState *state.State
	var baseState *state.State
	swapOrder := false
	switch p.mode {
	case plan.ModeDefault:
		operation = recipe.OperationConstruct

		// compute construct diff
		diff = state.ComputeConstructDiff(p.initial, p.target)
		planData.Diff.Construct = diff

		// compute update diff
		updateDiff, err = state.ComputeUpdateDiff(p.initial, p.target)
		if err != nil {
			return nil, err
		}
		planData.Diff.Update = updateDiff

		targetState = p.target
		baseState = p.initial

	case plan.ModeDestroy, plan.ModePrune:
		operation = recipe.OperationDestruct
		diff = state.ComputeDestructDiff(p.initial, p.target)
		planData.Diff.Destruct = diff
		targetState = p.initial
		baseState = nil
		swapOrder = true
	default:
		return nil, fmt.Errorf("invalid plan mode: %s", p.mode)
	}

	logger.Debug("Building plan", "operation", operation)
	ctx := NewContext(operation, p.finder, targetState, baseState, diff, updateDiff)
	err = CreateGraph(ctx)
	if err != nil {
		return nil, err
	}

	cg := ctx.DagGraph
	if swapOrder {
		cg.SwapOrder()
	}
	wf, err := workflow.BuildWorkflow(cg)
	if err != nil {
		return nil, err
	}

	switch operation {
	case recipe.OperationConstruct:
		planData.ExecutionPlan.ConstructGraph = wf
	case recipe.OperationDestruct:
		planData.ExecutionPlan.DestructGraph = wf
	}
	planData.PendingUpdates = ctx.SkippedRecipe()

	// compute hashes of referenced recipe.
	if err = computeRecipeHash(p.finder, planData.Metadata, p.filesystem); err != nil {
		return nil, err
	}

	return planData, nil
}

func (p *Planner) resolveTargetState() error {

	// Apply templates before normalization
	if err := p.target.ApplyTemplates(); err != nil {
		logger.Error("failed to apply templates in target state:", "error", err)
		return err
	}

	// Resolve params before normalization without binding/alias
	// only direct matching is applied here
	if err := p.target.ResolveParams(p.profile); err != nil {
		logger.Error("failed to resolve params in target state:", "error", err)
		return err
	}

	if err := p.target.Normalize(p.finder); err != nil {
		logger.Error("failed to normalize target state:", "error", err)
		return err
	}

	return nil
}

type contextWrapper struct {
	Context
	diffs      []*state.Difference
	operations []string
}

func newContextWrapper(ctx Context) *contextWrapper {
	diffs := []*state.Difference{}
	operations := []string{}

	if ctx.Diff() != nil {
		diffs = append(diffs, ctx.Diff())
		operations = append(operations, ctx.Operation())
	}
	if ctx.UpdateDiff() != nil {
		diffs = append(diffs, ctx.UpdateDiff())
		operations = append(operations, recipe.OperationUpdate)
	}

	return &contextWrapper{
		Context:    ctx,
		diffs:      diffs,
		operations: operations}
}

// overwrite Diff method to return the update diff instead of the original diff
func (cw *contextWrapper) Diff() *state.Difference {
	if len(cw.diffs) == 0 {
		return nil
	}
	return cw.diffs[0]
}

func (cw *contextWrapper) Consumable() bool {
	return len(cw.diffs) > 0
}

func (cw *contextWrapper) Consume() {
	if len(cw.diffs) == 0 {
		return
	}
	cw.diffs = cw.diffs[1:]
	cw.operations = cw.operations[1:]
}

func (cw *contextWrapper) Operation() string {
	if len(cw.operations) == 0 {
		return cw.Context.Operation()
	}
	return cw.operations[0]
}

func CreateGraph(ctx Context) error {

	cw := newContextWrapper(ctx)
	if !cw.Consumable() {
		return nil
	}

	phasePlanners := GetPhasePlanners()
	phases := workflow.GetOrderedPhase()

	for cw.Consumable() {
		// Build vertices
		for _, phase := range phases {
			if planner := phasePlanners[phase]; planner != nil {
				if err := planner.BuildVertices(cw); err != nil {
					return err
				}
			}
		}
		cw.Consume()
	}

	// Resolve dependencies and build edges
	vertices := cw.GetVertices()

	omit := true
	// keep the order of phases
	for _, phase := range phases {
		for _, v := range vertices {
			vPhase, _, _ := workflow.ParseId(v.GetId())
			if vPhase != phase {
				// skip vertices of other phases
				continue
			}
			if planner := phasePlanners[phase]; planner != nil {
				if err := planner.BuildEdges(cw, v); err != nil {
					return err
				}

				// If the vertex cannot be omitted, the whole plan cannot be omitted
				if !planner.CanOmit() {
					logger.Debug("Phase cannot be omitted:", "phase", phase)
					omit = false
				}
			}
		}
	}

	if omit {
		// Reset graph to indicate that no changes are needed
		logger.Debug("No changes detected, skipping plan generation")
		cw.Reset()
		return nil
	}

	return nil
}

func computeRecipeHash(finder *cmap.NameMapper, metadata *plan.Metadata, fs file.FileSystem) error {
	// Reload referenced recipes to compute their hashes.
	// And update the metadata with the computed hashes.

	for fqid, r := range finder.ReferenceRecipes {
		hash := r.Hash
		if hash == "" {
			basedir := filepath.Dir(r.FilePath)
			f, err := fs.Open(r.FilePath)
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()

			rWithHash, err := recipe.LoadWithHash(
				f,
				func(path string) (io.ReadCloser, error) {
					return fs.Open(filepath.Join(basedir, path))
				},
				r.FilePath)
			if err != nil {
				return err
			}
			hash = rWithHash.Hash
		}

		metadata.Recipes = append(metadata.Recipes, plan.RecipeFingerprint{
			Fqid: fqid,
			Hash: hash,
		})
	}

	return nil
}
