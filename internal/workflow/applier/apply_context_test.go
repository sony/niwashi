// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/cap"
	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/plan"
	"github.com/sony/niwashi/internal/recipe"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/types"
	"github.com/sony/niwashi/internal/workflow"
)

// applyContext must satisfy the interface used by the phase implementations.
var _ workflow.ApplyContext = (*applyContext)(nil)

type fakeTask struct {
	job *workflow.Job
}

func (t *fakeTask) GetName() string               { return "fake" }
func (t *fakeTask) GetSubject() string            { return "" }
func (t *fakeTask) GetAffectSubjects() []string   { return nil }
func (t *fakeTask) GetJob() *workflow.Job         { return t.job }
func (t *fakeTask) Run(ctx context.Context) error { return nil }
func (t *fakeTask) AfterRun(runError error) error { return nil }

type fakeTaskResult struct {
	task workflow.Task
	err  error
}

func (r *fakeTaskResult) GetTask() workflow.Task { return r.task }
func (r *fakeTaskResult) GetError() error        { return r.err }

// newStageStateManager returns a stageStateManager for the stage on top of a new stateManager.
func newStageStateManager(s *state.State, stage string) *stageStateManager {
	return &stageStateManager{
		stateManager: NewStateManager(s),
		stageName:    stage,
	}
}

func Test_applyContext_TargetStageView(t *testing.T) {
	t.Run("refers to the target state", func(t *testing.T) {
		target := state.NewState()
		target.Inventory.Nodes["node-1"] = &node.Node{}

		// the current state does not have node-1
		c := &applyContext{
			stageStateManager: newStageStateManager(state.NewState(), state.MainStageName),
			targetState:       target,
		}
		c, err := c.WithStage(state.DefaultStageName)
		if err != nil {
			t.Fatalf("WithStage() error = %v", err)
		}

		got := c.TargetStageView()
		if got == nil {
			t.Fatalf("TargetStageView() = nil, want non-nil")
		}
		if got.Inventory != target.Inventory || got.Infrastructure != target.Infrastructure || got.Runtime != target.Runtime {
			t.Errorf("TargetStageView() must refer to the fields of the target state")
		}
		if got.GetNode("node-1") == nil {
			t.Errorf("TargetStageView() must return the target state, not the current state")
		}
	})

	t.Run("nil target state", func(t *testing.T) {
		c := &applyContext{}
		if got := c.TargetStageView(); got != nil {
			t.Errorf("TargetStageView() = %+v, want nil", got)
		}
	})
}

func Test_applyContext_TaskFinished(t *testing.T) {
	// the job ID is recorded with the stage name as a prefix
	tests := []struct {
		name          string
		stage         string
		err           error
		wantCompleted []string
		wantFailed    int
		wantTaskId    string
	}{
		{
			name:          "completed task",
			stage:         state.MainStageName,
			err:           nil,
			wantCompleted: []string{"main/job-1"},
			wantFailed:    0,
		},
		{
			name:          "completed task in a pre-stage",
			stage:         "bootstrap",
			err:           nil,
			wantCompleted: []string{"bootstrap/job-1"},
			wantFailed:    0,
		},
		{
			name:          "failed task",
			stage:         state.MainStageName,
			err:           errors.New("task error"),
			wantCompleted: nil,
			wantFailed:    1,
			wantTaskId:    "main/job-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// File is nil, so SaveState does nothing.
			c := (&applyContext{
				stageStateManager: newStageStateManager(state.NewState(), tt.stage),
				operation:         "construct",
			})

			res := &fakeTaskResult{
				task: &fakeTask{job: &workflow.Job{Id: "job-1"}},
				err:  tt.err,
			}
			if err := c.TaskFinished(res); err != nil {
				t.Fatalf("TaskFinished() error = %v", err)
			}

			aa := c.State().Status.ActiveApply
			got := aa.CompletedTasks["construct"]
			if len(got) != len(tt.wantCompleted) || (len(got) > 0 && got[0] != tt.wantCompleted[0]) {
				t.Errorf("CompletedTasks[construct] = %v, want %v", got, tt.wantCompleted)
			}
			if len(aa.FailedTasks) != tt.wantFailed {
				t.Fatalf("len(FailedTasks) = %d, want %d", len(aa.FailedTasks), tt.wantFailed)
			}
			if tt.wantFailed > 0 {
				ft := aa.FailedTasks[0]
				if ft.TaskId != tt.wantTaskId || ft.Operation != "construct" || ft.Error != tt.err.Error() {
					t.Errorf("FailedTasks[0] = %+v, want taskId=%s operation=construct error=%q", ft, tt.wantTaskId, tt.err.Error())
				}
			}
		})
	}
}

