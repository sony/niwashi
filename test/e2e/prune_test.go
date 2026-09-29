// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

//go:build e2e

package e2e

import (
	"path/filepath"
	"testing"

	"github.com/sony/niwashi/test/e2e/helpers"
)

// TestPrune_RemovesCapabilities checks that pruning capabilities while keeping
// their node and cluster removes them from state, and that a second prune has
// nothing left to do. A full --destroy cannot catch a capability that is left
// behind, because it removes the whole node/cluster entry.
func TestPrune_RemovesCapabilities(t *testing.T) {
	workDir := t.TempDir()
	recipeDir, _ := filepath.Abs("testdata/prune/recipe")

	cli := helpers.NewCLI(testBinaryPath).
		WithWorkDir(workDir).
		WithEnv("NWS_LOG_LEVEL", "DEBUG")

	cli.RunSuccess(t, "init")

	withCaps := writeInventoryState(t, workDir, "desired-state.yaml", `  nodes:
    node1:
      capabilities:
      - integ-test/prune-node
      - integ-test/prune-node-no-destruct
  clusters:
    cluster1:
      nodes: [node1]
      capabilities:
      - integ-test/prune-cluster
      - integ-test/prune-cluster-no-destruct
`)
	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", withCaps)
	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")

	withoutCaps := writeInventoryState(t, workDir, "desired-state-pruned.yaml", `  nodes:
    node1:
      capabilities: []
  clusters:
    cluster1:
      nodes: [node1]
      capabilities: []
`)
	stdout, _ := cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", withoutCaps, "--prune")
	helpers.AssertContains(t, stdout, "node:integ-test/prune-node@1.0.0:node1")
	helpers.AssertContains(t, stdout, "cluster:integ-test/prune-cluster@1.0.0:cluster1")

	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json", "--prune", "--yes")

	s := readState(t, workDir)
	if _, ok := s.Inventory.Nodes["node1"]; !ok {
		t.Fatalf("node1 was removed from state; prune should only remove its capability")
	}
	if _, ok := s.Inventory.Clusters["cluster1"]; !ok {
		t.Fatalf("cluster1 was removed from state; prune should only remove its capability")
	}
	// Recipes without destruct tasks must be removed from state as well.
	for _, id := range []string{"integ-test/prune-node", "integ-test/prune-node-no-destruct"} {
		if c := nodeCapability(s, "node1", id); c != nil {
			t.Errorf("node capability %q is still in state after prune: %+v", id, c)
		}
	}
	for _, id := range []string{"integ-test/prune-cluster", "integ-test/prune-cluster-no-destruct"} {
		if c := clusterCapability(s, "cluster1", id); c != nil {
			t.Errorf("cluster capability %q is still in state after prune: %+v", id, c)
		}
	}

	stdout, _ = cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", withoutCaps, "--prune")
	helpers.AssertNotContains(t, stdout, "node:integ-test/prune-node")
	helpers.AssertNotContains(t, stdout, "cluster:integ-test/prune-cluster")
}
