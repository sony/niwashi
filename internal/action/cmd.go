// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package action

import (
	"context"
	"io"
	"os/exec"
)

type CmdFactory interface {
	NewCmd(ctx context.Context, name string, arg ...string) Cmd
}

type Cmd interface {
	Dir(string) Cmd
	Env([]string) Cmd
	Stdout(io.Writer) Cmd
	Stderr(io.Writer) Cmd

	Run() error
}

type cmdFactory struct{}

func NewCmdFactory() CmdFactory {
	return &cmdFactory{}
}

func (*cmdFactory) NewCmd(ctx context.Context, name string, arg ...string) Cmd {
	return NewExecCmd(ctx, name, arg...)
}

type ExecCmd struct {
	cmd *exec.Cmd
}

func NewExecCmd(ctx context.Context, name string, arg ...string) Cmd {
	return &ExecCmd{
		// #nosec G204 -- command is derived from recipe authored by a trusted user
		cmd: exec.CommandContext(ctx, name, arg...),
	}
}

func (d *ExecCmd) Dir(dir string) Cmd {
	d.cmd.Dir = dir
	return d
}

func (d *ExecCmd) Env(env []string) Cmd {
	d.cmd.Env = append(d.cmd.Env, env...)
	return d
}

func (d *ExecCmd) Stdout(stdout io.Writer) Cmd {
	d.cmd.Stdout = stdout
	return d
}

func (d *ExecCmd) Stderr(stderr io.Writer) Cmd {
	d.cmd.Stderr = stderr
	return d
}

func (d *ExecCmd) Run() error {
	return d.cmd.Run()
}
