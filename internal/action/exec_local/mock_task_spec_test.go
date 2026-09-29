// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package exec_local

import "github.com/sony/niwashi/internal/action"

type taskSpec struct {
	*recipeSpec
	name       string
	actionType string
	actionSpec action.ActionSpec
}

func newDefaultTaskSpec() *taskSpec {
	return &taskSpec{
		recipeSpec: &recipeSpec{
			fqid: "test/test.test.test@1.0.0",
			assets: []string{
				"testdata/asset.txt",
				"testdata/hoge",
				"xxxx.sh",
			},
		},
		name:       "test dummy",
		actionType: "exec.local",
		actionSpec: &Spec{
			ScriptTemplate: "",
			InheritEnv: []string{
				"HOME",
			},
			EnvTemplate: map[string]string{
				"TEST_ENV": "test_value",
			},
		},
	}
}

func (s *taskSpec) GetRecipe() action.RecipeSpec {
	return s.recipeSpec
}

func (s *taskSpec) GetName() string {
	return s.name
}

func (t *DefaultTaskContext) Store() map[string]any {
	return make(map[string]any)
}

func (s *taskSpec) GetActionType() string {
	return s.actionType
}

func (s *taskSpec) GetActionSpec() action.ActionSpec {
	return s.actionSpec
}

type callerTaskSpec struct {
	*taskSpec
	argvTpl    []string
	workDirTpl string
	inheritEnv []string
}

func newDefaultCallerTaskSpec() *callerTaskSpec {
	return &callerTaskSpec{
		taskSpec: newDefaultTaskSpec(),
		argvTpl: []string{
			"hello",
			"world",
		},
		workDirTpl: "./testdata/workspace",
		inheritEnv: []string{
			"HOME",
			"USER",
		},
	}
}

func (s *callerTaskSpec) GetArgvTpl() []string {
	return s.argvTpl
}

func (s *callerTaskSpec) GetWorkDirTpl() string {
	return s.workDirTpl
}

func (s *callerTaskSpec) GetInheritEnv() []string {
	return s.inheritEnv
}
