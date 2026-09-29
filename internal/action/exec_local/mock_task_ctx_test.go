// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_local

import (
	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/types"
)

type DefaultTaskContext struct {
	caller action.CallerTaskSpec
	callee action.TaskSpec
	params map[string]any
}

func NewTaskContext(
	callee action.TaskSpec,
	caller action.CallerTaskSpec,
	params map[string]any) *DefaultTaskContext {
	return &DefaultTaskContext{
		callee: callee,
		caller: caller,
		params: params,
	}
}

func (t *DefaultTaskContext) Params() map[string]any {
	return t.params
}

func (t *DefaultTaskContext) Stores() map[string]types.Dict {
	return make(map[string]types.Dict)
}

func (t *DefaultTaskContext) Callee() action.TaskSpec {
	return t.callee
}

func (t *DefaultTaskContext) Caller() action.CallerTaskSpec {
	return t.caller
}
