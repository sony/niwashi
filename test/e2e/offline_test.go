// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

//go:build e2e

package e2e

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/test/e2e/helpers"
)

// assertConstructed checks that state.json has everything declared in the
// desired state. Capabilities added through requires may also be present.
func assertConstructed(t *testing.T, desired, s *state.State) {
	t.Helper()

	for name, n := range desired.Inventory.Nodes {
		got, ok := s.Inventory.Nodes[name]
		if !ok {
			t.Errorf("node %q is not in state after construct", name)
			continue
		}
		for id := range n.Capabilities {
			if got.Capabilities[id] == nil {
				t.Errorf("capability %q of node %q is not in state after construct", id, name)
			}
		}
		if n.InstanceSelector != nil && !strings.HasPrefix(got.InstanceRef, n.InstanceSelector.Generator+":") {
			t.Errorf("node %q is bound to %q, want an instance of generator %q", name, got.InstanceRef, n.InstanceSelector.Generator)
		}
	}
	for name, c := range desired.Inventory.Clusters {
		got, ok := s.Inventory.Clusters[name]
		if !ok {
			t.Errorf("cluster %q is not in state after construct", name)
			continue
		}
		for id := range c.Capabilities {
			if got.Capabilities[id] == nil {
				t.Errorf("capability %q of cluster %q is not in state after construct", id, name)
			}
		}
	}
	for name := range desired.Infrastructure.Generators {
		got, ok := s.Infrastructure.Generators[name]
		if !ok {
			t.Errorf("generator %q is not in state after construct", name)
			continue
		}
		if len(got.Instances) == 0 {
			t.Errorf("generator %q has no instances after construct", name)
		}
	}
}

// assertDestroyed checks that state.json has no nodes, clusters or generators.
func assertDestroyed(t *testing.T, s *state.State) {
	t.Helper()

	if s.Inventory != nil {
		if len(s.Inventory.Nodes) != 0 {
			t.Errorf("nodes remain in state after destroy: %v", s.Inventory.Nodes)
		}
		if len(s.Inventory.Clusters) != 0 {
			t.Errorf("clusters remain in state after destroy: %v", s.Inventory.Clusters)
		}
	}
	if s.Infrastructure != nil && len(s.Infrastructure.Generators) != 0 {
		t.Errorf("generators remain in state after destroy: %v", s.Infrastructure.Generators)
	}
}

func TestOffline(t *testing.T) {
	workDir := t.TempDir()

	recipeDir, _ := filepath.Abs("testdata/offline/recipe")
	targetPath, _ := filepath.Abs("testdata/offline/desired-state.yaml")

	desired, err := state.Load(targetPath, file.NewDefaultFileSystem())
	if err != nil {
		t.Fatalf("failed to load the desired state: %v", err)
	}

	cli := helpers.NewCLI(testBinaryPath).
		WithWorkDir(workDir).
		WithEnv("NWS_LOG_LEVEL", "DEBUG")

	// construct: plan and apply
	cli.RunSuccess(t, "init")
	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", targetPath)
	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")
	assertConstructed(t, desired, readState(t, workDir))

	// destruct: plan and apply
	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", targetPath, "--destroy")
	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json", "--destroy", "--yes")
	assertDestroyed(t, readState(t, workDir))
}
