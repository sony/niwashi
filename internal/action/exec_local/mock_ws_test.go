// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_local

import (
	"io"
	"os"

	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/workspace"
)

type Workspace struct {
	*workspace.TaskWorkspace
	output                        action.Output
	fileSystem                    file.FileSystem
	runtime                       action.Runtime
	stateFileName, paramsFileName string
}

func NewWorkspace(p *workspace.TaskWorkspace, opts ...func(*Workspace)) *Workspace {
	w := &Workspace{
		TaskWorkspace: p,
		output:        NewDefaultOutput(),
		fileSystem:    file.NewDefaultFileSystem(),
	}
	w.Set(opts...)
	return w
}

func (w *Workspace) GetOutput() action.Output {
	return w.output
}

func (w *Workspace) GetFileSystem() file.FileSystem {
	return w.fileSystem
}

func (w *Workspace) GetRuntime() action.Runtime {
	return w.runtime
}

func (w *Workspace) GetTaskWorkspace() *workspace.TaskWorkspace {
	return w.TaskWorkspace
}

func (w *Workspace) GetStateFileName() string {
	return w.stateFileName
}

func (w *Workspace) GetParamsFileName() string {
	return w.paramsFileName
}

func (w *Workspace) Clone() action.Workspace {
	return &Workspace{
		TaskWorkspace:  w.TaskWorkspace.Clone(),
		output:         w.output,
		fileSystem:     w.fileSystem,
		runtime:        w.runtime,
		stateFileName:  w.stateFileName,
		paramsFileName: w.paramsFileName,
	}
}

func (w *Workspace) Set(opts ...func(*Workspace)) {
	for _, opt := range opts {
		opt(w)
	}
}

func WithPathResolver(p workspace.TaskWorkspace) func(*Workspace) {
	return func(w *Workspace) {
		w.TaskWorkspace = &p
	}
}

func WithOutput(o action.Output) func(*Workspace) {
	return func(w *Workspace) {
		w.output = o
	}
}

func WithFileSystem(f file.FileSystem) func(*Workspace) {
	return func(w *Workspace) {
		w.fileSystem = f
	}
}

func WithRuntime(r action.Runtime) func(*Workspace) {
	return func(w *Workspace) {
		w.runtime = r
	}
}

func WithStateFileName(name string) func(*Workspace) {
	return func(w *Workspace) {
		w.stateFileName = name
	}
}

func WithParamsFileName(name string) func(*Workspace) {
	return func(w *Workspace) {
		w.paramsFileName = name
	}
}

type DefaultOutput struct {
	StdoutWriter io.Writer
	StderrWriter io.Writer
}

func NewDefaultOutput() *DefaultOutput {
	return &DefaultOutput{
		StdoutWriter: os.Stdout,
		StderrWriter: os.Stderr,
	}
}

func (o *DefaultOutput) Close() error {
	// No resources to close for DefaultOutput
	return nil
}

func (o *DefaultOutput) Stdout() io.Writer {
	return o.StdoutWriter
}

func (o *DefaultOutput) Stderr() io.Writer {
	return o.StderrWriter
}
