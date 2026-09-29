// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_remote

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/transport"
	"github.com/sony/niwashi/internal/workflow"
	"github.com/sony/niwashi/internal/workspace"
)

type RemoteExecAction struct {
	local *action.ActionHelper
	ctx   action.ActionContext
	spec  *Spec
	t     transport.Transport
}

func BuildRemoteExecAction(ctx action.ActionContext) (action.Action, error) {
	// only allow node-level tasks for exec.remote action since it needs connection info from subject
	if ctx.Subject().GetScope() != workflow.PhaseNode {
		return nil, fmt.Errorf("ExecRemoteAction can only be used for node-level tasks")
	}

	a, err := NewRemoteExecAction(ctx)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func NewRemoteExecAction(ctx action.ActionContext) (*RemoteExecAction, error) {

	callee := ctx.Task().Callee()
	spec := callee.GetActionSpec().(*Spec)
	if spec == nil {
		return nil, fmt.Errorf("invalid spec for exec.remote action")
	}

	// get connection info from subject
	t, err := ctx.Subject().GetTransport()
	if err != nil {
		return nil, fmt.Errorf("failed to get transport from subject: %w", err)
	}

	localHelper, err := action.NewActionHelper(ctx, nil)
	if err != nil {
		return nil, err
	}

	// resolve transport
	err = t.Resolve(func(s string) (string, error) {
		return localHelper.RenderTemplate(s)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve transport fields: %w", err)
	}

	return &RemoteExecAction{
		local: localHelper,
		ctx:   ctx,
		t:     t,
		spec:  spec,
	}, nil
}

func (e *RemoteExecAction) Execute(ctx context.Context) (err error) {

	// Create remote session via Subject's operator
	op, err := e.ctx.Subject().GetOperator(e.t)
	if err != nil {
		return fmt.Errorf("failed to get operator from subject: %w", err)
	}
	// configure operator
	op = op.WithDryRunMode(e.ctx.RunContext().DryRun()).
		WithFileSystem(e.ctx.Workspace().GetFileSystem())

	session, err := op.NewRemoteSession(ctx, e.t)
	if err != nil {
		return fmt.Errorf("failed to create remote session: %w", err)
	}
	defer func() {
		err = errors.Join(err, session.Close())
	}()

	// Prepare remote workspace
	remoteEnv, helper, err := e.setupRemoteWorkspace(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to setup remote workspace: %w", err)
	}

	localEnv := e.ctx.Workspace()

	// generate all local files to be uploaded and get their remote paths
	files, err := e.GenerateLocalFiles(localEnv, remoteEnv, helper)
	if err != nil {
		return err
	}

	// Resolve workdir
	workdir, err := helper.ResolveWorkDir(e.ctx.Task())
	if err != nil {
		return err
	}

	// Resolve arguments
	args, err := helper.ResolveArguments()
	if err != nil {
		return err
	}

	if err := files.uploader.Upload(ctx, session); err != nil {
		return err
	}

	// generate envfile
	envvars, err := helper.ResolveEnvironments(remoteEnv)
	if err != nil {
		return err
	}

	// execute remote command
	cmdArgs := files.script.GetArgs(e.ctx.Subject().GetOs())
	cmdArgs = append(cmdArgs, args...)

	exec := transport.RemoteExecution{
		Id:         file.SanitizeFilename(action.TaskName(e.ctx.Task())),
		LocalRoot:  localEnv.GetEphemeralWorkDirPath(),
		RemoteRoot: remoteEnv.GetEphemeralWorkDirPath(),
		Cmd:        cmdArgs,
		Workdir:    workdir,
		Envs:       envvars,
		Stdout:     e.ctx.Workspace().GetOutput().Stdout(),
		Stderr:     e.ctx.Workspace().GetOutput().Stderr(),
	}
	err = session.Execute(ctx, exec)
	if err != nil {
		return err
	}

	// sync output files back to local
	err = e.syncOutputFiles(ctx, session, localEnv, remoteEnv)
	return err
}

func (e *RemoteExecAction) setupRemoteWorkspace(
	ctx context.Context,
	session transport.RemoteSession) (action.Workspace, *action.ActionHelper, error) {

	// Prepare remote workspace
	remoteEnv, recipeDirs, err := e.CreateRemoteWorkspace(ctx, session)
	if err != nil {
		return nil, nil, err
	}

	// Create template params with remote environment
	t := e.local.Task()
	params := action.GetFinalParams(
		t.Params(),
		func() action.RecipeSpec {
			if t.Caller() != nil {
				return t.Caller()
			}
			return nil
		}(),
		func() action.RecipeSpec {
			if t.Callee() != nil {
				return t.Callee()
			}
			return nil
		}(),
	)
	remoteTplParams := action.NewTemplateParams(
		action.WithTaskDirs(remoteEnv),
		action.WithParams(params),
		action.WithStore(e.ctx.Task().Store()),
		action.WithStores(e.ctx.Task().Stores()),
		action.WithTarget(e.ctx.Subject().GetId()),
		action.WithRuntime(e.ctx.Workspace().GetRuntime()),
	)

	// create helper with remote environment for rendering templates with remote env
	helper, err := action.NewActionHelper(e.ctx, remoteTplParams)
	if err != nil {
		return nil, nil, err
	}

	// update remote recipe dir
	helper.SetRecipeDirs(recipeDirs.CallerRecipeDir, recipeDirs.CalleeRecipeDir)

	return remoteEnv, helper, nil
}

func (e *RemoteExecAction) syncOutputFiles(
	ctx context.Context, session transport.RemoteSession, localEnv, remoteEnv action.Workspace) error {

	// Sync files
	downloader := NewDownloader(localEnv.GetFileSystem())
	downloader.AddDir(
		localEnv.GetOutputDirPath(),
		remoteEnv.GetOutputDirPath(),
		true)

	err := downloader.Download(ctx, session)
	if err != nil {
		return err
	}

	return nil
}

type localFiles struct {
	uploader *Uploader
	script   *action.Script
}

func (e *RemoteExecAction) GenerateLocalFiles(
	localEnv, remoteEnv action.Workspace, helper *action.ActionHelper) (*localFiles, error) {

	// Generate script with remote env as local file
	script, err := helper.NewScript(
		e.spec.ScriptTemplate, localEnv.GetScriptFilePath(), e.ctx.Subject().GetOs())
	if err != nil {
		return nil, err
	}
	localScriptPath := script.GetLocalPath()

	remoteScriptPath := remoteEnv.Join(
		remoteEnv.GetLogDirPath(),
		filepath.Base(localScriptPath))

	// Set remote execution path
	script.SetExecutePath(remoteScriptPath)

	uploader := NewUploader(localEnv.GetFileSystem(), remoteEnv)

	// Basic directories
	uploader.AddEmptyDir(
		remoteEnv.GetWorkDirPath(),
		remoteEnv.GetLogDirPath(),
		remoteEnv.GetInputDirPath(),
		remoteEnv.GetOutputDirPath())

	// Script
	var perm os.FileMode = 0755
	uploader.Add(localScriptPath, remoteScriptPath, &perm)

	// Context file and params file
	uploader.Add(action.GetStateFilePath(localEnv), action.GetStateFilePath(remoteEnv), nil)
	uploader.Add(action.GetParamsFilePath(localEnv), action.GetParamsFilePath(remoteEnv), nil)

	// Recipe assets(caller and callee)
	uploader.AddRecipeAssets(
		e.ctx.Task().Callee(),
		helper.GetCalleeRecipeDir())
	if e.ctx.Task().Caller() != nil {
		uploader.AddRecipeAssets(
			e.ctx.Task().Caller(),
			helper.GetCallerRecipeDir())
	}

	// Output directory sync(recursive upload)
	uploader.AddDir(
		localEnv.GetOutputDirPath(),
		remoteEnv.GetOutputDirPath(),
		true)

	return &localFiles{
		uploader: uploader,
		script:   script,
	}, nil
}

type RecipeDirs struct {
	CalleeRecipeDir string
	CallerRecipeDir string
}

func (e *RemoteExecAction) CreateRemoteWorkspace(
	ctx context.Context,
	session transport.RemoteSession) (action.Workspace, *RecipeDirs, error) {

	home, err := session.GetWorkRoot(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Path separator for the remote host. Node.Os is empty/"unknown" for
	// hosts that haven't been probed yet, which defaults to "/" (the SSH
	// case up to now).
	sep := remoteSeparator(e.ctx.Subject().GetOs())

	// Remote environment
	// $HOME/.niwashi/...
	remoteRootDir := workspace.NewPath(
		workspace.SimplePath(home),
		workspace.SimplePath(".niwashi"),
	).ToString(workspace.NewSeparatorResolver(sep))

	// Clone and replace root dir/separator in remote environment
	remoteEnv := e.ctx.Workspace().Clone()
	remoteEnv.ResetRootDir(remoteRootDir)
	remoteEnv.ResetSeparator(sep)

	// $HOME/.niwashi/recipe/<recipe_fqid>/...
	recipeRoot := remoteEnv.Join(remoteRootDir, "recipe")

	callee := e.ctx.Task().Callee()
	caller := e.ctx.Task().Caller()

	recipeDirs := &RecipeDirs{
		CalleeRecipeDir: remoteEnv.Join(
			recipeRoot,
			file.SanitizeFilename(callee.GetFqid())),
		CallerRecipeDir: func() string {
			if caller != nil {
				return remoteEnv.Join(
					recipeRoot,
					file.SanitizeFilename(caller.GetFqid()))
			}
			return ""
		}(),
	}

	return remoteEnv, recipeDirs, nil
}

// remoteSeparator returns the path separator used on a remote node with the
// given OS, matching the same "windows" comparison DefaultShell (script.go)
// uses to pick a shell.
func remoteSeparator(os string) string {
	if os == "windows" {
		return "\\"
	}
	return "/"
}
