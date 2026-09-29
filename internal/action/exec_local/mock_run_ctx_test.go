// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_local

import "github.com/sony/niwashi/internal/action"

type runContext struct {
	action.CmdFactory
	dryRun string
}

func NewDefaultRunContext() *runContext {
	return NewRunContext(
		WithDryRunMode("off"),
		WithCmdFactory(action.NewCmdFactory()),
	)
}

func NewRunContext(opts ...func(*runContext)) *runContext {
	rc := &runContext{}
	for _, opt := range opts {
		opt(rc)
	}
	return rc
}

func WithDryRunMode(mode string) func(*runContext) {
	return func(rc *runContext) {
		rc.dryRun = mode
	}
}

func WithCmdFactory(f action.CmdFactory) func(*runContext) {
	return func(rc *runContext) {
		rc.CmdFactory = f
	}
}

func (drc *runContext) DryRun() string {
	if drc == nil {
		return ""
	}
	return drc.dryRun
}

func (drc *runContext) DryRunEnabled() bool {
	if drc == nil {
		return false
	}
	return drc.dryRun != "" && drc.dryRun != "off"
}
