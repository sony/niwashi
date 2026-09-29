// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

//go:build e2e

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/test/e2e/helpers"
)

// writeUpdateDesiredState writes a desired-state.yaml with a single node1
// carrying the given capability block (already indented YAML for the
// "capabilities:" list), plus the minimal infrastructure boilerplate needed
// for plan/apply to succeed without provisioning anything real.
func writeUpdateDesiredState(t *testing.T, workDir, filename, capabilityBlock string) string {
	t.Helper()

	content := `
version: nws.state/v1
inventory:
  nodes:
    node1:
      capabilities:
` + capabilityBlock + `

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

// readState loads and parses <workDir>/state.json using the same types the
// production code uses, so the test stays in sync with the real schema.
func readState(t *testing.T, workDir string) *state.State {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(workDir, ".niwashi", "state", "state.json"))
	if err != nil {
		t.Fatalf("failed to read state.json: %v", err)
	}

	var s state.State
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("failed to parse state.json: %v", err)
	}
	return &s
}

// readMarkerLines reads the lines appended to a marker file by the fixture
// recipes' construct/update tasks. It exists outside niwashi's own state, so
// it lets the test detect whether a task actually executed, independent of
// whatever value it wrote to the store (an idempotent task that writes the
// same store value on every run would otherwise look identical to one that
// never ran again).
func readMarkerLines(t *testing.T, path string) []string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("failed to read marker file %s: %v", path, err)
	}

	trimmed := strings.TrimRight(string(data), "\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

// TestUpdate_ParamsChange covers Proposal 027's core scenario: a capability's
// params change (same recipe version) between two applies, with an
// `operation: update` task defined, so the update task runs instead of a
// full destruct+construct. It also checks that re-applying unchanged params
// is a true no-op (no false-positive "changed" detection).
func TestUpdate_ParamsChange(t *testing.T) {
	workDir := t.TempDir()
	recipeDir, _ := filepath.Abs("testdata/update/recipe")
	markerFile := filepath.Join(workDir, "marker.log")

	cli := helpers.NewCLI(testBinaryPath).
		WithWorkDir(workDir).
		WithEnv("NWS_LOG_LEVEL", "DEBUG")

	cli.RunSuccess(t, "init")

	capBlock := `      - id: integ-test/update-cap
        params:
          memory: "1024"
          markerFile: "` + markerFile + `"
`
	targetV1 := writeUpdateDesiredState(t, workDir, "desired-state.yaml", capBlock)

	// Step 1: initial construct.
	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", targetV1)
	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")

	if lines := readMarkerLines(t, markerFile); len(lines) != 1 || lines[0] != "construct memory=1024" {
		t.Fatalf("after initial construct, marker lines = %v, want exactly [%q]", lines, "construct memory=1024")
	}

	s := readState(t, workDir)
	cap := s.Inventory.Nodes["node1"].Capabilities["integ-test/update-cap"]
	if cap == nil {
		t.Fatalf("capability not found in state.json after construct")
	}
	if got := cap.Params["memory"]; got != "1024" {
		t.Errorf("state.json params.memory = %v, want %q", got, "1024")
	}
	if got, _ := cap.Store["last_operation"].(string); got != "construct" {
		t.Errorf("state.json store.last_operation = %q, want %q", got, "construct")
	}

	// Step 2: re-apply the SAME target. Nothing changed, so neither construct
	// nor update should run again.
	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", targetV1)
	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")

	if lines := readMarkerLines(t, markerFile); len(lines) != 1 {
		t.Fatalf("after no-op re-apply, marker lines = %v, want no new lines", lines)
	}

	// Step 3: change params (same recipe version) and re-apply.
	capBlockV2 := `      - id: integ-test/update-cap
        params:
          memory: "2048"
          markerFile: "` + markerFile + `"
`
	targetV2 := writeUpdateDesiredState(t, workDir, "desired-state-updated.yaml", capBlockV2)

	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", targetV2)
	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")

	lines := readMarkerLines(t, markerFile)
	if len(lines) != 2 || lines[1] != "update memory=2048" {
		t.Fatalf("after params change, marker lines = %v, want a second line %q", lines, "update memory=2048")
	}

	s = readState(t, workDir)
	cap = s.Inventory.Nodes["node1"].Capabilities["integ-test/update-cap"]
	if cap == nil {
		t.Fatalf("capability not found in state.json after update")
	}
	if got := cap.Params["memory"]; got != "2048" {
		t.Errorf("state.json params.memory = %v, want %q", got, "2048")
	}
	if got, _ := cap.Store["last_operation"].(string); got != "update" {
		t.Errorf("state.json store.last_operation = %q, want %q (construct must not have re-run)", got, "update")
	}
	// Values written to store/ by construct must survive the update.
	if got, _ := cap.Store["from_construct"].(string); got != "kept" {
		t.Errorf("state.json store.from_construct = %q, want %q (store must be kept across updates)", got, "kept")
	}
}

// TestUpdate_FailedUpdateKeepsState checks that a failing update task makes
// apply fail, leaves the recorded capability (params, version, store) as it
// was, and that the update is planned again on the next run.
func TestUpdate_FailedUpdateKeepsState(t *testing.T) {
	workDir := t.TempDir()
	recipeDir, _ := filepath.Abs("testdata/update/recipe")
	capId := "integ-test/update-cap-fail"

	cli := helpers.NewCLI(testBinaryPath).
		WithWorkDir(workDir).
		WithEnv("NWS_LOG_LEVEL", "DEBUG")

	cli.RunSuccess(t, "init")

	target1024 := writeUpdateDesiredState(t, workDir, "desired-state.yaml",
		"      - id: "+capId+"\n        params:\n          memory: \"1024\"\n")
	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", target1024)
	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")

	target2048 := writeUpdateDesiredState(t, workDir, "desired-state-updated.yaml",
		"      - id: "+capId+"\n        params:\n          memory: \"2048\"\n")
	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", target2048)
	cli.RunExpectError(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")

	cap := readState(t, workDir).Inventory.Nodes["node1"].Capabilities[capId]
	if cap == nil {
		t.Fatalf("capability was removed from state.json after a failed update")
	}
	if got := cap.Params["memory"]; got != "1024" {
		t.Errorf("state.json params.memory = %v, want %q (a failed update must not record new params)", got, "1024")
	}
	if cap.Version != "1.0.0" {
		t.Errorf("state.json capability version = %q, want %q", cap.Version, "1.0.0")
	}
	if got, _ := cap.Store["from_construct"].(string); got != "kept" {
		t.Errorf("state.json store.from_construct = %q, want %q", got, "kept")
	}

	// The change is still pending, so the next plan must schedule the update again.
	stdout, _ := cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", target2048)
	helpers.AssertContains(t, stdout, "node:"+capId+"@1.0.0:node1")
}

// TestUpdate_VersionUpgrade covers the other update trigger: the recipe
// version resolved for a capability moves forward (target omits the version,
// so the latest available one is picked) even though params did not change.
// This exercises differenceCapChange's version-comparison branch, distinct
// from the params-hash branch covered by TestUpdate_ParamsChange.
func TestUpdate_VersionUpgrade(t *testing.T) {
	workDir := t.TempDir()
	recipeDir, _ := filepath.Abs("testdata/update/recipe")
	markerFile := filepath.Join(workDir, "marker.log")

	cli := helpers.NewCLI(testBinaryPath).
		WithWorkDir(workDir).
		WithEnv("NWS_LOG_LEVEL", "DEBUG")

	cli.RunSuccess(t, "init")

	capBlockPinned := `      - id: integ-test/update-cap-upgrade
        version: "1.0.0"
        params:
          memory: "1024"
          markerFile: "` + markerFile + `"
`
	targetPinned := writeUpdateDesiredState(t, workDir, "desired-state.yaml", capBlockPinned)

	// Step 1: construct against the explicitly pinned 1.0.0 recipe.
	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", targetPinned)
	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")

	if lines := readMarkerLines(t, markerFile); len(lines) != 1 || lines[0] != "construct-v1.0.0 memory=1024" {
		t.Fatalf("after initial construct, marker lines = %v, want exactly [%q]", lines, "construct-v1.0.0 memory=1024")
	}

	s := readState(t, workDir)
	cap := s.Inventory.Nodes["node1"].Capabilities["integ-test/update-cap-upgrade"]
	if cap == nil {
		t.Fatalf("capability not found in state.json after construct")
	}
	if cap.Version != "1.0.0" {
		t.Fatalf("state.json capability version = %q, want %q", cap.Version, "1.0.0")
	}

	// Step 2: omit the version (so the latest, 1.1.0, is resolved) and keep
	// params unchanged, to isolate the version-upgrade trigger from the
	// params-change trigger already covered by TestUpdate_ParamsChange.
	capBlockLatest := `      - id: integ-test/update-cap-upgrade
        params:
          memory: "1024"
          markerFile: "` + markerFile + `"
`
	targetLatest := writeUpdateDesiredState(t, workDir, "desired-state-updated-version.yaml", capBlockLatest)

	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", targetLatest)
	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")

	lines := readMarkerLines(t, markerFile)
	if len(lines) != 2 || lines[1] != "update-v1.1.0 memory=1024" {
		t.Fatalf("after version upgrade, marker lines = %v, want a second line %q", lines, "update-v1.1.0 memory=1024")
	}

	s = readState(t, workDir)
	cap = s.Inventory.Nodes["node1"].Capabilities["integ-test/update-cap-upgrade"]
	if cap == nil {
		t.Fatalf("capability not found in state.json after upgrade")
	}
	if cap.Version != "1.1.0" {
		t.Errorf("state.json capability version = %q, want %q", cap.Version, "1.1.0")
	}
	if got, _ := cap.Store["last_operation"].(string); got != "update-v1.1.0" {
		t.Errorf("state.json store.last_operation = %q, want %q (construct must not have re-run)", got, "update-v1.1.0")
	}
}

// constructThenPlanParamsChange constructs capId with memory=1024, then
// writes a target with memory=2048 so the next plan schedules an update.
// It returns the updated target path.
func constructThenPlanParamsChange(t *testing.T, cli *helpers.CLI, workDir, recipeDir, capId string) string {
	t.Helper()

	cli.RunSuccess(t, "init")

	target1024 := writeUpdateDesiredState(t, workDir, "desired-state.yaml",
		"      - id: "+capId+"\n        params:\n          memory: \"1024\"\n")
	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", target1024)
	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")

	return writeUpdateDesiredState(t, workDir, "desired-state-updated.yaml",
		"      - id: "+capId+"\n        params:\n          memory: \"2048\"\n")
}

// TestUpdate_InvalidToolRunFailsAtPlan checks that a tool.run appearing only
// in an `operation: update` task is validated at plan time, the same way as
// one in a construct task. Without this, the invalid command slipped through
// plan and made apply panic.
func TestUpdate_InvalidToolRunFailsAtPlan(t *testing.T) {
	workDir := t.TempDir()
	recipeDir, _ := filepath.Abs("testdata/update/recipe")

	cli := helpers.NewCLI(testBinaryPath).
		WithWorkDir(workDir).
		WithEnv("NWS_LOG_LEVEL", "DEBUG")

	target := constructThenPlanParamsChange(t, cli, workDir, recipeDir, "integ-test/update-cap-bad-toolrun")

	_, stderr := cli.RunExpectError(t, "plan", "--recipe-dir", recipeDir, "--target", target)
	helpers.AssertContains(t, stderr, `failed to find task by command "no-such-command"`)
}

// TestUpdate_ToolRunAdapterIsFingerprinted checks that an adapter referenced
// only from an `operation: update` task is recorded in the plan's recipe
// fingerprints, so that changes to it between plan and apply are detected.
func TestUpdate_ToolRunAdapterIsFingerprinted(t *testing.T) {
	workDir := t.TempDir()
	recipeDir, _ := filepath.Abs("testdata/update/recipe")

	cli := helpers.NewCLI(testBinaryPath).
		WithWorkDir(workDir).
		WithEnv("NWS_LOG_LEVEL", "DEBUG")

	target := constructThenPlanParamsChange(t, cli, workDir, recipeDir, "integ-test/update-cap-toolrun")

	cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", target)

	data, err := os.ReadFile(filepath.Join(workDir, "plan.json"))
	if err != nil {
		t.Fatalf("failed to read plan.json: %v", err)
	}
	var p plan.Plan
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("failed to parse plan.json: %v", err)
	}

	const adapterFqid = "integ-test/update-adapter@1.0.0"
	found := false
	for _, r := range p.Metadata.Recipes {
		if r.Fqid == adapterFqid {
			found = true
		}
	}
	if !found {
		t.Errorf("plan metadata.recipes does not include %q: %+v", adapterFqid, p.Metadata.Recipes)
	}

	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")
}

func readPlan(t *testing.T, workDir string) *plan.Plan {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(workDir, "plan.json"))
	if err != nil {
		t.Fatalf("failed to read plan.json: %v", err)
	}
	var p plan.Plan
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("failed to parse plan.json: %v", err)
	}
	return &p
}

// TestUpdate_NoUpdateTasksLeavesChangePending checks that a params change on a
// capability whose recipe has no `operation: update` task is not applied: no
// job is planned, the change is reported as pending, and the recorded params
// stay as they were so that state keeps matching what was actually applied.
func TestUpdate_NoUpdateTasksLeavesChangePending(t *testing.T) {
	workDir := t.TempDir()
	recipeDir, _ := filepath.Abs("testdata/update/recipe")
	capId := "integ-test/update-cap-no-update"
	jobId := "node:" + capId + "@1.0.0:node1"

	cli := helpers.NewCLI(testBinaryPath).
		WithWorkDir(workDir).
		WithEnv("NWS_LOG_LEVEL", "DEBUG")

	target := constructThenPlanParamsChange(t, cli, workDir, recipeDir, capId)

	stdout, _ := cli.RunSuccess(t, "plan", "--recipe-dir", recipeDir, "--target", target)
	helpers.AssertContains(t, stdout, "Pending Updates (recipe has no update tasks):\n- "+jobId)

	p := readPlan(t, workDir)
	if !slices.Contains(p.PendingUpdates, jobId) {
		t.Errorf("plan pendingUpdates = %v, want it to contain %q", p.PendingUpdates, jobId)
	}
	if g := p.ExecutionPlan.ConstructGraph; g != nil {
		for _, j := range g.Jobs {
			if j.Id == jobId {
				t.Errorf("job %q was planned although its recipe has no update tasks", jobId)
			}
		}
	}

	cli.RunSuccess(t, "apply", "--recipe-dir", recipeDir, "--plan", "plan.json")

	cap := readState(t, workDir).Inventory.Nodes["node1"].Capabilities[capId]
	if cap == nil {
		t.Fatalf("capability %q not found in state.json", capId)
	}
	if got := cap.Params["memory"]; got != "1024" {
		t.Errorf("state.json params.memory = %v, want %q (a change that was not applied must not be recorded)", got, "1024")
	}
}
