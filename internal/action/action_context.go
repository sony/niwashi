// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package action

import (
	"io"

	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/transport"
	"github.com/sony/niwashi/internal/types"
	"github.com/sony/niwashi/internal/workspace"
)

type ActionContext interface {
	RunContext() RunContext

	Task() TaskContext
	Subject() Subject

	Workspace() Workspace

	TemplateParams() *TemplateParams
}

type RunContext interface {
	DryRun() string
	DryRunEnabled() bool

	CmdFactory
}

type TaskContext interface {
	Params() map[string]any
	Store() map[string]any
	Stores() map[string]types.Dict

	Callee() TaskSpec
	Caller() CallerTaskSpec
}

func TaskName(ctx TaskContext) string {
	if ctx.Caller() != nil {
		return ctx.Caller().GetName()
	}
	return ctx.Callee().GetName()
}

type Subject interface {
	GetScope() string
	GetId() string
	GetTransport() (transport.Transport, error)
	GetOperator(t transport.Transport) (transport.Operator, error)

	GetOs() string
}

type TaskDirs interface {
	GetEphemeralWorkDirPath() string
	GetRootDirPath() string
	GetInputDirPath() string
	GetOutputDirPath() string
	GetLogDirPath() string
	GetWorkDirPath() string
}

type TaskFiles interface {
	GetScriptFilePath() string
	GetStateFileName() string
	GetParamsFileName() string
}

type Workspace interface {
	workspace.PathResolver
	TaskDirs
	TaskFiles

	Clone() Workspace

	ResetRootDir(rootDir string)
	ResetSeparator(separator string)

	GetOutput() Output
	GetFileSystem() file.FileSystem
	GetRuntime() Runtime
}

type Output interface {
	Stdout() io.Writer
	Stderr() io.Writer

	io.Closer
}

type Runtime interface {
	GetToolAlias() map[string]string
	GetCapabilityBinding() map[string]string
	GetTools() types.Dict
	GetService() types.Dict
}

func GetParamsFilePath(ws Workspace) string {
	return ws.Join(
		ws.GetInputDirPath(),
		file.SanitizeFilename(ws.GetParamsFileName()))
}

func GetStateFilePath(ws Workspace) string {
	return ws.Join(
		ws.GetInputDirPath(),
		file.SanitizeFilename(ws.GetStateFileName()))
}