func TestNewRunContext(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *Config
		wantSim      bool
		wantDryRun   string
		wantDryRunOn bool
	}{
		{
			name:         "default",
			cfg:          &Config{},
			wantSim:      false,
			wantDryRun:   "",
			wantDryRunOn: false,
		},
		{
			name:         "dry run off",
			cfg:          &Config{DryRunMode: "off"},
			wantSim:      false,
			wantDryRun:   "off",
			wantDryRunOn: false,
		},
		{
			name:         "dry run enabled",
			cfg:          &Config{DryRunMode: "on"},
			wantSim:      true,
			wantDryRun:   "on",
			wantDryRunOn: true,
		},
		{
			name:         "simCmd dev option",
			cfg:          &Config{DevOption: NewDevOption(map[string]string{"simCmd": ""})},
			wantSim:      true,
			wantDryRun:   "",
			wantDryRunOn: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := NewRunContext(tt.cfg)

			_, isSim := rc.CmdFactory.(*SimCmdFactory)
			if isSim != tt.wantSim {
				t.Errorf("CmdFactory is %T, want simulated = %v", rc.CmdFactory, tt.wantSim)
			}
			if got := rc.DryRun(); got != tt.wantDryRun {
				t.Errorf("DryRun() = %q, want %q", got, tt.wantDryRun)
			}
			if got := rc.DryRunEnabled(); got != tt.wantDryRunOn {
				t.Errorf("DryRunEnabled() = %v, want %v", got, tt.wantDryRunOn)
			}
		})
	}

	t.Run("nil run context", func(t *testing.T) {
		var rc *runContext
		if got := rc.DryRun(); got != "" {
			t.Errorf("DryRun() = %q, want empty", got)
		}
		if rc.DryRunEnabled() {
			t.Errorf("DryRunEnabled() = true, want false")
		}
	})
}

func newStageGraphs(stages ...string) map[string]*plan.Workflow {
	m := map[string]*plan.Workflow{}
	for _, stage := range stages {
		m[stage] = &plan.Workflow{Jobs: []*plan.Job{{Id: stage}}}
	}
	return m
}

// newWithStageFixture returns a root context whose target state has a pre-stage
// named "bootstrap". built is set to true when the buildAction is called.
func newWithStageFixture(built *bool) *applyContext {
	target := state.NewState()
	target.Stages["bootstrap"] = &state.Stage{}
	target.SetDefaults()

	return &applyContext{
		stageStateManager: newStageStateManager(state.NewState(), state.MainStageName),
		plan: &plan.Plan{
			Mode: plan.ModeDefault,
			ExecutionPlan: plan.ExecutionPlan{
				ConstructGraph: newStageGraphs(state.MainStageName, "bootstrap"),
				DestructGraph:  newStageGraphs(state.MainStageName),
			},
		},
		targetState: target,
		operation:   recipe.OperationConstruct,
		buildAction: func(ctx action.ActionContext) (action.Action, error) {
			*built = true
			return nil, nil
		},
	}
}

func Test_applyContext_WithStage(t *testing.T) {
	var built bool
	root := newWithStageFixture(&built)
	target := root.targetState
	pln := root.plan

	t.Run("pre-stage", func(t *testing.T) {
		ctx, err := root.WithStage("bootstrap")
		if err != nil {
			t.Fatalf("WithStage() error = %v", err)
		}
		if ctx == root {
			t.Fatalf("WithStage() must return a copy")
		}
		if ctx.TargetStageView().Inventory != target.Stages["bootstrap"].Inventory {
			t.Errorf("TargetStageView() must refer to the pre-stage of the target state")
		}
		if ctx.TargetStageView().Runtime != target.Runtime {
			t.Errorf("TargetStageView() must refer to the runtime shared by all stages")
		}
		if ctx.workflowGraph != pln.ExecutionPlan.ConstructGraph["bootstrap"] {
			t.Errorf("workflowGraph must be the construct graph of the stage")
		}
	})

	t.Run("stage without graph", func(t *testing.T) {
		ctx, err := root.WithStage(state.InitStageName)
		if err != nil {
			t.Fatalf("WithStage() error = %v", err)
		}
		if ctx.workflowGraph != nil {
			t.Errorf("workflowGraph = %+v, want nil", ctx.workflowGraph)
		}
	})

	t.Run("stage missing in the target state", func(t *testing.T) {
		// e.g. a pre-stage removed from the target state by prune
		ctx, err := root.WithStage("removed")
		if err != nil {
			t.Fatalf("WithStage() error = %v", err)
		}
		assertEmptyStageView(t, ctx.TargetStageView())
		if ctx.TargetStageView().Runtime != target.Runtime {
			t.Errorf("TargetStageView() must refer to the runtime shared by all stages")
		}
		if ctx.TargetState() != target {
			t.Errorf("TargetState() must be the same as the root context")
		}
	})

	t.Run("nil target state", func(t *testing.T) {
		c := &applyContext{
			stageStateManager: newStageStateManager(state.NewState(), state.MainStageName),
			plan:              pln,
		}
		ctx, err := c.WithStage(state.MainStageName)
		if err != nil {
			t.Fatalf("WithStage() error = %v", err)
		}
		assertEmptyStageView(t, ctx.TargetStageView())
	})
}

