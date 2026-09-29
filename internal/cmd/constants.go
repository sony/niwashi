// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package cmd

const (
	FlagWorkDir    = "work-dir"
	DefaultWorkDir = ".niwashi"
	UsageWorkDir   = "Path to the working directory."

	FlagRecipeDir  = "recipe-dir"
	UsageRecipeDir = "Directory containing recipes. Can be specified multiple times."

	DryRunOff      = "off"
	DryRunSimulate = "simulate"
)

var (
	DefaultRecipeDir = []string{}
	DryRunLevels     = []string{DryRunOff, DryRunSimulate}
)
