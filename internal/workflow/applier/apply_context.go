// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workspace"
)

type applyContext struct {
	*stateManager

	plan        *plan.Plan
	finder      cmap.RecipeFinder
	applyCfg    *Config
	runsWs      *workspace.RunWorkspace
	targetState *state.Accessor
	buildAction func(ctx action.ActionContext) (action.Action, error)
	operation   string
	runContext  action.RunContext
}

func NewApplyContext(cfg *Config) (*applyContext, error) {

	// plan
	pln, err := cfg.LoadPlan()
	if err != nil {
		return nil, err
	}

	// validate plan and mode
	if pln.Mode != cfg.Mode {
		err = fmt.Errorf("plan mode %q does not match with config mode %q", pln.Mode, cfg.Mode)
		switch pln.Mode {
		case plan.ModePrune:
			err = errors.Join(err, fmt.Errorf("add --prune option to apply a prune plan or check the plan file"))
		case plan.ModeDestroy:
			err = errors.Join(err, fmt.Errorf("add --destroy option to apply a destroy plan or check the plan file"))
		}
		return nil, err
	}

	// Create workspace
	runsWs, err := cfg.NewWorkspace(pln.Hash)
	if err != nil {
		return nil, err
	}

	// recipe finder
	finder, err := cfg.NewRecipeFinder(pln.Metadata.Recipes)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to load recipes in apply: %w", err))
	}

	// Load state file
	stateManager, err := cfg.LoadState(runsWs, pln)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to load state in apply: %w", err))
	}

	ctx := &applyContext{
		applyCfg:     cfg,
		plan:         pln,
		finder:       finder,
		stateManager: stateManager,
		targetState:  state.NewAccessor(pln.States.Target),
		buildAction:  action.BuildAction,
		runContext:   NewRunContext(cfg),
		runsWs:       runsWs,
	}

	return ctx, nil
}

func (c *applyContext) WithOperation(op string) *applyContext {
	c.operation = op
	return c
}

func (c *applyContext) RecipeFinder() cmap.RecipeFinder {
	return c.finder
}

func (c *applyContext) Plan() *plan.Plan {
	return c.plan
}

func (c *applyContext) TargetState() *state.Accessor {
	return c.targetState
}

func (ac *applyContext) Operation() string {
	return ac.operation
}

func (ac *applyContext) BuildAction(ctx action.ActionContext) (action.Action, error) {
	return ac.buildAction(ctx)
}

func (c *applyContext) RunContext() action.RunContext {
	return c.runContext
}

func (c *applyContext) RunWorkspace() *workspace.RunWorkspace {
	return c.runsWs
}

func (c *applyContext) FileSystem() file.FileSystem {
	return c.applyCfg.FileSystem
}

func (c *applyContext) TaskFinished(res workflow.TaskResult) error {
	return c.saveTaskResult(c.operation, res.GetTask().GetJob(), res.GetError())
}

type runContext struct {
	action.CmdFactory
	dryRun string
}

func NewRunContext(cfg *Config) *runContext {
	simCmd := cfg.DevOption.SimCmd()
	if cfg.DryRunMode != "" && cfg.DryRunMode != "off" {
		simCmd = true
	}

	var f action.CmdFactory
	if simCmd {
		f = NewSimCmdFactory()
	} else {
		f = action.NewCmdFactory()
	}

	return &runContext{
		CmdFactory: f,
		dryRun:     cfg.DryRunMode,
	}
}

func (drc *runContext) DryRun() string {
	if drc == nil {
		return ""
	}
	return drc.dryRun
}

func (drc *runContext) DryRunEnabled() bool {
	if drc == nil {
		return false
	}
	return drc.dryRun != "" && drc.dryRun != "off"
}
