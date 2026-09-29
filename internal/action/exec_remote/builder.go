// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_remote

import (
	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/logger"
)

const LegacyActionType = "exec.ssh"
const ActionType = "exec.remote"

func init() {
	action.RegisterBuilder(ActionType, &Builder{})

	// remove this later
	action.RegisterBuilder(LegacyActionType, &RegacyBuilder{})
}

type Builder struct{}

func (b *Builder) NewSpec() action.ActionSpec {
	return &Spec{}
}

func (b *Builder) Build(ctx action.ActionContext) (action.Action, error) {
	return BuildRemoteExecAction(ctx)
}

type RegacyBuilder struct{}

func (b *RegacyBuilder) NewSpec() action.ActionSpec {
	logger.Warn("Action type exec.ssh is deprecated. Please use exec.remote instead.")
	return &Spec{}
}

func (b *RegacyBuilder) Build(ctx action.ActionContext) (action.Action, error) {
	logger.Warn("Action type exec.ssh is deprecated. Please use exec.remote instead.")
	return BuildRemoteExecAction(ctx)
}
