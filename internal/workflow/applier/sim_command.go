// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package applier

import (
	"context"
	"fmt"
	"io"

	"github.com/sony/niwashi/internal/action"
	"gopkg.in/yaml.v3"
)

type SimCmdFactory struct{}

func (f *SimCmdFactory) NewCmd(ctx context.Context, name string, arg ...string) action.Cmd {
	return newSimulateCmd(ctx, name, arg...)
}

func NewSimCmdFactory() *SimCmdFactory {
	return &SimCmdFactory{}
}

type SimulateCmd struct {
	Name         string   `yaml:"cmd"`
	Arg          []string `yaml:"args"`
	WorkDir      string   `yaml:"workdir"`
	Environments []string `yaml:"env"`
	stdout       io.Writer
	stderr       io.Writer
}

func newSimulateCmd(ctx context.Context, name string, arg ...string) action.Cmd {
	return &SimulateCmd{
		Name: name,
		Arg:  arg,
	}
}

func (d *SimulateCmd) Dir(dir string) action.Cmd {
	d.WorkDir = dir
	return d
}

func (d *SimulateCmd) Env(env []string) action.Cmd {
	d.Environments = append(d.Environments, env...)
	return d
}

func (d *SimulateCmd) Stdout(stdout io.Writer) action.Cmd {
	d.stdout = stdout
	return d
}

func (d *SimulateCmd) Stderr(stderr io.Writer) action.Cmd {
	d.stderr = stderr
	return d
}

func (d *SimulateCmd) Run() error {

	v, err := yaml.Marshal(d)
	if err != nil {
		fmt.Println("Error marshaling SimulateCmd:", err)
	}
	fmt.Println("> Simulate action executed")
	fmt.Println(string(v))

	return nil
}
