// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_local

import (
	"github.com/sony/niwashi/internal/action"
)

const ActionType = "exec.local"

func init() {
	builder := &Builder{}
	action.RegisterBuilder(ActionType, builder)
}

type Builder struct{}

func (b *Builder) NewSpec() action.ActionSpec {
	return &Spec{}
}

func (b *Builder) Build(ctx action.ActionContext) (action.Action, error) {
	return BuildExecLocalAction(ctx)
}
