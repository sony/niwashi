// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sony/niwashi/test/e2e/helpers"
)

func generateDesiredStateFile(t *testing.T, workDir string, capabilities ...string) string {
	capExp := ""
	for _, cap := range capabilities {
		capExp += "      - " + cap + "\n"
	}

	content := `
version: nws.state/v1
inventory:
  nodes:
    node1:
      capabilities:
` + capExp + `

infrastructure:
  generators:
    dummy:
      provisioner: external-instance
      params:
        instances:
          dummy1: {}
`

	path := filepath.Join(workDir, "desired-state.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write desired-state.yaml: %v", err)
	}
	return path
}

func TestErrorRecipe(t *testing.T) {
	cases := []struct {
		name           string
		capability     string
		containMessage string
	}{
		{
			name:           "not found capability",
			capability:     "xxxx", // this pattern will be searched as alias name
			containMessage: "no recipe provides",
		},
		{
			name:           "not found capability2",
			capability:     "dummy/dummy.recipe@1.0.0", // this pattern will be searched as FQID
			containMessage: "no id found",
		},
		{
			name:           "error: call unexistent command",
			capability:     "test/node.call-unexistent-command",
			containMessage: "command \"unexistent_command\" not found",
		},
	}

	workDir := t.TempDir()

	// generate state file to use the recipe

	recipeDir, _ := filepath.Abs("testdata/error_recipe")

	cli := helpers.NewCLI(testBinaryPath).
		WithWorkDir(workDir).
		WithEnv("NWS_LOG_LEVEL", "DEBUG")

	cli.RunSuccess(t, "init")

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			targetPath := generateDesiredStateFile(t, workDir, tt.capability)

			_, stderr := cli.RunExpectError(t, "plan", "--recipe-dir", recipeDir, "--target", targetPath)
			if !strings.Contains(stderr, tt.containMessage) {
				t.Errorf("expected error message to contain %q, got %q", tt.containMessage, stderr)
			}
		})
	}
}
