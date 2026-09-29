// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_local

import (
	"context"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/types"
	"github.com/sony/niwashi/internal/workspace"
)

func newMockWorkspace() *Workspace {
	return NewWorkspace(
		func() *workspace.TaskWorkspace {
			r := workspace.NewSimplePathResolver(
				"./testdata",
				"xxxxxx",
				"node",
				"testnode",
				"test/test.test.test@1.0.0",
			)
			r.Task.Name = "mock_task"
			return r
		}(),
		WithRuntime(&runtime{}),
		WithFileSystem(&mockFileSystem{}),
	)
}

type runtime struct{}

func (r *runtime) GetToolAlias() map[string]string {
	return nil
}

func (r *runtime) GetCapabilityBinding() map[string]string {
	return nil
}

func (r *runtime) GetTools() types.Dict {
	return nil
}

func (r *runtime) GetService() types.Dict {
	return nil
}

func TestExecLocalAction_Execute(t *testing.T) {

	cmdFactory := NewMockCmdFactory()

	tests := []struct {
		name        string
		deps        action.ActionContext
		wantErr     bool
		expectedCmd *mockCmd
	}{
		{
			name: "normal execution",
			deps: NewDefaultActionContext(
				WithTask(NewTaskContext(
					newDefaultTaskSpec(),
					nil,
					nil,
				)),
				WithRunContext(
					NewRunContext(WithCmdFactory(cmdFactory)),
				),
				WithSubject(NewSubject("dummyScope", "dummyId")),
				WithWorkspace(newMockWorkspace()),
				WithTemplateParams(&action.TemplateParams{
					Paths: map[string]string{
						"work_dir": "./testdata/workspace",
					},
				}),
			),
			wantErr: false,
			expectedCmd: &mockCmd{
				args: []string{
					"./testdata/runs/xxxxxx/node-testnode/test_test.test.test_1.0.0/logs/mock_task.action.sh",
				},
				env: []string{
					"HOME=" + os.Getenv("HOME"),
					"TEST_ENV=test_value",
					"PATH=" + os.Getenv("PATH"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cmdFactory.LastCmd = nil

			e, err := BuildExecLocalAction(tt.deps)
			if err != nil {
				t.Fatalf("BuildExecLocalAction() error = %v", err)
			}
			if err := e.Execute(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("ExecLocalAction.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.expectedCmd == nil {
				return
			}

			for _, expected := range tt.expectedCmd.env {
				if !slices.Contains(cmdFactory.LastCmd.env, expected) {
					t.Errorf("Env %s not found in %v", strings.Split(expected, "=")[0], cmdFactory.LastCmd.env)
				}
			}

			if len(cmdFactory.LastCmd.args) != len(tt.expectedCmd.args) {
				t.Errorf("Expected args length %d, got %d", len(tt.expectedCmd.args), len(cmdFactory.LastCmd.args))
			}
			for i, expected := range tt.expectedCmd.args {
				if cmdFactory.LastCmd.args[i] != expected {
					t.Errorf("Expected=%s, Got=%s", expected, cmdFactory.LastCmd.args[i])
				}
			}
		})
	}
}

func TestExecLocalAction_ExecuteAdapter(t *testing.T) {

	cmdFactory := NewMockCmdFactory()

	tests := []struct {
		name        string
		deps        action.ActionContext
		wantErr     bool
		expectedCmd *mockCmd
	}{
		{
			name: "normal execution via adapter",
			deps: NewDefaultActionContext(
				WithTask(NewTaskContext(
					newDefaultTaskSpec(),
					newDefaultCallerTaskSpec(),
					nil,
				)),
				WithRunContext(
					NewRunContext(WithCmdFactory(cmdFactory)),
				),
				WithSubject(NewSubject("dummyScope", "dummyId")),
				WithWorkspace(newMockWorkspace()),
				WithTemplateParams(&action.TemplateParams{
					Paths: map[string]string{
						"work_dir": "./testdata/workspace",
					},
				}),
			),
			wantErr: false,
			expectedCmd: &mockCmd{
				args: []string{
					"./testdata/runs/xxxxxx/node-testnode/test_test.test.test_1.0.0/logs/mock_task.action.sh",
					"hello",
					"world",
				},
				env: []string{
					"HOME=" + os.Getenv("HOME"),
					"TEST_ENV=test_value",
					"PATH=" + os.Getenv("PATH"),
					"USER=" + os.Getenv("USER"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cmdFactory.LastCmd = nil

			e, err := BuildExecLocalAction(tt.deps)
			if err != nil {
				t.Fatalf("BuildExecLocalAction() error = %v", err)
			}
			if err := e.Execute(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("ExecLocalAction.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.expectedCmd == nil {
				return
			}

			for _, expected := range tt.expectedCmd.env {
				if !slices.Contains(cmdFactory.LastCmd.env, expected) {
					t.Errorf("Env %s not found in %v", strings.Split(expected, "=")[0], cmdFactory.LastCmd.env)
				}
			}

			if len(cmdFactory.LastCmd.args) != len(tt.expectedCmd.args) {
				t.Errorf("Expected args length %d, got %d", len(tt.expectedCmd.args), len(cmdFactory.LastCmd.args))
			}
			for i, expected := range tt.expectedCmd.args {
				if cmdFactory.LastCmd.args[i] != expected {
					t.Errorf("Expected=%s, Got=%s", expected, cmdFactory.LastCmd.args[i])
				}
			}
		})
	}
}
