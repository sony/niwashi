// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_cluster

import (
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/runtime"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workflow/recipe_exec"
)

type Operator struct{}

func (o *Operator) CreatePathConstraint(r recipe_exec.RecipeSpec, target string) patch.PathConstraint {
	prefix := recipe_exec.CapabilityRoot("/inventory/clusters", r, target)
	return &patch.SimplePathConstraint{
		DefaultPrefix: prefix,
		AllowedPrefixes: []string{
			prefix + "/store", // allow only store
		},
	}
}

func (o *Operator) FilterTask(data *recipe_exec.FilterArg) ([]*recipe_exec.FilteredTask, []*recipe_exec.SubjectId, error) {
	return filterClusterTask(data)
}

func (o *Operator) GetRuntimeNode(s *runtime.State, subject *recipe_exec.SubjectId) *runtime.Node {
	return s.Host
}

func (o *Operator) GetPatchValidator() recipe_exec.Validator {
	return func(s *state.Accessor, target string) error {
		return nil
	}
}
