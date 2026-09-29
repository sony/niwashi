// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

//go:build e2e

package e2e

import (
	"path/filepath"
	"testing"

	"github.com/sony/niwashi/test/e2e/helpers"
)

func TestRemote(t *testing.T) {

	workDir := t.TempDir()

	srv := helpers.StartSSHServer(t, workDir)
	srv.WriteInstancesYAML(t, workDir)

	recipeDir, _ := filepath.Abs("testdata/remote/recipe")
	targetPath, _ := filepath.Abs("testdata/remote/desired-state.yaml")

	cli := helpers.NewCLI(testBinaryPath).
		WithWorkDir(workDir).
		WithEnv("NWS_LOG_LEVEL", "DEBUG")

	// construct: plan and apply
	cli.RunSuccess(t, "init")
	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", targetPath)
	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")

	// destruct: plan and apply
	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", targetPath, "--destroy")
	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json", "--destroy")
}
