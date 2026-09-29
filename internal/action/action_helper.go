// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package action

import (
	"errors"
	"fmt"
	"os"
)

type ActionHelper struct {
	ActionContext
	templateEngine  TemplateEngine
	callerRecipeDir string
	calleeRecipeDir string
}

func NewActionHelper(ctx ActionContext, tp *TemplateParams) (*ActionHelper, error) {
	ah := &ActionHelper{
		ActionContext: ctx,
	}

	if tp == nil {
		tp = ctx.TemplateParams()
	}

	ah.templateEngine = NewDefaultTemplateEngine(tp)

	return ah, nil
}

func (ah *ActionHelper) SetRecipeDirs(callerDir, calleeDir string) {
	ah.callerRecipeDir = callerDir
	ah.calleeRecipeDir = calleeDir
}

func (ah *ActionHelper) GetCallerRecipeDir() string {
	if ah.callerRecipeDir != "" {
		return ah.callerRecipeDir
	}

	if ah.Task().Caller() != nil {
		return ah.Task().Caller().GetDir()
	}
	return ""
}

func (ah *ActionHelper) GetCalleeRecipeDir() string {
	if ah.calleeRecipeDir != "" {
		return ah.calleeRecipeDir
	}

	if ah.Task().Callee() != nil {
		return ah.Task().Callee().GetDir()
	}
	return ""
}

func (ah *ActionHelper) RenderTemplate(input string) (string, error) {
	if ah.templateEngine == nil {
		return "", fmt.Errorf("template engine is not initialized")
	}

	return ah.templateEngine.Render(input)
}

func (ah *ActionHelper) NewScript(scriptTpl, path, targetOS string) (*Script, error) {

	// replace {{ .Assets }} with callee's recipe dir
	finalScriptTpl := replaceAssetsDir(scriptTpl, ah.GetCalleeRecipeDir())

	script, err := NewScript(
		ah.Task(),
		ah.Workspace().GetFileSystem(),
		finalScriptTpl,
		path,
		ah.templateEngine,
		targetOS)
	if err != nil {
		return nil, err
	}

	return script, nil
}

func (ah *ActionHelper) WriteFile(path string, data []byte, perm os.FileMode) (err error) {
	w, err := ah.Workspace().GetFileSystem().OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, w.Close())
	}()

	_, err = w.Write(data)
	return err
}

func (ah *ActionHelper) ResolveArguments() ([]string, error) {
	caller := ah.Task().Caller()
	if caller == nil {
		return []string{}, nil
	}

	recipeDir := ah.GetCallerRecipeDir()

	args := []string{}
	for _, v := range caller.GetArgvTpl() {
		// replace {{ .Assets }} with caller's recipe dir
		// and render template

		arg, err := ah.templateEngine.Render(replaceAssetsDir(v, recipeDir))
		if err != nil {
			return nil, fmt.Errorf("failed to render argv template in task %q: %w", TaskName(ah.Task()), err)
		}
		args = append(args, arg)
	}
	return args, nil
}

func (ah *ActionHelper) getFinalEnvironment(task TaskContext) map[string]string {

	env := map[string]string{}

	updateEnv := func(spec ActionSpec, dir string) {
		for k, v := range spec.GetEnvTpl() {
			// replace {{ .Assets }} with callee's recipe dir
			env[k] = replaceAssetsDir(v, dir)
		}
	}

	callee := task.Callee()
	if callee != nil {

		// 1. callee recipe's default env
		for k, v := range callee.GetDefaultEnv() {
			env[k] = v
		}

		// 2. callee's envTpl
		updateEnv(callee.GetActionSpec(), ah.GetCalleeRecipeDir())
	}

	caller := task.Caller()
	if caller != nil {
		// 3. caller's envTpl
		updateEnv(caller.GetActionSpec(), ah.GetCallerRecipeDir())
	}

	return env
}

func (ah *ActionHelper) ResolveEnvironments(ws Workspace) (map[string]string, error) {

	envs := map[string]string{}

	task := ah.Task()

	// Resolve variables with template engine
	for k, v := range ah.getFinalEnvironment(task) {
		ev, err := ah.templateEngine.Render(v)
		if err != nil {
			return nil, fmt.Errorf("failed to render env template for key %q in task %q: %w", k, TaskName(task), err)
		}
		envs[k] = ev
	}

	// system variables

	tbl := map[string]string{
		// System Global Environments
		"NWS_DRY_RUN": ah.RunContext().DryRun(),
		"NWS_DRY_RUN_ENABLED": func() string {
			if ah.RunContext().DryRunEnabled() {
				return "1"
			}
			return "0"
		}(),
		"NWS_LOG_LEVEL": os.Getenv("NWS_LOG_LEVEL"),

		// Recipe/Action specific environments
		"NWS_SCOPE":      ah.Subject().GetScope(), // running phase
		"NWS_TARGET_ID":  ah.Subject().GetId(),    // target id(ex. node id/cluster id)
		"NWS_WORK_DIR":   "{{ .Paths.work_dir }}", // workdir for this action
		"NWS_LOG_DIR":    "{{ .Paths.log_dir }}",
		"NWS_INPUT_DIR":  "{{ .Paths.input_dir }}",
		"NWS_OUTPUT_DIR": "{{ .Paths.output_dir }}",
		"NWS_PARAMS":     GetParamsFilePath(ws),
		"NWS_STATE":      GetStateFilePath(ws),
		"NWS_CONTEXT":    GetStateFilePath(ws), // todo: deprecate
	}

	for k, v := range tbl {
		ev, err := ah.templateEngine.Render(v)
		if err != nil {
			return nil, fmt.Errorf("failed to render env template for key %q in task %q: %w", k, TaskName(task), err)
		}
		envs[k] = ev
	}
	return envs, nil
}

func (ah *ActionHelper) ResolveWorkDir(task TaskContext) (string, error) {
	var err error

	var workdir string
	if task.Caller() != nil {
		workdir = task.Caller().GetWorkDirTpl()
	}

	// make workdir from template
	if workdir == "" {
		workdir = "{{ .Paths.work_dir }}"
	}

	workdir, err = ah.templateEngine.Render(workdir)
	if err != nil {
		return "", fmt.Errorf("failed to render workdir template in task %q: %w", TaskName(ah.Task()), err)
	}

	return workdir, nil
}
