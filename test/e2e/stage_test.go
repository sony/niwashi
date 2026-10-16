// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/test/e2e/helpers"
)

const (
	stageNodeCap = "integ-test/stage-node"
	stageHostJob = "host:integ-test/stage-host@1.0.0:"
)

func stageNodeJob(node string) string {
	return "node:" + stageNodeCap + "@1.0.0:" + node
}

// copyStageRecipes copies the stage recipes into a temporary directory,
// replacing @MARKER_FILE@ with the marker file path. The host recipe has no
// params, so the path is written into the recipes instead.
func copyStageRecipes(t *testing.T, markerFile string) string {
	t.Helper()

	src, _ := filepath.Abs("testdata/stage/recipe")
	dst := filepath.Join(t.TempDir(), "recipe")
	if err := os.MkdirAll(dst, 0o750); err != nil {
		t.Fatalf("failed to create the recipe directory: %v", err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("failed to read %s: %v", src, err)
	}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			t.Fatalf("failed to read %s: %v", e.Name(), err)
		}
		content := strings.ReplaceAll(string(data), "@MARKER_FILE@", markerFile)
		if err := os.WriteFile(filepath.Join(dst, e.Name()), []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write %s: %v", e.Name(), err)
		}
	}
	return dst
}

func writeStateFile(t *testing.T, workDir, filename, content string) string {
	t.Helper()
	path := filepath.Join(workDir, filename)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write %s: %v", filename, err)
	}
	return path
}

// writeMainState writes the main stage with node "cp".
func writeMainState(t *testing.T, workDir string) string {
	t.Helper()
	return writeInventoryState(t, workDir, "target.yaml", `  nodes:
    cp:
      capabilities:
      - `+stageNodeCap+`
`)
}

// writePreStageState writes a pre-stage with node "gateway". Generator names
// are closed in each stage, so the pre-stage has its own generator.
func writePreStageState(t *testing.T, workDir, stage string) string {
	t.Helper()
	return writeStateFile(t, workDir, stage+".yaml", `version: nws.state/v1
stages:
  `+stage+`:
    inventory:
      nodes:
        gateway:
          capabilities:
          - `+stageNodeCap+`
    infrastructure:
      generators:
        dummy:
          provisioner: external-instance
          params:
            instances:
              dummy1: {}
`)
}

func jobIdsOf(wf *plan.Workflow) []string {
	ids := []string{}
	if wf == nil {
		return ids
	}
	for _, j := range wf.Jobs {
		ids = append(ids, j.Id)
	}
	return ids
}

func completedTasks(s *state.State) []string {
	tasks := []string{}
	if s.Status == nil {
		return tasks
	}
	for _, ids := range s.Status.ActiveApply.CompletedTasks {
		tasks = append(tasks, ids...)
	}
	return tasks
}

func assertMarkerLines(t *testing.T, got, want []string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("marker lines = %q, want %q", got, want)
	}
}

// assertStageOrderInOutput checks that the plan output lists the stages in the given order.
func assertStageOrderInOutput(t *testing.T, stdout string, stages ...string) {
	t.Helper()
	last := -1
	for _, s := range stages {
		i := strings.Index(stdout, "'"+s+"' stage jobs:")
		if i < 0 {
			t.Errorf("plan output does not list the jobs of stage %q:\n%s", s, stdout)
			return
		}
		if i < last {
			t.Errorf("plan output lists stage %q out of order %v:\n%s", s, stages, stdout)
		}
		last = i
	}
}

func newStageCLI(t *testing.T) (*helpers.CLI, string, string, string) {
	t.Helper()
	workDir := t.TempDir()
	markerFile := filepath.Join(workDir, "marker.log")
	recipeDir := copyStageRecipes(t, markerFile)
	cli := helpers.NewCLI(testBinaryPath).
		WithWorkDir(workDir).
		WithEnv("NWS_LOG_LEVEL", "DEBUG")
	cli.RunSuccess(t, "init")
	return cli, workDir, recipeDir, markerFile
}

