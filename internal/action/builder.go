// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package action

import (
	"fmt"

	"github.com/sony/niwashi/internal/logger"
)

type Builder interface {
	NewSpec() ActionSpec
	Build(ctx ActionContext) (Action, error)
}

var builders map[string]Builder = make(map[string]Builder)

// RegisterBuilder registers a builder for the given action type.
// The builder will be called to create an Action instance when executing
// a task with the corresponding action type.
// This interface is typically called in the init() function of each action
// implementation package to register the builder for that action type.
// For example, in internal/action/exec_local/exec_local.go, the init()
// function calls RegisterBuilder to register the builder for the "exec_local"
// action type.
//
// Example usage in an action implementation package:
//
//	func init() {
//	    action.RegisterBuilder(
//	        string(recipe.ActionTypeExecLocal), &exec_local.Builder{})
//	}
//
// Note:
//   - The implementation package must be imported in internal/cmd/register.go;
//     otherwise, init() will not be called, the action builder will not be
//     registered, and an "unknown action type" error will occur at runtime.
//   - The action type string should match the ActionType defined in the recipe
//     package for that action.
//   - The builder should create and return an instance of the Action
//     implementation for that action type, using the provided dependencies.
//   - If an action type is not registered, attempting to create an Action with
//     that type will result in an error.
func RegisterBuilder(actionType string, builder Builder) {
	logger.Debug("registering action builder", "actionType", actionType)
	builders[actionType] = builder
}

func GetBuilder(actionType string) Builder {
	return builders[actionType]
}

func NewSpec(actionType string) (ActionSpec, error) {
	builder, ok := builders[actionType]
	if !ok {
		return nil, fmt.Errorf("unknown action type: %q", actionType)
	}

	return builder.NewSpec(), nil
}

func BuildAction(ctx ActionContext) (Action, error) {
	callee := ctx.Task().Callee()
	actionType := callee.GetActionType()
	builder, ok := builders[actionType]
	if !ok {
		return nil, fmt.Errorf("unknown action type: %q", actionType)
	}
	return builder.Build(ctx)
}
