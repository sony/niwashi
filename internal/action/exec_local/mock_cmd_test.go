// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_local

import (
	"context"
	"io"

	"github.com/sony/niwashi/internal/action"
)

// mockCmdFactory is a mock implementation of CmdFactory for testing.
type mockCmdFactory struct {
	History []string
	LastCmd *mockCmd
}

func (m *mockCmdFactory) NewCmd(ctx context.Context, name string, arg ...string) action.Cmd {
	m.History = append(m.History, "NewCmd: "+name+" "+joinArgs(arg))
	m.LastCmd = &mockCmd{
		parent: m,
		name:   name,
		args:   arg,
	}
	return m.LastCmd
}

type mockCmd struct {
	parent *mockCmdFactory
	name   string
	args   []string
	dir    string
	env    []string
	stdout io.Writer
	stderr io.Writer
}

func (m *mockCmd) Dir(dir string) action.Cmd {
	m.dir = dir
	return m
}

func (m *mockCmd) Env(env []string) action.Cmd {
	m.env = env
	return m
}

func (m *mockCmd) Stdout(w io.Writer) action.Cmd {
	m.stdout = w
	return m
}

func (m *mockCmd) Stderr(w io.Writer) action.Cmd {
	m.stderr = w
	return m
}

func (m *mockCmd) Run() error {
	m.parent.History = append(m.parent.History, "Run: "+m.name+" "+joinArgs(m.args))
	return nil
}

func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}

func NewMockCmdFactory() *mockCmdFactory {
	return &mockCmdFactory{
		History: []string{},
	}
}