// assertEmptyStageView checks that the view is not nil and has no nodes, clusters and generators.
func assertEmptyStageView(t *testing.T, sv *state.StageView) {
	t.Helper()
	if sv == nil || sv.Inventory == nil || sv.Infrastructure == nil {
		t.Fatalf("TargetStageView() = %+v, want an empty view", sv)
	}
	if len(sv.Inventory.Nodes) != 0 || len(sv.Inventory.Clusters) != 0 || len(sv.Infrastructure.Generators) != 0 {
		t.Errorf("TargetStageView() = %+v, want an empty view", sv)
	}
}

func Test_applyContext_WithStage_Copy(t *testing.T) {
	var built bool
	root := newWithStageFixture(&built)

	t.Run("inherits the root context", func(t *testing.T) {
		ctx, err := root.WithStage(state.MainStageName)
		if err != nil {
			t.Fatalf("WithStage() error = %v", err)
		}
		if ctx.Operation() != recipe.OperationConstruct {
			t.Errorf("Operation() = %q, want %q", ctx.Operation(), recipe.OperationConstruct)
		}
		// stateManager is the one embedded in stageStateManager
		if ctx.stateManager != root.stateManager {
			t.Errorf("the stateManager must be shared with the root context")
		}
		built = false
		if _, err := ctx.BuildAction(nil); err != nil || !built {
			t.Errorf("BuildAction() must call the buildAction of the root context")
		}
	})

	t.Run("does not modify the root context", func(t *testing.T) {
		if _, err := root.WithStage("bootstrap"); err != nil {
			t.Fatalf("WithStage() error = %v", err)
		}
		if root.targetStageView != nil || root.workflowGraph != nil {
			t.Errorf("WithStage() must not modify the root context")
		}
		if root.stageName != state.MainStageName {
			t.Errorf("root stageName = %q, want %q", root.stageName, state.MainStageName)
		}
	})

	t.Run("switches the stage of the state manager", func(t *testing.T) {
		ctx, err := root.WithStage("bootstrap")
		if err != nil {
			t.Fatalf("WithStage() error = %v", err)
		}
		if ctx.stageStateManager == root.stageStateManager {
			t.Errorf("WithStage() must create a new stageStateManager")
		}
		if ctx.stageName != "bootstrap" {
			t.Errorf("stageName = %q, want %q", ctx.stageName, "bootstrap")
		}
	})

}

func Test_applyContext_WithStage_GraphByMode(t *testing.T) {
	tests := []struct {
		mode      plan.Mode
		wantGraph func(p *plan.Plan) *plan.Workflow
	}{
		{plan.ModeDefault, func(p *plan.Plan) *plan.Workflow { return p.ExecutionPlan.ConstructGraph[state.MainStageName] }},
		{plan.ModeDestroy, func(p *plan.Plan) *plan.Workflow { return p.ExecutionPlan.DestructGraph[state.MainStageName] }},
		{plan.ModePrune, func(p *plan.Plan) *plan.Workflow { return p.ExecutionPlan.DestructGraph[state.MainStageName] }},
	}
	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			pln := &plan.Plan{
				Mode: tt.mode,
				ExecutionPlan: plan.ExecutionPlan{
					ConstructGraph: newStageGraphs(state.MainStageName),
					DestructGraph:  newStageGraphs(state.MainStageName),
				},
			}
			c := &applyContext{
				stageStateManager: newStageStateManager(state.NewState(), state.MainStageName),
				plan:              pln,
				targetState:       state.NewState(),
			}

			ctx, err := c.WithStage(state.MainStageName)
			if err != nil {
				t.Fatalf("WithStage() error = %v", err)
			}
			if ctx.workflowGraph == nil || ctx.workflowGraph != tt.wantGraph(pln) {
				t.Errorf("workflowGraph = %+v, want %+v", ctx.workflowGraph, tt.wantGraph(pln))
			}
		})
	}
}

