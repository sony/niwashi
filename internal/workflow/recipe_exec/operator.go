// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import (
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/runtime"
	"github.com/sony/niwashi/internal/state"
)

type Validator func(s *state.Accessor, target string) error

type Operator interface {
	CreatePathConstraint(RecipeSpec, string) patch.PathConstraint
	FilterTask(data *FilterArg) ([]*FilteredTask, []*SubjectId, error)
	GetRuntimeNode(s *runtime.State, subject *SubjectId) *runtime.Node
	GetPatchValidator() Validator
}

var operators = map[string]Operator{}

func RegisterOperator(phase string, operator Operator) {
	operators[phase] = operator
}

func getOperator(phase string) Operator {
	return operators[phase]
}

func createPathConstraint(phase string, task TaskSpec, target string) patch.PathConstraint {
	return getOperator(phase).CreatePathConstraint(task, target)
}

func filterTask(phase string, data *FilterArg) ([]*FilteredTask, []*SubjectId, error) {
	return getOperator(phase).FilterTask(data)
}

func getRuntimeNode(phase string, s *runtime.State, subject *SubjectId) *runtime.Node {
	return getOperator(phase).GetRuntimeNode(s, subject)
}

func getPatchValidator(phase string) Validator {
	return getOperator(phase).GetPatchValidator()
}