// TestStage_ConstructAndDestroy checks that a pre-stage runs after the host
// recipes and before the main stage, that each stage is written back to its
// own place in state.json, and that destroy runs the stages in reverse order.
func TestStage_ConstructAndDestroy(t *testing.T) {
	cli, workDir, recipeDir, markerFile := newStageCLI(t)
	mainState := writeMainState(t, workDir)
	preState := writePreStageState(t, workDir, "bootstrap")

	stdout, _ := cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", mainState, "--target", preState)
	assertStageOrderInOutput(t, stdout, state.InitStageName, "bootstrap", state.MainStageName)

	p := readPlan(t, workDir)
	graphs := p.ExecutionPlan.ConstructGraph
	// the host recipe required by both stages is planned once in _init
	if ids := jobIdsOf(graphs[state.InitStageName]); !slices.Equal(ids, []string{stageHostJob}) {
		t.Errorf("jobs of %s = %v, want [%s]", state.InitStageName, ids, stageHostJob)
	}
	if ids := jobIdsOf(graphs["bootstrap"]); !slices.Contains(ids, stageNodeJob("gateway")) || slices.Contains(ids, stageNodeJob("cp")) {
		t.Errorf("jobs of bootstrap = %v, want gateway only", ids)
	}
	if ids := jobIdsOf(graphs[state.MainStageName]); !slices.Contains(ids, stageNodeJob("cp")) || slices.Contains(ids, stageNodeJob("gateway")) {
		t.Errorf("jobs of main = %v, want cp only", ids)
	}

	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")
	assertMarkerLines(t, readMarkerLines(t, markerFile), []string{"construct host", "construct gateway", "construct cp"})

	s := readState(t, workDir)
	pre, ok := s.Stages["bootstrap"]
	if !ok || pre.Inventory == nil {
		t.Fatalf("state.json has no pre-stage bootstrap: %+v", s.Stages)
	}
	if c := pre.Inventory.Nodes["gateway"].Capabilities[stageNodeCap]; c == nil || c.Store["constructed_on"] != "constructed" {
		t.Errorf("capability of gateway in the pre-stage = %+v, want the store written by the recipe", c)
	}
	if _, ok := pre.Inventory.Nodes["cp"]; ok {
		t.Errorf("the pre-stage must not have node cp of the main stage")
	}
	if c := nodeCapability(s, "cp", stageNodeCap); c == nil || c.Store["constructed_on"] != "constructed" {
		t.Errorf("capability of cp in the main stage = %+v, want the store written by the recipe", c)
	}
	if _, ok := s.Inventory.Nodes["gateway"]; ok {
		t.Errorf("the main stage must not have node gateway of the pre-stage")
	}
	tasks := completedTasks(s)
	for _, want := range []string{
		state.InitStageName + "/" + stageHostJob,
		"bootstrap/" + stageNodeJob("gateway"),
		state.MainStageName + "/" + stageNodeJob("cp"),
	} {
		if !slices.Contains(tasks, want) {
			t.Errorf("completed tasks %v do not contain %q", tasks, want)
		}
	}

	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--destroy")
	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json", "--destroy", "--yes")
	lines := readMarkerLines(t, markerFile)
	assertMarkerLines(t, lines[3:], []string{"destruct cp", "destruct gateway"})

	s = readState(t, workDir)
	if _, ok := s.Stages["bootstrap"]; ok {
		t.Errorf("state.json still has the destroyed pre-stage: %+v", s.Stages["bootstrap"])
	}
}

// TestStage_PruneRemovesPreStage checks that removing the pre-stage file from
// the targets prunes the whole pre-stage, removes it from state.json, and
// allows a pre-stage with another name in the next plan.
func TestStage_PruneRemovesPreStage(t *testing.T) {
	cli, workDir, recipeDir, markerFile := newStageCLI(t)
	mainState := writeMainState(t, workDir)
	preState := writePreStageState(t, workDir, "bootstrap")

	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", mainState, "--target", preState)
	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")

	stdout, _ := cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", mainState, "--prune")
	helpers.AssertContains(t, stdout, stageNodeJob("gateway"))
	helpers.AssertNotContains(t, stdout, stageNodeJob("cp"))
	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json", "--prune", "--yes")

	lines := readMarkerLines(t, markerFile)
	assertMarkerLines(t, lines[3:], []string{"destruct gateway"})

	s := readState(t, workDir)
	if _, ok := s.Stages["bootstrap"]; ok {
		t.Errorf("state.json still has the pruned pre-stage: %+v", s.Stages["bootstrap"])
	}
	if nodeCapability(s, "cp", stageNodeCap) == nil {
		t.Errorf("prune of the pre-stage must keep node cp of the main stage")
	}

	// the removed pre-stage must not count as a second pre-stage
	proxyState := writePreStageState(t, workDir, "proxy")
	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", mainState, "--target", proxyState)
}
