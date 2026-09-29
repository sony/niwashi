// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_remote

import (
	"context"
	"strings"
	"testing"

	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/platform"
	"github.com/sony/niwashi/internal/transport"
	"github.com/sony/niwashi/internal/transport/ssh"
	"github.com/sony/niwashi/internal/types"
	"github.com/sony/niwashi/internal/workspace"
)

type taskSpec struct {
	*recipeSpec
	name       string
	actionType string
	actionSpec action.ActionSpec
}

func (s *taskSpec) GetRecipe() action.RecipeSpec {
	return s.recipeSpec
}

func (s *taskSpec) GetName() string {
	return s.name
}

func (s *taskSpec) GetActionType() string {
	return s.actionType
}

func (s *taskSpec) GetActionSpec() action.ActionSpec {
	return s.actionSpec
}

type recipeSpec struct {
	fqid          string
	dir           string
	assets        []string
	defaultParams map[string]any
	defaultEnv    map[string]string
}

func (s *recipeSpec) GetFqid() string {
	return s.fqid
}

func (s *recipeSpec) GetDir() string {
	return s.dir
}

func (s *recipeSpec) GetAssets() []string {
	return s.assets
}

func (s *recipeSpec) GetDefaultParams() map[string]any {
	return s.defaultParams
}

func (s *recipeSpec) GetDefaultEnv() map[string]string {
	return s.defaultEnv
}

type callerTaskSpec struct {
	*recipeSpec
	name       string
	actionType string
	actionSpec action.ActionSpec
}

func (s *callerTaskSpec) GetRecipe() action.RecipeSpec {
	return s.recipeSpec
}

func (s *callerTaskSpec) GetName() string {
	return s.name
}

func (s *callerTaskSpec) GetActionType() string {
	return s.actionType
}

func (s *callerTaskSpec) GetActionSpec() action.ActionSpec {
	return s.actionSpec
}

func (s *callerTaskSpec) GetArgvTpl() []string {
	return nil
}

func (s *callerTaskSpec) GetWorkDirTpl() string {
	return ""
}

func (s *callerTaskSpec) GetInheritEnv() []string {
	return nil
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

func TestSshExecAction_Execute(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		deps    action.ActionContext
		wantErr bool
	}{
		{
			name: "normal execution",
			deps: NewDefaultActionContext(
				WithTask(NewTaskContext(
					&taskSpec{
						recipeSpec: &recipeSpec{
							fqid: "test/test.test.test@1.0.0",
							dir:  "testdata",
							assets: []string{
								"testdata/asset.txt",
								"testdata/hoge",
								"xxxx.sh",
							},
						},
						name:       "test dummy",
						actionType: "exec.remote",
						actionSpec: &Spec{},
					},
					nil,
					nil,
				)),
				WithSubject(NewSubject("dummyScope", "dummyId",
					WithConnection(&transport.Connection{
						Data: map[string]transport.Transport{
							ssh.TypeName: &ssh.Transport{
								Address: &ssh.Address{
									Host: "localhost",
									Port: 22,
									User: "testuser",
								},
								Auth: &ssh.Auth{
									Method:         "privateKey",
									PrivateKeyPath: "testprivatekeypath",
								},
								HostKey: &ssh.HostKey{
									KnownHostsPath: "/tmp/path/to/known_hosts",
								},
							},
						},
					}),
					WithOperator(&dummyOperator{}),
				),
				),
				WithWorkspace(NewWorkspace(
					workspace.NewSimplePathResolver(
						"./testdata",
						"xxxxxx",
						"node",
						"testnode",
						"test/test.test.test@1.0.0",
					),
					WithRuntime(&runtime{}),
					WithFileSystem(&mockFileSystem{}),
				)),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := NewRemoteExecAction(tt.deps)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			gotErr := e.Execute(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Execute() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Execute() succeeded unexpectedly")
			}

			// for _, h := range tt.deps.SshClientFactory.(*mockSshClientFactory).History {
			// 	t.Error(h)
			// }
		})
	}
}

