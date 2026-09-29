// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package action

import (
	"github.com/sony/niwashi/internal/types"
)

type RecipeMetaSpec interface {
	GetFqid() string
	GetDir() string
}

type RecipeSpec interface {
	RecipeMetaSpec
	GetAssets() []string
	GetDefaultParams() map[string]any
	GetDefaultEnv() map[string]string
}

type TaskSpec interface {
	RecipeSpec
	GetName() string
	GetRecipe() RecipeSpec
	GetActionType() string
	GetActionSpec() ActionSpec
}

type CallerTaskSpec interface {
	TaskSpec
	GetArgvTpl() []string
	GetWorkDirTpl() string
	GetInheritEnv() []string
}

// Priority: userParams > recipes[0] > recipes[1] ...
func GetFinalParams(userParams map[string]any, recipes ...RecipeSpec) map[string]any {

	params := userParams
	for _, r := range recipes {
		if r != nil {
			params, _ = types.Merge(r.GetDefaultParams(), params)
		}
	}

	return params
}
