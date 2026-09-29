// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package recipe

import (
	"errors"
	"fmt"

	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/types"
)

const ToolRunActionType = "tool.run"

func init() {
	action.RegisterBuilder(ToolRunActionType, &ToolRunBuilder{})
}

type ToolRunBuilder struct{}

func (b *ToolRunBuilder) NewSpec() action.ActionSpec {
	return &ToolRunSpec{}
}

func (b *ToolRunBuilder) Build(ctx action.ActionContext) (action.Action, error) {
	return nil, nil
}

type ToolRunSpec struct {
	ToolRef         string            `json:"toolRef" yaml:"toolRef"`
	Command         string            `json:"command" yaml:"command"`
	ArgvTemplate    []string          `json:"argvTpl" yaml:"argvTpl"`
	WorkdirTemplate string            `json:"workdirTpl" yaml:"workdirTpl"`
	EnvTemplate     map[string]string `json:"envTpl" yaml:"envTpl"`
	InheritEnv      []string          `json:"inheritEnv" yaml:"inheritEnv"`
}

func (s *ToolRunSpec) GetEnvTpl() map[string]string {
	return s.EnvTemplate
}

func (s *ToolRunSpec) Validate() error {
	var err error

	if s.ToolRef == "" {
		err = errors.Join(err, fmt.Errorf("toolRef is required"))
	}
	if s.Command == "" {
		err = errors.Join(err, fmt.Errorf("command is required"))
	}

	for k := range s.EnvTemplate {
		if !types.EnvVarNamePattern.MatchString(k) {
			err = errors.Join(err, fmt.Errorf("invalid env name: %s", k))
		}
	}

	return err
}
