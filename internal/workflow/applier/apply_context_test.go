// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier

import (
	"context"
	"errors"
	"testing"

	"github.com/sony/niwashi/internal/node"
	"github.com/sony/niwashi/internal/state"
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

func Test_applyContext_TargetStageView(t *testing.T) {
	t.Run("refers to the target state", func(t *testing.T) {
		target := state.NewState()
		target.Inventory.Nodes["node-1"] = &node.Node{}

		// the current state does not have node-1
		c := &applyContext{
			stateManager: NewStateManager(state.NewState()),
			targetState:  target,
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

func Test_applyContext_WithOperation(t *testing.T) {
	c := &applyContext{}
	got := c.WithOperation("construct")

	if got != c {
		t.Errorf("WithOperation() must return the receiver itself")
	}
	if c.Operation() != "construct" {
		t.Errorf("Operation() = %q, want %q", c.Operation(), "construct")
	}
}

func Test_applyContext_TaskFinished(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantCompleted []string
		wantFailed    int
	}{
		{
			name:          "completed task",
			err:           nil,
			wantCompleted: []string{"job-1"},
			wantFailed:    0,
		},
		{
			name:          "failed task",
			err:           errors.New("task error"),
			wantCompleted: nil,
			wantFailed:    1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// File is nil, so SaveState does nothing.
			c := (&applyContext{
				stateManager: NewStateManager(state.NewState()),
			}).WithOperation("construct")

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
				if ft.TaskId != "job-1" || ft.Operation != "construct" || ft.Error != tt.err.Error() {
					t.Errorf("FailedTasks[0] = %+v, want taskId=job-1 operation=construct error=%q", ft, tt.err.Error())
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