// newExportFixture returns a root context whose current state has the pre-stage
// "bootstrap" exporting "proxy_ca_cert" from the node "gateway". The main stage
// has no node, so looking up the source in the main stage fails.
func newExportFixture(mode plan.Mode, path string) *applyContext {
	current := state.NewState()
	bootstrap := &state.Stage{Inventory: state.NewInventory()}
	bootstrap.Inventory.Nodes["gateway"] = &node.Node{Capabilities: cap.CapabilityList{
		"org/squid-proxy": {Store: types.Dict{"ca": types.Dict{"cert": "PEM"}}},
	}}
	bootstrap.Export = state.Exports{
		"proxy_ca_cert": {Node: "gateway", Capability: "org/squid-proxy", Path: path},
	}
	current.Stages["bootstrap"] = bootstrap
	current.SetDefaults()

	graphs := newStageGraphs(state.InitStageName, "bootstrap", state.MainStageName)
	return &applyContext{
		stageStateManager: newStageStateManager(current, state.MainStageName),
		plan: &plan.Plan{
			Mode:          mode,
			ExecutionPlan: plan.ExecutionPlan{ConstructGraph: graphs, DestructGraph: graphs},
		},
		targetState: state.NewState(),
		operation:   recipe.OperationConstruct,
	}
}

func Test_applyContext_WithStage_ExportValues(t *testing.T) {
	tests := []struct {
		name  string
		mode  plan.Mode
		stage string
		want  map[string]any
	}{
		{
			// the source is in the pre-stage, not in the main stage
			name:  "main stage gets the exports of the pre-stage",
			mode:  plan.ModeDefault,
			stage: state.MainStageName,
			want:  map[string]any{"proxy_ca_cert": "PEM"},
		},
		{
			// a stage can import only from the preceding stages
			name:  "pre-stage does not get its own exports",
			mode:  plan.ModeDefault,
			stage: "bootstrap",
			want:  map[string]any{},
		},
		{
			name:  "init stage gets nothing",
			mode:  plan.ModeDefault,
			stage: state.InitStageName,
			want:  map[string]any{},
		},
		{
			// the stages run in reverse order, but "preceding" follows the construct order
			name:  "destroy: main stage gets the exports of the pre-stage",
			mode:  plan.ModeDestroy,
			stage: state.MainStageName,
			want:  map[string]any{"proxy_ca_cert": "PEM"},
		},
		{
			name:  "prune: main stage gets the exports of the pre-stage",
			mode:  plan.ModePrune,
			stage: state.MainStageName,
			want:  map[string]any{"proxy_ca_cert": "PEM"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := newExportFixture(tt.mode, "ca/cert").WithStage(tt.stage)
			if err != nil {
				t.Fatalf("WithStage() error = %v", err)
			}
			got := ctx.GetExportValues()
			if len(got) != len(tt.want) {
				t.Fatalf("GetExportValues() = %v, want %v", got, tt.want)
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("GetExportValues()[%q] = %v, want %v", k, got[k], v)
				}
			}
		})
	}

	t.Run("path not found stops at the start of the stage", func(t *testing.T) {
		_, err := newExportFixture(plan.ModeDefault, "ca/missing").WithStage(state.MainStageName)
		if err == nil {
			t.Fatal("WithStage() succeeded unexpectedly")
		}
		if !strings.Contains(err.Error(), `"ca/missing"`) {
			t.Errorf("WithStage() error = %v, want the path in it", err)
		}
	})

	t.Run("no export", func(t *testing.T) {
		root := newExportFixture(plan.ModeDefault, "ca/cert")
		root.state.Stages["bootstrap"].Export = nil
		ctx, err := root.WithStage(state.MainStageName)
		if err != nil {
			t.Fatalf("WithStage() error = %v", err)
		}
		if got := ctx.GetExportValues(); len(got) != 0 {
			t.Errorf("GetExportValues() = %v, want empty", got)
		}
	})
}
