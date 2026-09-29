// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe_exec

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/logger"
	"github.com/sony/niwashi/internal/patch"
	"github.com/sony/niwashi/internal/types"
	"github.com/sony/niwashi/internal/workflow"
)

func CapabilityRoot(prefix string, r RecipeSpec, target string) string {
	escapedId := strings.ReplaceAll(r.GetId(), "/", "~1")
	return fmt.Sprintf("%s/%s/capabilities/%s", prefix, target, escapedId)
}

func AddCapability(
	prefix string,
	sm workflow.StateManager,
	r RecipeSpec,
	target string,
	job *workflow.Job) error {

	path := CapabilityRoot(prefix, r, target)
	params := types.Params{}
	if p, exists := job.GetFromData("params"); exists {
		if x, ok := p.(map[string]any); ok {
			params = x
		}
	}

	err := sm.ApplyPatches([]patch.Patch{
		patch.NewAddPatch(
			path,
			&cap.Capability{
				Version: r.GetVersion(),
				Params:  params,
			},
		)}, nil, nil)

	if err != nil {
		logger.Error("failed to apply set capability patches: ", err)
		return errors.Join(fmt.Errorf("failed to apply set capability patches"), err)
	}
	return nil
}

func UpdateCapability(
	prefix string,
	sm workflow.StateManager,
	r RecipeSpec,
	target string,
	job *workflow.Job) error {

	path := CapabilityRoot(prefix, r, target)
	params := map[string]any{}
	if p, exists := job.GetFromData("params"); exists {
		if x, ok := p.(map[string]any); ok {
			params = x
		}
	}

	err := sm.ApplyPatches([]patch.Patch{
		patch.NewAddPatch(
			path+"/version",
			r.GetVersion(),
		),
		patch.NewAddPatch(
			path+"/params",
			params,
		),
	}, nil, nil)

	if err != nil {
		logger.Error("failed to apply update capability patches: ", err)
		return errors.Join(fmt.Errorf("failed to apply update capability patches"), err)
	}
	return nil
}

func RemoveCapability(prefix string, sm workflow.StateManager, r RecipeSpec, target string) error {
	path := CapabilityRoot(prefix, r, target)

	err := sm.ApplyPatches([]patch.Patch{
		patch.NewRemovePatch(path),
	}, nil, nil)

	if err != nil {
		logger.Error("failed to apply remove capability patches: ", err)
		return errors.Join(fmt.Errorf("failed to apply remove capability patches"), err)
	}
	return nil
}
