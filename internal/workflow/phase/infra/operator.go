// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_infra

import (
	"fmt"

	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/runtime"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workflow/recipe_exec"
)

type Operator struct{}

func (o *Operator) CreatePathConstraint(_ recipe_exec.RecipeSpec, target string) patch.PathConstraint {
	prefix := fmt.Sprintf("/infrastructure/generators/%s", target)
	return &patch.SimplePathConstraint{
		DefaultPrefix: prefix,
		AllowedPrefixes: []string{
			prefix + "/instances",
			prefix + "/store",
		},
	}
}

func (o *Operator) FilterTask(data *recipe_exec.FilterArg) ([]*recipe_exec.FilteredTask, []*recipe_exec.SubjectId, error) {
	s, err := data.FilterDefault(func() map[string]any {
		return map[string]any{
			"os":   data.State.Host.Os,
			"arch": data.State.Host.Arch,
		}
	})
	return s, nil, err
}

func (o *Operator) GetRuntimeNode(s *runtime.State, subject *recipe_exec.SubjectId) *runtime.Node {
	return s.Host
}

func (o *Operator) GetPatchValidator() recipe_exec.Validator {
	return func(s *state.Accessor, target string) error {
		g := s.GetGenerator(target)
		return g.Validate()
	}
}
