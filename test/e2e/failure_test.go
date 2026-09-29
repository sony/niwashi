// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/test/e2e/helpers"
)

// writeInventoryState writes a desired-state.yaml with the given inventory
// block (YAML indented under "inventory:") plus the minimal infrastructure
// boilerplate needed for plan/apply to succeed without provisioning anything.
func writeInventoryState(t *testing.T, workDir, filename, inventory string) string {
	t.Helper()

	content := "version: nws.state/v1\ninventory:\n" + inventory + `
infrastructure:
  generators:
    dummy:
      provisioner: external-instance
      params:
        instances:
          dummy1: {}
`
	path := filepath.Join(workDir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", filename, err)
	}
	return path
}

func nodeCapability(s *state.State, nodeName, capId string) *cap.Capability {
	if s.Inventory == nil {
		return nil
	}
	n, ok := s.Inventory.Nodes[nodeName]
	if !ok {
		return nil
	}
	return n.Capabilities[capId]
}

func clusterCapability(s *state.State, clusterName, capId string) *cap.Capability {
	if s.Inventory == nil {
		return nil
	}
	c, ok := s.Inventory.Clusters[clusterName]
	if !ok {
		return nil
	}
	return c.Capabilities[capId]
}

func newFailureCLI(t *testing.T) (*helpers.CLI, string, string) {
	t.Helper()
	workDir := t.TempDir()
	recipeDir, _ := filepath.Abs("testdata/failure/recipe")
	cli := helpers.NewCLI(testBinaryPath).
		WithWorkDir(workDir).
		WithEnv("NWS_LOG_LEVEL", "DEBUG")
	cli.RunSuccess(t, "init")
	return cli, workDir, recipeDir
}

// TestFailure_ConstructIsRetried checks that a failed construct does not leave
// the capability registered in state, so the next plan schedules it again.
func TestFailure_ConstructIsRetried(t *testing.T) {
	cli, workDir, recipeDir := newFailureCLI(t)
	capId := "integ-test/construct-fail"

	target := writeInventoryState(t, workDir, "desired-state.yaml", `  nodes:
    node1:
      capabilities:
      - `+capId+`
`)
	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", target)
	cli.RunExpectError(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")

	if c := nodeCapability(readState(t, workDir), "node1", capId); c != nil {
		t.Errorf("capability %q is still registered after a failed construct: %+v", capId, c)
	}

	stdout, _ := cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", target)
	helpers.AssertContains(t, stdout, "node:"+capId+"@1.0.0:node1")
}

// TestFailure_ParallelTasksRunFailureHooks checks the failure path when one
// task fails while other independent tasks of the same recipe are pending or
// running: the failure hooks must still run once the remaining work settles,
// and the task error must be reported exactly once.
func TestFailure_ParallelTasksRunFailureHooks(t *testing.T) {
	cli, workDir, recipeDir := newFailureCLI(t)
	capId := "integ-test/parallel-fail"

	target := writeInventoryState(t, workDir, "desired-state.yaml", `  nodes:
    node1:
      capabilities:
      - `+capId+`
`)
	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", target)
	_, stderr := cli.RunExpectError(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json", "--concurrency", "4")

	if c := nodeCapability(readState(t, workDir), "node1", capId); c != nil {
		t.Errorf("capability %q is still registered after a failed construct: %+v", capId, c)
	}

	// The final error printed by nwsctl must contain the failed task's error once.
	if n := strings.Count(stderr, "failed to execute action"); n != 1 {
		t.Errorf("task error appears %d times in the final error, want 1:\n%s", n, stderr)
	}
}

// TestFailure_LaterPhasesDoNotStart checks that once a job fails, jobs of
// later phases are not started.
func TestFailure_LaterPhasesDoNotStart(t *testing.T) {
	cli, workDir, recipeDir := newFailureCLI(t)
	markerFile := filepath.Join(workDir, "cluster-marker.log")

	target := writeInventoryState(t, workDir, "desired-state.yaml", `  nodes:
    node1:
      capabilities:
      - integ-test/construct-fail
  clusters:
    cluster1:
      nodes: [node1]
      capabilities:
      - id: integ-test/cluster-marker
        params:
          markerFile: "`+markerFile+`"
`)
	stdout, _ := cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", target)
	// Guard against a vacuous pass: the cluster job must actually be planned.
	helpers.AssertContains(t, stdout, "cluster:integ-test/cluster-marker@1.0.0:cluster1")

	cli.RunExpectError(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")

	helpers.AssertFileNotExists(t, markerFile)
}

// TestFailure_DestructKeepsState checks that a failed destruct leaves the
// capability in state, so the next destroy plan schedules it again.
func TestFailure_DestructKeepsState(t *testing.T) {
	cli, workDir, recipeDir := newFailureCLI(t)
	capId := "integ-test/destruct-fail"

	target := writeInventoryState(t, workDir, "desired-state.yaml", `  nodes:
    node1:
      capabilities:
      - `+capId+`
`)
	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", target)
	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")

	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--destroy")
	cli.RunExpectError(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json", "--destroy", "--yes")

	if c := nodeCapability(readState(t, workDir), "node1", capId); c == nil {
		t.Errorf("capability %q was removed from state although its destruct failed", capId)
	}

	stdout, _ := cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--destroy")
	helpers.AssertContains(t, stdout, "node:"+capId+"@1.0.0:node1")
}
