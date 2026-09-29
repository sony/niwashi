// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sony/niwashi/internal/action"
)

type execLocalAction struct {
	local *action.ActionHelper
	ctx   action.ActionContext
	spec  *Spec
}

func BuildExecLocalAction(ctx action.ActionContext) (action.Action, error) {

	callee := ctx.Task().Callee()
	spec := callee.GetActionSpec().(*Spec)
	if spec == nil {
		return nil, fmt.Errorf("invalid spec for exec.local action")
	}

	helper, err := action.NewActionHelper(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &execLocalAction{
		local: helper,
		ctx:   ctx,
		spec:  spec,
	}, nil
}

func (e *execLocalAction) Execute(ctx context.Context) error {
	ws := e.ctx.Workspace()

	workdir, err := e.local.ResolveWorkDir(e.ctx.Task())
	if err != nil {
		return err
	}

	targetOS := e.ctx.Subject().GetOs()

	script, err := e.local.NewScript(
		e.spec.ScriptTemplate,
		ws.GetScriptFilePath(),
		targetOS,
	)
	if err != nil {
		return err
	}

	args, err := e.local.ResolveArguments()
	if err != nil {
		return err
	}

	cmdArgs := script.GetArgs(targetOS) // interpreter + args from shebang + script Path
	cmdArgs = append(cmdArgs, args...)  // + args from spec
	cmd := e.ctx.RunContext().NewCmd(ctx, cmdArgs[0], cmdArgs[1:]...)

	envvars, err := e.local.ResolveEnvironments(ws)
	if err != nil {
		return err
	}

	env := []string{}
	for k, v := range envvars {
		env = append(env, k+"="+v)
	}

	// inherit env from current process. exec.local and tool.run
	env = append(inheritSafeEnvironments(
		e.spec.InheritEnv,
		func() []string {
			if e.ctx.Task().Caller() != nil {
				return e.ctx.Task().Caller().GetInheritEnv()
			}
			return nil
		}(),
	), env...)

	cmd.Dir(workdir)
	cmd.Env(env)

	output := ws.GetOutput()
	cmd.Stdout(output.Stdout())
	cmd.Stderr(output.Stderr())

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return err
	}
	return nil
}

func isAllowedEnvVar(key string, allowedEnvVars []string) bool {
	for _, allowed := range allowedEnvVars {
		if strings.Contains(allowed, "*") {
			if matched, _ := filepath.Match(allowed, key); matched {
				return true
			}
		} else {
			if key == allowed {
				return true
			}
		}
	}
	return false
}

func inheritSafeEnvironments(inheritEnv ...[]string) []string {
	allowedEnvVars := []string{}

	for _, env := range inheritEnv {
		allowedEnvVars = append(allowedEnvVars, env...)
	}

	allowedEnvVars = append([]string{"PATH"}, allowedEnvVars...)

	env := []string{}
	for _, key := range os.Environ() {
		if isAllowedEnvVar(strings.SplitN(key, "=", 2)[0], allowedEnvVars) {
			env = append(env, key)
		}
	}
	return env
}
