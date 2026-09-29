// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package workflow

import (
	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/cmap"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workspace"
)

type ApplyContext interface {

	// Current state
	StateManager

	// Target state
	TargetState() *state.Accessor

	// Plan
	Plan() *plan.Plan

	// Recipe finder
	RecipeFinder() cmap.RecipeFinder

	// Current operation
	Operation() string

	// Run level workspace
	RunWorkspace() *workspace.RunWorkspace

	FileSystem() file.FileSystem

	BuildAction(ctx action.ActionContext) (action.Action, error)

	RunContext() action.RunContext
}

type StateManager interface {
	ApplyPatches(patches []patch.Patch, constraint patch.PathConstraint, validate func(*state.Accessor) error) error
	State() *state.Accessor
	SaveState() error
	IsStateDirty() bool
}