func TestRemoteSeparator(t *testing.T) {
	tests := []struct {
		os   string
		want string
	}{
		{os: "windows", want: "\\"},
		{os: "linux", want: "/"},
		{os: platform.OsUnknown, want: "/"},
		{os: "", want: "/"},
	}
	for _, tt := range tests {
		t.Run(tt.os, func(t *testing.T) {
			if got := remoteSeparator(tt.os); got != tt.want {
				t.Errorf("remoteSeparator(%q) = %q, want %q", tt.os, got, tt.want)
			}
		})
	}
}

func TestRemoteExecAction_CreateRemoteWorkspace(t *testing.T) {
	calleeFqid := "test/test.test.test@1.0.0"
	callerFqid := "test/caller@1.0.0"

	tests := []struct {
		name string
		os   string
		sep  string
	}{
		{name: "unknown OS defaults to unix separator", os: platform.OsUnknown, sep: "/"},
		{name: "windows node uses backslash separator", os: "windows", sep: "\\"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewDefaultActionContext(
				WithTask(NewTaskContext(
					&taskSpec{
						recipeSpec: &recipeSpec{fqid: calleeFqid, dir: "testdata"},
						name:       "test dummy",
						actionType: "exec.remote",
						actionSpec: &Spec{},
					},
					&callerTaskSpec{
						recipeSpec: &recipeSpec{fqid: callerFqid, dir: "testdata"},
						name:       "caller",
						actionType: "exec.remote",
					},
					nil,
				)),
				WithSubject(NewSubject("dummyScope", "dummyId",
					WithOs(tt.os),
					WithConnection(&transport.Connection{
						Data: map[string]transport.Transport{
							ssh.TypeName: &ssh.Transport{
								Address: &ssh.Address{
									Host: "localhost",
									Port: 22,
									User: "testuser",
								},
								Auth: &ssh.Auth{
									Method:         "privateKey",
									PrivateKeyPath: "testprivatekeypath",
								},
								HostKey: &ssh.HostKey{
									KnownHostsPath: "/tmp/path/to/known_hosts",
								},
							},
						},
					}),
					WithOperator(&dummyOperator{}),
				)),
				WithWorkspace(NewWorkspace(
					workspace.NewSimplePathResolver(
						"./testdata",
						"xxxxxx",
						"node",
						"testnode",
						calleeFqid,
					),
					WithRuntime(&runtime{}),
					WithFileSystem(&mockFileSystem{}),
				)),
			)

			e, err := NewRemoteExecAction(ctx)
			if err != nil {
				t.Fatalf("could not construct receiver: %v", err)
			}

			remoteEnv, recipeDirs, err := e.CreateRemoteWorkspace(context.Background(), &mockRemoteSession{})
			if err != nil {
				t.Fatalf("CreateRemoteWorkspace() failed: %v", err)
			}

			wantCalleeSuffix := tt.sep + file.SanitizeFilename(calleeFqid)
			if !strings.HasSuffix(recipeDirs.CalleeRecipeDir, wantCalleeSuffix) {
				t.Errorf("CalleeRecipeDir = %q, want suffix %q", recipeDirs.CalleeRecipeDir, wantCalleeSuffix)
			}

			// Regression test: CallerRecipeDir used to always be empty
			// because the join result inside the closure was discarded.
			wantCallerSuffix := tt.sep + file.SanitizeFilename(callerFqid)
			if !strings.HasSuffix(recipeDirs.CallerRecipeDir, wantCallerSuffix) {
				t.Errorf("CallerRecipeDir = %q, want suffix %q", recipeDirs.CallerRecipeDir, wantCallerSuffix)
			}

			logDir := remoteEnv.GetLogDirPath()
			if !strings.Contains(logDir, tt.sep) {
				t.Errorf("GetLogDirPath() = %q, want to contain separator %q", logDir, tt.sep)
			}
		})
	}
}
