// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package phase_host

import (
	"fmt"

	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/runtime"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/workflow/recipe_exec"
)

type Operator struct{}

func (o *Operator) CreatePathConstraint(r recipe_exec.RecipeSpec, target string) patch.PathConstraint {

	rt := r.GetRuntime()
	if rt == nil {
		// reject all paths if runtime is not defined
		logger.Error("runtime is not defined in the recipe spec", "recipe", r.GetFqid())
		return &patch.ErrorPathConstraint{}
	}

	path := fmt.Sprintf("/runtime/%s/%s", rt.Type, rt.Name)
	return &patch.SimplePathConstraint{
		DefaultPrefix: path,
		AllowedPrefixes: []string{
			path,
		},
	}
}

func (o *Operator) FilterTask(data *recipe_exec.FilterArg) ([]*recipe_exec.FilteredTask, []*recipe_exec.SubjectId, error) {

	data.Target = "host" // replace target to "host"

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
		return nil
	}
}
